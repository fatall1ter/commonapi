/*Package infra contain lowlevel implementation all other abstraction layers

Author: Karpov Artem, mailto: karpov@watcom.ru
Date: 2020-01-24
*/
package infra

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"git.countmax.ru/countmax/commonapi/domain"
	"git.countmax.ru/countmax/commonapi/repos"

	// nolint:gosec
	_ "net/http/pprof" // for remote profiling

	_ "git.countmax.ru/countmax/commonapi/docs" // docs is generated by Swag CLI, you have to import it.

	consulapi "github.com/hashicorp/consul/api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

const (
	requestIDName                   = "request_id"
	periodHealthCheck time.Duration = 30 * time.Second
	repoINFO          string        = "cminfo"
	repoISDB          string        = "intraservice_db"
	repoISAPI         string        = "intraservice_api"
	repoRef           string        = "reference"
	repoToken         string        = "keycloak"
)

// Server - api main web server
type Server struct {
	log       *zap.SugaredLogger
	consul    *consulapi.Agent
	mux       *echo.Echo
	version   string
	githash   string
	build     string
	config    *viper.Viper
	mService  *prometheus.GaugeVec
	mAPI      *prometheus.SummaryVec
	mCache    *prometheus.GaugeVec
	assetRepo domain.AssetRepo
	infoRepo  domain.CustomerRepo
	sdRepo    domain.SDRepo
	refRepo   domain.RefRepo
	chCancel  <-chan struct{}
	fnCancel  context.CancelFunc
	token     string
	mu        *sync.Mutex
	repoReady map[string]bool
}

// NewServer builder main document server
func NewServer(ctx context.Context, version, build, githash string) *Server {
	ctxembed, cancel := context.WithCancel(ctx)
	s := &Server{
		version:   version,
		githash:   githash,
		build:     build,
		mService:  common,
		mAPI:      httpDuration,
		mCache:    redis,
		chCancel:  ctxembed.Done(),
		fnCancel:  cancel,
		mu:        &sync.Mutex{},
		repoReady: make(map[string]bool),
	}
	// fill repo not ready
	s.setRepoState(repoINFO, false)
	s.setRepoState(repoISDB, false)
	s.setRepoState(repoISAPI, false)
	s.setRepoState(repoRef, false)
	s.setRepoState(repoToken, false)
	// fill config
	s.setConfig()

	s.setLogger(version, build, githash)
	// fill token
	s.token = s.config.GetString("httpd.hardcode_token")
	// init asset repo
	timeout := s.config.GetDuration("intraservice.sqltimeout_sec") * time.Second
	assetRepo, err := repos.NewAssetSQLRepo(s.config.GetString("intraservice.dsn"), timeout)
	if err != nil {
		s.log.Errorf("connection to Intraservice DB error, %v", err)
		s.mService.WithLabelValues("intraserviceDB",
			getSrvPortDB(s.config.GetString("intraservice.dsn")), s.version, s.githash, s.build).Set(0)
		go s.assetReconnector(timeout, s.chCancel)
	} else {
		s.setRepoState(repoISDB, true)
		s.mService.WithLabelValues("intraserviceDB",
			getSrvPortDB(s.config.GetString("intraservice.dsn")), s.version, s.githash, s.build).Set(1)
		s.assetRepo = assetRepo
	}

	timeoutInfo := s.config.GetDuration("cminfo.sqltimeout_sec") * time.Second
	infoRepo, err := repos.NewCustomersRepo(s.config.GetString("cminfo.dsn"), timeoutInfo)
	if err != nil {
		s.log.Errorf("connection to CM_INFO DB error, %v", err)
		s.mService.WithLabelValues("CM_INFO",
			getSrvPortDB(s.config.GetString("cminfo.dsn")), s.version, s.githash, s.build).Set(0)
		go s.infoReconnector(timeoutInfo, s.chCancel)
	} else {
		s.setRepoState(repoINFO, true)
		s.mService.WithLabelValues("CM_INFO",
			getSrvPortDB(s.config.GetString("cminfo.dsn")), s.version, s.githash, s.build).Set(1)
		s.infoRepo = infoRepo
	}

	timeoutRef := s.config.GetDuration("ref.sqltimeout_sec") * time.Second
	refRepo, err := repos.NewRefRepo(s.config.GetString("ref.dsn"), timeoutRef)
	if err != nil {
		s.log.Errorf("connection to evolution DB error, %v", err)
		s.mService.WithLabelValues("evolution",
			getSrvPortDB(s.config.GetString("ref.dsn")), s.version, s.githash, s.build).Set(0)
		go s.refReconnector(timeoutRef, s.chCancel)
	} else {
		s.setRepoState(repoRef, true)
		s.mService.WithLabelValues("evolution",
			getSrvPortDB(s.config.GetString("ref.dsn")), s.version, s.githash, s.build).Set(1)
		s.refRepo = refRepo
	}

	timeoutSD := s.config.GetDuration("intraservice.httptimeout_sec") * time.Second
	sdRepo := repos.NewISAPI(s.config.GetString("intraservice.url"),
		s.config.GetString("intraservice.user"), s.config.GetString("intraservice.pass"), timeoutSD, s.log)
	s.setRepoState(repoISAPI, true)
	s.sdRepo = sdRepo

	// start healthChecker
	go s.healthChecker(periodHealthCheck, s.chCancel)
	return s
}

