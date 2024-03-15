package repos

import (
	"testing"
	//"git.countmax.ru/countmax/commonapi/domain"
	// "go.uber.org/zap"
	// "go.uber.org/zap/zapcore"
)

const (
	testINFO = "sqlserver://root:master@study-app.watcom.local:1433?database=CM_Info_Test&connection_timeout=0&encrypt=disable"
	//timeout    time.Duration = 5 * time.Second
)

func TestCustomersRepoHealth(t *testing.T) {
	cr, err := NewCustomersRepo(testINFO, timeout)
	if err != nil {
		t.Errorf("Expected success make CustomersRepo, but error, %v", err)
	}
	err = cr.Health()
	if err != nil {
		t.Errorf("Expected success cr.Health(), but error %v", err)
	}
}

func TestFindCustomersConfig(t *testing.T) {
	cr, err := NewCustomersRepo(testINFO, timeout)
	if err != nil {
		t.Errorf("Expected success make CustomersRepo, but error, %v", err)
	}
	customers, count, err := cr.FindCustomersConfig(0, 10, true)
	if err != nil {
		t.Errorf("Expected success cr.FindCustomersConfig(), but error %v", err)
	}
	if len(customers) == 0 {
		t.Error("Expected some customers, but got 0")
	}
	t.Logf("count of customers=%d", count)
}
