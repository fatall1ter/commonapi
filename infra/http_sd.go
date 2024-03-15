package infra

import (
	"errors"
	"net/http"

	"git.countmax.ru/countmax/commonapi/domain"
	"github.com/labstack/echo/v4"
)

var (
	errTasksCntrlNotFound error = errors.New("tasks for controller not found")
	errEmptySN            error = errors.New("empty sn not allowed")
)

// SDTasksResponse http wrapper with metadata
type SDTasksResponse struct {
	Data domain.ControllerTasks `json:"data"`
	Metadata
}

// apiTasksByControllerSN godoc
// @Summary Get all tasks for controller's serial number
// @Description Get slice of tasks (json) for controller by serial number with limit parameters
// @Produce  json
// @Security ApiKeyAuth
// @Tags intraservice
// @Param limit query integer false "default=10"
// @Param sn path string true "SerialNumber of controller"
// @Success 200 {object} infra.SDTasksResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/tasks/controllers/{sn} [get]
func (s *Server) apiTasksByControllerSN(c echo.Context) error {
	offset, limit := s.getPageParams(c)
	sn := c.Param("sn")
	if sn == "" {
		s.log.Errorf("bad request apiTasksByControllerSN, %v", errEmptySN)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptySN))
	}
	tasks, count, err := s.sdRepo.FindAllBySN(sn, offset, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	if tasks == nil || len(tasks) == 0 {
		return c.JSON(http.StatusNotFound, ErrNotFound(errTasksCntrlNotFound))
	}
	response := SDTasksResponse{
		Data: tasks,
		Metadata: Metadata{
			ResultSet: ResultSet{
				Count:  int64(len(tasks)),
				Offset: 0,
				Limit:  limit,
				Total:  count,
			},
		},
	}
	return c.JSON(http.StatusOK, response)
}

// apiTaskAddComment godoc
// @Summary Add comment to task
// @Description Add comment to the Task in the ServiceDesk by TaskID
// @Produce  json
// @Security ApiKeyAuth
// @Tags intraservice
// @Param comment body domain.TaskComment true "comment content"
// @Param id path string true "TaskID"
// @Success 200 {object} infra.SuccessResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/tasks/{id}/comment [put]
func (s *Server) apiTaskAddComment(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		s.log.Errorf("bad request apiTaskAddComment, %v", errEmptyID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyID))
	}
	comment := &domain.TaskComment{}
	if err := c.Bind(comment); err != nil {
		s.log.Errorf("apiTaskAddComment, bad request error %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, ErrInvalidRequest(err))
	}
	err := s.sdRepo.TaskAddComment(id, comment.Comment)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	return c.JSON(http.StatusOK, OkStatus("success"))
}

// apiTaskChangeStatus godoc
// @Summary Change status task
// @Description Change status of the Task in the ServiceDesk by TaskID
// @Produce  json
// @Security ApiKeyAuth
// @Tags intraservice
// @Param status body domain.TaskStatus true "status data description"
// @Param id path string true "TaskID"
// @Success 200 {object} infra.SuccessResponse
// @Failure 400 {object} infra.HTTPError
// @Failure 401 {object} infra.HTTPError
// @Failure 405 {object} infra.HTTPError
// @Failure 404 {object} infra.ErrResponse
// @Failure 500 {object} infra.ErrResponse
// @Router /v2/tasks/{id}/status [put]
func (s *Server) apiTaskChangeStatus(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		s.log.Errorf("bad request apiTaskChangeStatus, %v", errEmptyID)
		return c.JSON(http.StatusBadRequest, ErrInvalidRequest(errEmptyID))
	}
	status := domain.TaskStatus{}
	if err := c.Bind(&status); err != nil {
		s.log.Errorf("apiTaskChangeStatus, bad request error %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, ErrInvalidRequest(err))
	}
	err := s.sdRepo.TaskSetStatus(id, status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrServerInternal(err))
	}
	return c.JSON(http.StatusOK, OkStatus("success"))
}