func (s *Server) isReadyRepo(name string) bool {
	s.mu.Lock()
	ready, ok := s.repoReady[name]
	s.mu.Unlock()
	if !ok {
		ready = false
	}
	return ready
}

func (s *Server) setRepoState(name string, state bool) {
	s.mu.Lock()
	s.repoReady[name] = state
	s.mu.Unlock()
}

// assetReconnector...
func (s *Server) assetReconnector(period time.Duration, cancel <-chan struct{}) {
	s.log.Debugf("starting assetReconnector")
	defer s.log.Debugf("stopped assetReconnector")
	tick := time.NewTicker(period)
	for {
		select {
		case <-cancel:
			tick.Stop()
			return
		case <-tick.C:
			assetRepo, err := repos.NewAssetSQLRepo(s.config.GetString("intraservice.dsn"), period)
			if err != nil {
				s.log.Errorf("connection to Intraservice (%s) DB error, %v, wait %v and repeat",
					shadowConnString(s.config.GetString("intraservice.dsn")), err, period)
				continue
			}
			s.assetRepo = assetRepo
			tick.Stop()
			return
		}
	}
}

// infoReconnector...
func (s *Server) infoReconnector(period time.Duration, cancel <-chan struct{}) {
	s.log.Debugf("starting infoReconnector")
	defer s.log.Debugf("stopped infoReconnector")
	tick := time.NewTicker(period)
	for {
		select {
		case <-cancel:
			tick.Stop()
			return
		case <-tick.C:
			infoRepo, err := repos.NewCustomersRepo(s.config.GetString("cminfo.dsn"), period)
			if err != nil {
				s.log.Errorf("connection to CM_INFO (%s) DB error, %v, wait %v and repeat",
					shadowConnString(s.config.GetString("cminfo.dsn")), err, period)
				continue
			}
			s.infoRepo = infoRepo
			tick.Stop()
			return
		}
	}
}

// refReconnector...
func (s *Server) refReconnector(period time.Duration, cancel <-chan struct{}) {
	s.log.Debugf("starting refReconnector")
	defer s.log.Debugf("stopped refReconnector")
	tick := time.NewTicker(period)
	for {
		select {
		case <-cancel:
			tick.Stop()
			return
		case <-tick.C:
			refRepo, err := repos.NewRefRepo(s.config.GetString("ref.dsn"), period)
			if err != nil {
				s.log.Errorf("connection to evolution (%s) DB error, %v, wait %v and repeat",
					shadowConnString(s.config.GetString("ref.dsn")), err, period)
				continue
			}
			s.refRepo = refRepo
			tick.Stop()
			return
		}
	}
}

// Run is running Server
func (s *Server) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	s.fnCancel = cancel
	s.chCancel = ctx.Done()
	s.registerRoutes()
	port := s.config.GetString("httpd.port")
	host := s.config.GetString("httpd.host") + ":" + port
	s.log.Infof("http server starting on the [%s] tcp port", host)
	go func() {
		if err := s.mux.Start(host); err != http.ErrServerClosed {
			s.log.Fatalf("http server error: %v", err)
		}
	}()
	s.consulRegister()
}

// Stop is stopping Server
func (s *Server) Stop() {
	s.log.Infof("got signal to stopping server")
	stopDuration := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), stopDuration)
	defer cancel()
	defer s.consulDeRegister()
	s.fnCancel()
	if err := s.mux.Shutdown(ctx); err != nil {
		s.log.Fatal(err)
	}
}

