package repos

import (
	"testing"
	"time"
	//"git.countmax.ru/countmax/commonapi/domain"
	// "go.uber.org/zap"
	// "go.uber.org/zap/zapcore"
)

const (
	testRefCS                = "server=sql2-caravan;user id=commonapi;password=commonapi;port=1433;database=evolution;"
	timeoutRef time.Duration = 5 * time.Second
)

func TestRefHealth(t *testing.T) {
	tr, err := NewRefSQLRepo(testRefCS, timeoutRef)
	if err != nil {
		t.Errorf("Expected success make RefSQLRepo, but error, %v", err)
	}
	err = tr.Health()
	if err != nil {
		t.Errorf("Expected success tr.Health(), but error %v", err)
	}
}

func TestFindEntities(t *testing.T) {
	tr, err := NewRefSQLRepo(testRefCS, timeoutRef)
	if err != nil {
		t.Errorf("Expected success make RefSQLRepo, but error, %v", err)
	}
	entities, count, err := tr.FindEntities(1, 1)
	if err != nil {
		t.Errorf("Expected success tr.FindEntities(), but error %v", err)
	}
	if entities == nil {
		t.Error("Expected entities with values, but not")
	}
	if len(entities) == 0 {
		t.Error("Expected entities with values, but not")
	}
	if len(entities) != 1 {
		t.Errorf("Expected count of entities with %d values, but got with %d", 1, count)
	}
}

func TestFindEntityByID(t *testing.T) {
	tr, err := NewRefSQLRepo(testRefCS, timeoutRef)
	if err != nil {
		t.Errorf("Expected success make RefSQLRepo, but error, %v", err)
	}
	id := "chain"
	entity, err := tr.FindEntityByID(id)
	if err != nil {
		t.Errorf("Expected success tr.FindEntityByID(), but error %v", err)
	}
	if entity == nil {
		t.Error("Expected entities with values, but not")
	}
	if entity.ID != id {
		t.Errorf("Expected entity with id=%s, but got %v", id, entity)
	}
}

// func BenchmarkRefHealth(b *testing.B) {
// 	tr, err := NewRefSQLRepo(testRefCS, timeoutRef)
// 	if err != nil {
// 		b.Errorf("Expected success make RefSQLRepo, but error, %v", err)
// 	}
// 	for i := 0; i < b.N; i++ {
// 		tr.Health()
// 	}
// }
