package repos

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	apiUser        string        = "tester"
	apiPass        string        = "watcom"
	apiURL         string        = "http://sd-test.watcom.ru"
	apiTimeout     time.Duration = 5 * time.Second
	taskID         string        = "524308"
	openStatus     int           = 31
	workStatus     int           = 27
	needInfoStatus int           = 35
	doneStatus     int           = 29
	closeStats     int           = 28
)

var (
	log *zap.SugaredLogger = makeLogger()
)

// func TestAPIFindAllBySN(t *testing.T) {
// 	api := NewISAPI(apiURL, apiUser, apiPass, apiTimeout, log)
// 	if api == nil {
// 		t.Error("Expected success api init, but not")
// 	}
// 	_, count, err := api.FindAllBySN("Мега", 0, 2)
// 	if err != nil {
// 		t.Errorf("Expected success FindAllBySN, but error %v", err)
// 	}
// 	t.Logf("FindAllBySN tasks count=%d", count)
// }

// func TestAPITaskAddComment(t *testing.T) {
// 	api := NewISAPI(apiURL, apiUser, apiPass, apiTimeout, log)
// 	if api == nil {
// 		t.Error("Expected success api init, but not")
// 	}
// 	err := api.TaskAddComment(taskID, "go test -v -> testing")
// 	if err != nil {
// 		t.Errorf("Expected success TaskAddComment, but error %v", err)
// 	}
// }

// func TestAPITaskSetStatus(t *testing.T) {
// 	api := NewISAPI(apiURL, apiUser, apiPass, apiTimeout, log)
// 	if api == nil {
// 		t.Error("Expected success api init, but not")
// 	}
// 	status0 := domain.TaskStatus{
// 		StatusID: workStatus,
// 		Comment:  "set work status",
// 	}
// 	err := api.TaskSetStatus(taskID, status0)
// 	if err != nil {
// 		t.Errorf("Expected success TaskSetStatus, but error %v", err)
// 	}
// 	status1 := domain.TaskStatus{
// 		StatusID: openStatus,
// 		Comment:  "repair open status",
// 	}
// 	err = api.TaskSetStatus(taskID, status1)
// 	if err != nil {
// 		t.Errorf("Expected success TaskSetStatus, but error %v", err)
// 	}
// }

// func TestAPIHealth(t *testing.T) {
// 	api := NewISAPI(apiURL, apiUser, apiPass, apiTimeout, log)
// 	if api == nil {
// 		t.Error("Expected success api init, but not")
// 	}
// 	err := api.Health()
// 	if err != nil {
// 		t.Errorf("Expected success Health check, but error %v", err)
// 	}
// }

//
func makeLogger() *zap.SugaredLogger {
	atom := zap.NewAtomicLevel()
	atom.SetLevel(zap.DebugLevel)
	// To keep the example deterministic, disable timestamps in the output.
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	),
		zap.AddCaller())
	return logger.Sugar()
}