func (s *Server) registerRoutes() {
	e := echo.New()
	e.HidePort = true
	e.HideBanner = true // hide banner ECHO
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(s.customHTTPLogger)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowOrigins:     s.config.GetStringSlice("httpd.allow_origins"),
	}))
	e.GET("/", s.redirectToSwag)
	// metric handler
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.GET("/health", s.apiHealthCheck)
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	// group for report implementation

	// secure area
	auth := e.Group("/v2")
	auth.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == s.token, nil
	}))
	auth.GET("/assets/:id", s.apiAssetByID)
	auth.GET("/assets", s.apiAssets)
	auth.GET("/customers/:id/configs", s.apiCustomerConfigByID)
	auth.GET("/customers/configs", s.apiCustomerConfigs)

	auth.GET("/projects/:id", s.apiProjectByID)
	auth.GET("/projects/:id/ftpinfo", s.apiProjectFTPByID)
	auth.GET("/projects/:id/controllers/:cid/manualcnts", s.apiProjectMC)
	auth.GET("/projects", s.apiProjects)
	e.GET("/v2/videochecks/configs/:pid", s.apiGetCustomerVCConfigByID)
	auth.PUT("/videochecks/configs/:pid", s.apiUpdCustomerVCConfigByID)
	auth.DELETE("/videochecks/configs/:pid", s.apiDelCustomerVCConfigByID)
	e.GET("/v2/videochecks/configs", s.apiCustomerVCConfigs)
	auth.POST("/videochecks/configs", s.apiNewCustomerVCConfigByID)

	// Intraservice API
	auth.GET("/tasks/controllers/:sn", s.apiTasksByControllerSN)
	auth.PUT("/tasks/:id/comment", s.apiTaskAddComment)
	auth.PUT("/tasks/:id/status", s.apiTaskChangeStatus)

	// evolution
	e.GET("/v2/entities", s.apiEntities)
	e.GET("/v2/entities/:id", s.apiEntityByID)

	// keycloak token validation
	// auth.GET("/v2/layouts", s.apiLayouts)
	// auth.GET("/v2/devices", s.apiDevices)

	// pprof
	dbg := e.Group("/debug")
	dbg.GET("/pprof/*", echo.WrapHandler(http.DefaultServeMux))
	s.mux = e
}

func (s *Server) healthChecker(period time.Duration, cancel <-chan struct{}) {
	s.log.Infof("starting healthChecker with scheduller %v", period)
	defer s.log.Info("stopped healthChecker")
	tick := time.NewTicker(period)
	for {
		select {
		case <-cancel:
			tick.Stop()
			return
		case <-tick.C:
			s.log.Debug("time to execute healthChecker")
			err := s.healthCheck()
			if err != nil {
				s.log.Errorf("healthCheck error %s", err)
			}
		}
	}
}

func (s *Server) healthCheck() error {
	s.log.Infof("starting healthCheck")
	defer s.log.Infof("stopped healthCheck")
	var resErr error
	genHealth := true

	// intraserviceDB
	if s.isReadyRepo(repoISDB) {
		err := s.assetRepo.Health()
		if err != nil {
			genHealth = false
			resErr = err
			s.log.With(zap.String("repository",
				getSrvPortDB(s.config.GetString("intraservice.dsn")))).Errorf("healthCheck error, %v", err)
			s.mService.WithLabelValues("intraserviceDB",
				getSrvPortDB(s.config.GetString("intraservice.dsn")), s.version, s.githash, s.build).Set(0)
		} else {

			s.mService.WithLabelValues("intraserviceDB",
				getSrvPortDB(s.config.GetString("intraservice.dsn")), s.version, s.githash, s.build).Set(1)
		}
	} else {
		genHealth = false
	}

	// CM_INFO
	if s.isReadyRepo(repoINFO) {
		err := s.infoRepo.Health()
		if err != nil {
			genHealth = false
			resErr = err
			s.log.With(zap.String("repository",
				getSrvPortDB(s.config.GetString("cminfo.dsn")))).Errorf("healthCheck error, %v", err)
			s.mService.WithLabelValues("CM_INFO",
				getSrvPortDB(s.config.GetString("cminfo.dsn")), s.version, s.githash, s.build).Set(0)
		} else {
			s.mService.WithLabelValues("CM_INFO",
				getSrvPortDB(s.config.GetString("cminfo.dsn")), s.version, s.githash, s.build).Set(1)
		}
	} else {
		genHealth = false
	}

	// intraserviceDB
	if s.isReadyRepo(repoRef) {
		err := s.refRepo.Health()
		if err != nil {
			genHealth = false
			resErr = err
			s.log.With(zap.String("repository",
				getSrvPortDB(s.config.GetString("ref.dsn")))).Errorf("healthCheck error, %v", err)
			s.mService.WithLabelValues("evolution",
				getSrvPortDB(s.config.GetString("ref.dsn")), s.version, s.githash, s.build).Set(0)
		} else {

			s.mService.WithLabelValues("evolution",
				getSrvPortDB(s.config.GetString("ref.dsn")), s.version, s.githash, s.build).Set(1)
		}
	} else {
		genHealth = false
	}

	// intraserviceAPI
	if s.isReadyRepo(repoISAPI) {
		err := s.sdRepo.Health()
		if err != nil {
			genHealth = false
			resErr = err
			s.log.With(zap.String("repository",
				s.config.GetString("intraservice.url"))).Errorf("healthCheck error, %v", err)
			s.mService.WithLabelValues("IntraserviceAPI",
				s.config.GetString("intraservice.url"),
				s.version, s.githash, s.build).Set(0)
		} else {
			s.mService.WithLabelValues("IntraserviceAPI",
				s.config.GetString("intraservice.url"), s.version, s.githash, s.build).Set(1)
		}
	} else {
		genHealth = false
	}

	// general state
	if genHealth {
		s.mService.WithLabelValues("general", "localhost", s.version, s.githash, s.build).Set(1)
	} else {
		s.mService.WithLabelValues("general", "localhost", s.version, s.githash, s.build).Set(0)
	}
	return resErr
}

