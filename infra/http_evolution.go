package infra

import (
	"errors"
	"net/http"

	"git.countmax.ru/countmax/commonapi/domain"
	"github.com/labstack/echo/v4"
)

var (
	errEmptyEntityID error = errors.New("empty entity id not allowed")
)

// ReferenceResponse http wrapper with metadata
type ReferenceResponse struct {
	Data domain.Entities `json:"data"`
	Metadata
}

// apiEntities godoc
// @Summary Get all entities types
// @Description get slice of entities with offset, limit parameters
// @Produce  json
// @Tags reference
// @Param offset query integer false "default=0"
// @Param limit query integer false "default=10"
// @Success 200 {object} infra.ReferenceResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 405 {object} infra.HTTPError
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/entities [get]
func (s *Server) apiEntities(c echo.Context) error {
	offset, limit := s.getPageParams(c)
	entities, count, err := s.refRepo.FindEntities(offset, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	if entities == nil || len(entities) == 0 {
		return c.JSON(http.StatusNotFound, ErrNotFound(errVCConfigNotFound))
	}
	response := ReferenceResponse{
		Data: entities,
		Metadata: Metadata{
			ResultSet: ResultSet{
				Count:  int64(len(entities)),
				Offset: offset,
				Limit:  limit,
				Total:  count,
			},
		},
	}
	return c.JSON(http.StatusOK, response)
}

// apiEntityByID godoc
// @Summary Get entity by id
// @Description get entity by entity ID
// @Produce  json
// @Tags reference
// @Param id path string true "entity id"
// @Success 200 {object} domain.Entity
// @Failure 400 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/entities/{id} [get]
func (s *Server) apiEntityByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		s.log.Errorf("bad request apiEntityByID, %v", errEmptyEntityID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyEntityID))
	}
	e, err := s.refRepo.FindEntityByID(id)
	if err != nil {
		s.log.Errorf("apiEntityByID, error %v", err)
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	if e == nil {
		s.log.Warnf("apiEntityByID for id=%s not found", id)
		return c.JSON(http.StatusNotFound, ErrNotFound(errVCConfigNotFound))
	}
	return c.JSON(http.StatusOK, *e)
}
