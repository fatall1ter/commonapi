package infra

import (
	"errors"
	"fmt"
	"net/http"

	"git.countmax.ru/countmax/commonapi/domain"
	"github.com/labstack/echo/v4"
)

var (
	errVCConfigNotFound error = errors.New("videocheck configs not found")
	errEmptyPID         error = errors.New("empty pid not allowed")
	errEmptyPrjID       error = errors.New("empty projectId not allowed")
)

// AssetsResponse http wrapper with metadata
type VideoCheckResponse struct {
	Data domain.VideocheckConfigs `json:"data"`
	Metadata
}

// apiCustomerVCConfigs godoc
// @Summary Get all videocheck configurations
// @Description get slice of customer videocheck configuration with offset, limit parameters
// @Produce  json
// @Tags cm_info
// @Param offset query integer false "default=0"
// @Param limit query integer false "default=10"
// @Success 200 {object} infra.VideoCheckResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 405 {object} infra.HTTPError
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/videochecks/configs [get]
func (s *Server) apiCustomerVCConfigs(c echo.Context) error {
	offset, limit := s.getPageParams(c)
	vcs, count, err := s.infoRepo.FindVideoChechCfgs(offset, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	if vcs == nil || len(vcs) == 0 {
		return c.JSON(http.StatusNotFound, ErrNotFound(errVCConfigNotFound))
	}
	response := VideoCheckResponse{
		Data: vcs,
		Metadata: Metadata{
			ResultSet: ResultSet{
				Count:  int64(len(vcs)),
				Offset: offset,
				Limit:  limit,
				Total:  count,
			},
		},
	}
	return c.JSON(http.StatusOK, response)
}

// apiGetCustomerVCConfigByID godoc
// @Summary Get customer's videocheck configuration
// @Description get customer videocheck configuration by customerID parameter
// @Produce  json
// @Tags cm_info
// @Param pid path integer true "Code 1S"
// @Success 200 {object} domain.CustomerConfig
// @Failure 400 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/videochecks/configs/{pid} [get]
func (s *Server) apiGetCustomerVCConfigByID(c echo.Context) error {
	pid := c.Param("pid")
	if pid == "" {
		s.log.Errorf("bad request apiGetCustomerVCConfigByID, %v", errEmptyPID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyPID))
	}
	vc, err := s.infoRepo.FindVideoCheckCfgByID(atoi64(pid))
	if err != nil {
		s.log.Errorf("apiGetCustomerVCConfigByID, error %v", err)
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	if vc == nil {
		s.log.Warnf("apiGetCustomerVCConfigByID for id=%s not found", pid)
		return c.JSON(http.StatusNotFound, ErrNotFound(errVCConfigNotFound))
	}
	return c.JSON(http.StatusOK, vc)
}

// apiNewCustomerVCConfigByID godoc
// @Summary Inserts new customer's videocheck configuration
// @Description inserts customer videocheck configuration
// @Security ApiKeyAuth
// @Tags cm_info
// @Accept json
// @Produce json
// @Param videocheckConfig body domain.VideocheckConfig true "New videocheck configuration"
// @Success 201 {object} infra.SuccessResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/videochecks/configs [post]
func (s *Server) apiNewCustomerVCConfigByID(c echo.Context) error {
	vchkcfg := domain.VideocheckConfig{}
	if err := c.Bind(&vchkcfg); err != nil {
		s.log.Errorf("apiNewCustomerVCConfigByID, bad request error %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, ErrInvalidRequest(err))
	}
	if vchkcfg.ProjectID == 0 {
		s.log.Errorf("c.Bind incorrect parse payload, empty ProjectID, payload=%v", vchkcfg)
		return echo.NewHTTPError(http.StatusBadRequest, ErrInvalidRequest(errEmptyPrjID))
	}
	err := s.infoRepo.StoreVideoCheckCfg(vchkcfg)
	if err != nil {
		s.log.Errorf("apiNewCustomerVCConfigByID, error %v", err)
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	c.Response().Header().Set("Location", fmt.Sprintf("/v2/videochecks/configs/%d", vchkcfg.ProjectID))
	return c.JSON(http.StatusCreated, CreatedStatus(fmt.Sprintf("%d", vchkcfg.ProjectID)))
}

// apiUpdCustomerVCConfigByID godoc
// @Summary Update customer's videocheck configuration
// @Description updates customer videocheck configuration
// @Security ApiKeyAuth
// @Tags cm_info
// @Accept json
// @Produce json
// @Param pid path integer true "Code 1S"
// @Param videocheckConfig body domain.VideocheckConfig true "New videocheck configuration"
// @Success 200 {object} infra.SuccessResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/videochecks/configs/{pid} [put]
func (s *Server) apiUpdCustomerVCConfigByID(c echo.Context) error {
	vchkcfg := domain.VideocheckConfig{}
	if err := c.Bind(&vchkcfg); err != nil {
		s.log.Errorf("apiUpdCustomerVCConfigByID, bad request error %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, ErrInvalidRequest(err))
	}
	pid := c.Param("pid")
	if pid == "" || vchkcfg.ProjectID == 0 {
		s.log.Errorf("bad request apiUpdCustomerVCConfigByID, %v", errEmptyPID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyPID))
	}
	vchkcfg.ProjectID = atoi64(pid)
	err := s.infoRepo.UpSertVideoCheckCfg(vchkcfg)
	if err != nil {
		s.log.Errorf("apiUpdCustomerVCConfigByID, error %v", err)
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	c.Response().Header().Set("Location", fmt.Sprintf("/v2/videochecks/configs/%d", vchkcfg.ProjectID))
	return c.JSON(http.StatusOK, OkStatus("updated"))
}

// apiDelCustomerVCConfigByID godoc
// @Summary Delete customer's videocheck configuration
// @Description delete customer videocheck configuration
// @Security ApiKeyAuth
// @Tags cm_info
// @Accept json
// @Produce json
// @Param pid path integer true "Code 1S"
// @Success 200 {object} infra.SuccessResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/videochecks/configs/{pid} [delete]
func (s *Server) apiDelCustomerVCConfigByID(c echo.Context) error {
	pid := c.Param("pid")
	if pid == "" {
		s.log.Errorf("bad request apiDelCustomerVCConfigByID, %v", errEmptyPID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyPID))
	}
	rowCount, err := s.infoRepo.DeleteVideoCheckCfgByID(atoi64(pid))
	if err != nil {
		s.log.Errorf("apiDelCustomerVCConfigByID, error %v", err)
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	if rowCount == 0 {
		s.log.Warnf("apiDelCustomerVCConfigByID for id=%s not found", pid)
		return c.JSON(http.StatusNotFound, ErrNotFound(errVCConfigNotFound))
	}
	return c.JSON(http.StatusOK, OkStatus(fmt.Sprintf("delete %d records", rowCount)))
}