// shadowConnString returns connString with userId and password replaced to ***
func shadowConnString(cstring string) string {
	// "server=sql2-caravan.watcom.local;user id=commonapi;password=commonapi;port=1433;database=evolution;"
	// "postgres://commonapi:commonapi@elk-01:15432/evolution?sslmode=disable&pool_max_conns=2"

	var parts []string
	if strings.Contains(cstring, "postgres://") {
		parts = strings.Split(cstring, ":")
	} else {
		parts = strings.Split(cstring, ";")
	}

	resSlice := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.Contains(part, "user id") {
			part = "user id=******"
		}
		if strings.Contains(part, "password") {
			part = "password=******"
		}
		if strings.Contains(part, "//") {
			part = "//******"
		}
		if strings.Contains(part, "@") {
			subparts := strings.Split(part, "@")
			if len(subparts) >= 1 {
				part = "******@" + subparts[1]
			} else {
				part = "******@"
			}
		}
		resSlice = append(resSlice, part)
	}
	return fmt.Sprintf("%v", resSlice)
}

// nolint:gocyclo
// getSrvPortDB returns uniq string with server.port.dbname parts
func getSrvPortDB(connStr string) string {
	// "server=some.domain.ip;user id=root;password=master;port=1433;database=CM_Net523"
	// "postgres://commonapi:commonapi@elk-01:15432/evolution?sslmode=disable&pool_max_conns=2"
	// sqlserver://root:master@study-app.watcom.local:1433?database=CM_Karpov523&connection_timeout=0&encrypt=disable
	const (
		csURL string = "url"
		csDSN string = "dsn"
	)
	kind := csURL
	u, err := url.Parse(connStr)
	if err != nil || u.Scheme == "" {
		kind = csDSN
	}
	var srv, port, db string
	switch kind {
	case csURL:
		// postgres
		if u.Scheme == "postgres" {
			db = strings.ReplaceAll(u.Path, "/", "")
		}
		if u.Scheme == "sqlserver" {
			params := strings.Split(u.RawQuery, "&")
			for _, par := range params {
				if strings.Contains(par, "database=") {
					dbs := strings.Split(par, "=")
					if len(dbs) == 2 {
						db = dbs[1]
					}
				}
			}
		}
		srv = u.Hostname()
		port = u.Port()
	case csDSN:
		parts := strings.Split(connStr, ";")
		for _, part := range parts {
			if strings.Contains(part, "database=") {
				dbs := strings.Split(part, "=")
				if len(dbs) == 2 {
					db = dbs[1]
				}
			}
			if strings.Contains(part, "server=") {
				dbs := strings.Split(part, "=")
				if len(dbs) == 2 {
					srv = dbs[1]
				}
			}
			if strings.Contains(part, "port=") {
				dbs := strings.Split(part, "=")
				if len(dbs) == 2 {
					port = dbs[1]
				}
			}
		}
	}
	return fmt.Sprintf("[%s].[%s].[%s]", srv, port, db)
}
