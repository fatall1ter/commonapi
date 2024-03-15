package infra

import (
	"errors"
	"net/http"
	"strconv"

	"git.countmax.ru/countmax/commonapi/domain"
	"github.com/labstack/echo/v4"
)

var (
	errAssetNotFound error = errors.New("assets not found")
	errEmptyID       error = errors.New("empty id not allowed")
)

// AssetsResponse http wrapper with metadata
type AssetsResponse struct {
	Data domain.Assets `json:"data"`
	Metadata
}

// apiAssets godoc
// @Summary Get all assets
// @Description get slice of Assets with offset, limit parameters
// @Produce  json
// @Security ApiKeyAuth
// @Tags intraservice
// @Param offset query integer false "default=0"
// @Param limit query integer false "default=10"
// @Success 200 {object} infra.AssetsResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/assets [get]
func (s *Server) apiAssets(c echo.Context) error {
	offset, limit := s.getPageParams(c)
	assets, count, err := s.assetRepo.FindAll(offset, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	if assets == nil || len(assets) == 0 {
		return c.JSON(http.StatusNotFound, ErrNotFound(errAssetNotFound))
	}
	response := AssetsResponse{
		Data: assets,
		Metadata: Metadata{
			ResultSet: ResultSet{
				Count:  int64(len(assets)),
				Offset: offset,
				Limit:  limit,
				Total:  count,
			},
		},
	}
	return c.JSON(http.StatusOK, response)
}

// apiAssetByID godoc
// @Summary Get asset by id
// @Description Get specified Asset by id
// @Produce  json
// @Security ApiKeyAuth
// @Tags intraservice
// @Param id path integer true "AssetID"
// @Success 200 {object} domain.Asset
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/assets/{id} [get]
func (s *Server) apiAssetByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		s.log.Errorf("bad request apiAssetByID, %v", errEmptyID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyID))
	}
	asset, err := s.assetRepo.FindByID(atoi64(id))
	if err != nil {
		s.log.Errorf("bad request apiAssetByID, error %v", err)
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	if asset == nil {
		s.log.Warnf("apiAssetByID for id=%s not found", id)
		return c.JSON(http.StatusNotFound, ErrNotFound(errAssetNotFound))
	}
	return c.JSON(http.StatusOK, asset)
}

func atoi64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		i = 0
	}
	return i
}
