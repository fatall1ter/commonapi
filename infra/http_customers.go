package infra

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"git.countmax.ru/countmax/commonapi/domain"
	"github.com/labstack/echo/v4"
)

var (
	errCustomerConfigNotFound error          = errors.New("customer configs not found")
	errProjectNotFound        error          = errors.New("project not found")
	errProjectFTPNotFound     error          = errors.New("project' ftp settings not found")
	errEmptyCID               error          = errors.New("empty cid not allowed")
	errDB                     error          = errors.New("internal db error")
	reDBType                  *regexp.Regexp = regexp.MustCompile(`(?m)[1,2,3,4,10]`)
	errWrongDBType            error          = errors.New("dbType wrong format, allow only 1,2,3,4,10")
)

// AssetsResponse http wrapper with metadata
type CustomerResponse struct {
	Data domain.CustomerConfigs `json:"data"`
	Metadata
}

// apiCustomerConfigs godoc
// @Summary Get all customer configurations
// @Description get slice of customer configuration with offset, limit, active parameters
// @Produce  json
// @Security ApiKeyAuth
// @Tags cm_info
// @Param offset query integer false "default=0"
// @Param limit query integer false "default=10"
// @Param active query string false "default=true"
// @Success 200 {object} infra.CustomerResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/customers/configs [get]
func (s *Server) apiCustomerConfigs(c echo.Context) error {
	offset, limit := s.getPageParams(c)
	enabled, err := strconv.ParseBool(c.QueryParam("active"))
	if err != nil {
		enabled = true
	}
	customers, count, err := s.infoRepo.FindCustomersConfig(offset, limit, enabled)
	if err != nil {
		s.log.Errorf("infoRepo.FindCustomersConfig, error %v", err)
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(errDB))
	}
	if customers == nil || len(customers) == 0 {
		return c.JSON(http.StatusNotFound, ErrNotFound(errCustomerConfigNotFound))
	}
	response := CustomerResponse{
		Data: customers,
		Metadata: Metadata{
			ResultSet: ResultSet{
				Count:  int64(len(customers)),
				Offset: offset,
				Limit:  limit,
				Total:  count,
			},
		},
	}
	return c.JSON(http.StatusOK, response)
}

// apiCustomerConfigByID godoc
// @Summary Get customer config by id
// @Description Get specified Customer by id
// @Produce  json
// @Security ApiKeyAuth
// @Tags cm_info
// @Param id path integer true "Code 1S"
// @Success 200 {object} domain.CustomerConfig
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/customers/{id}/configs [get]
func (s *Server) apiCustomerConfigByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		s.log.Errorf("bad request apiCustomerConfigByID, %v", errEmptyID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyID))
	}
	customer, err := s.infoRepo.FindCustomerConfig(atoi64(id))
	if err != nil {
		s.log.Errorf("apiCustomerConfigByID, error %v", err)
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(errDB))
	}
	if customer == nil {
		s.log.Warnf("apiCustomerConfigByID for id=%s not found", id)
		return c.JSON(http.StatusNotFound, ErrNotFound(errCustomerConfigNotFound))
	}
	return c.JSON(http.StatusOK, customer)
}

// ProjectsResponse http wrapper with metadata
type ProjectsResponse struct {
	Data domain.Projects `json:"data"`
	Metadata
}

// nolint:lll
// apiProjects godoc
// @Summary Get all projects configurations
// @Description get slice of customer/project configuration with DataBase connection and offset, limit, active parameters
// @Produce  json
// @Security ApiKeyAuth
// @Tags cm_info
// @Param offset query integer false "default=0"
// @Param limit query integer false "default=10"
// @Param active query string false "default=true"
// @Success 200 {object} infra.ProjectsResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/projects [get]
func (s *Server) apiProjects(c echo.Context) error {
	offset, limit := s.getPageParams(c)
	enabled, err := strconv.ParseBool(c.QueryParam("active"))
	if err != nil {
		enabled = true
	}
	projects, count, err := s.infoRepo.FindProjects(offset, limit, enabled)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	response := ProjectsResponse{
		Data: projects,
		Metadata: Metadata{
			ResultSet: ResultSet{
				Count:  int64(len(projects)),
				Offset: offset,
				Limit:  limit,
				Total:  count,
			},
		},
	}
	return c.JSON(http.StatusOK, response)
}

// apiProjectByID godoc
// @Summary Get customer's project config by id
// @Description Get specified customer/project with database connection,
// @Description id can be colon separated format <customerID>:<DBTypeID>
// @Description example: 1234:1 or 1234:2, <DBTypeID> allowed 1,2,3,4,10
// @Produce  json
// @Security ApiKeyAuth
// @Tags cm_info
// @Param id path string true "Code 1S[:DBTypeID]"
// @Success 200 {object} domain.Project
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/projects/{id} [get]
func (s *Server) apiProjectByID(c echo.Context) error {
	rawID := c.Param("id")
	if rawID == "" {
		s.log.Errorf("bad request apiProjectByID, %s", errEmptyID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyID))
	}
	id, dbType := colonSeparate(rawID)
	ptr := &dbType
	if dbType != "" && !reDBType.MatchString(dbType) {
		s.log.Warnf("dbType=%s wrong format, allow only 1,2,3,4,10", dbType)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errWrongDBType))
	}
	if dbType == "" {
		ptr = nil
	}
	s.log.Debugf("id=%s, dbType=%s", id, dbType)
	project, err := s.infoRepo.FindProjectByID(atoi64(id), ptr)
	if err != nil {
		s.log.Errorf("apiProjectByID, error %v", err)
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	if project == nil {
		return c.JSON(http.StatusNotFound, ErrServerInternal(errProjectNotFound))
	}
	return c.JSON(http.StatusOK, project)
}

func colonSeparate(raw string) (string, string) {
	sep := "%3A"
	raw = strings.ToUpper(raw)
	if strings.Contains(raw, ":") {
		sep = ":"
	}
	parts := strings.Split(raw, sep)
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

// apiProjectFTPByID godoc
// @Summary Get customer's ftp settings by id
// @Description Get specified customer/project ftp settings
// @Produce  json
// @Security ApiKeyAuth
// @Tags cm_info
// @Param id path integer true "Code 1S"
// @Success 200 {object} domain.FTPinfo
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/projects/{id}/ftpinfo [get]
func (s *Server) apiProjectFTPByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		s.log.Errorf("bad request apiProjectFTPByID, %v", errEmptyID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyID))
	}
	ftp, err := s.infoRepo.FindFTPByID(atoi64(id))
	if err != nil {
		s.log.Errorf("apiProjectFTPByID, error %v", err)
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	if ftp == nil {
		s.log.Warnf("apiProjectFTPByID for id=%s not found", id)
		return c.JSON(http.StatusNotFound, ErrNotFound(errProjectFTPNotFound))
	}
	return c.JSON(http.StatusOK, ftp)
}

// ProjectMCResponse http wrapper with metadata
type ProjectMCResponse struct {
	Data domain.ManualCountings `json:"data"`
	Metadata
}

// apiProjectMC godoc
// @Summary Get all manual countings of controller
// @Description get slice of manual countings of specified controller in the  project of customer with parameters id,
// @Description cid and offset, limit parameters
// @Produce  json
// @Security ApiKeyAuth
// @Tags cm_info
// @Param offset query integer false "default=0"
// @Param limit query integer false "default=10"
// @Param id path integer true "Code 1S of customer"
// @Param cid path integer true "ControllerID"
// @Success 200 {object} infra.ProjectMCResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/projects/{id}/controllers/{cid}/manualcnts [get]
func (s *Server) apiProjectMC(c echo.Context) error {
	offset, limit := s.getPageParams(c)
	id := c.Param("id")
	if id == "" {
		s.log.Errorf("bad request apiProjectMC, %v", errEmptyID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyID))
	}
	cid := c.Param("cid")
	if id == "" {
		s.log.Errorf("bad request apiProjectMC, %v", errEmptyID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyCID))
	}
	mcs, count, err := s.infoRepo.FindManualCountings(atoi64(id), atoi64(cid), offset, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	if mcs == nil || len(mcs) == 0 {
		return c.JSON(http.StatusNotFound, ErrNotFound(errProjectNotFound))
	}
	response := ProjectMCResponse{
		Data: mcs,
		Metadata: Metadata{
			ResultSet: ResultSet{
				Count:  int64(len(mcs)),
				Offset: offset,
				Limit:  limit,
				Total:  count,
			},
		},
	}
	return c.JSON(http.StatusOK, response)
}
