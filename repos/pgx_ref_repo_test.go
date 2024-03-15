package repos

//"time"
//"git.countmax.ru/countmax/commonapi/domain"
// "go.uber.org/zap"
// "go.uber.org/zap/zapcore"

const (
	testRefPGXCS string = "postgres://commonapi:commonapi@elk-01:15432/evolution?sslmode=disable&pool_max_conns=2"
)

// func TestPGHealth(t *testing.T) {
// 	tr, err := NewRefPGRepo(testRefPGXCS, timeoutRef)
// 	if err != nil {
// 		t.Errorf("Expected success make NewRefPGRepo, but error, %v", err)
// 	}
// 	err = tr.Health()
// 	if err != nil {
// 		t.Errorf("Expected success tr.Health(), but error %v", err)
// 	}
// }

// func TestPGFindEntities(t *testing.T) {
// 	tr, err := NewRefPGRepo(testRefPGXCS, timeoutRef)
// 	if err != nil {
// 		t.Errorf("Expected success make NewRefPGRepo, but error, %v", err)
// 	}
// 	entities, count, err := tr.FindEntities(1, 1)
// 	if err != nil {
// 		t.Errorf("Expected success tr.FindEntities(), but error %v", err)
// 	}
// 	if entities == nil {
// 		t.Error("Expected entities with values, but not")
// 	}
// 	if len(entities) == 0 {
// 		t.Error("Expected entities with values, but not")
// 	}
// 	if len(entities) != 1 {
// 		t.Errorf("Expected count of entities with %d values, but got with %d", 1, count)
// 	}
// }

// func TestPGFindEntitiesNoRows(t *testing.T) {
// 	tr, err := NewRefPGRepo(testRefPGXCS, timeoutRef)
// 	if err != nil {
// 		t.Errorf("Expected success make NewRefPGRepo, but error, %v", err)
// 	}
// 	entities, _, err := tr.FindEntities(10000, 99)
// 	if err != nil {
// 		t.Errorf("Expected success tr.FindEntities(), but error %v", err)
// 	}
// 	if entities == nil {
// 		t.Error("Expected empty not nil entities, but not")
// 	}
// 	if len(entities) != 0 {
// 		t.Error("Expected entities with zero values, but not")
// 	}
// }

// func TestPGFindEntityByID(t *testing.T) {
// 	tr, err := NewRefPGRepo(testRefPGXCS, timeoutRef)
// 	if err != nil {
// 		t.Errorf("Expected success make NewRefPGRepo, but error, %v", err)
// 	}
// 	id := "chain"
// 	entity, err := tr.FindEntityByID(id)
// 	if err != nil {
// 		t.Errorf("Expected success tr.FindEntityByID(), but error %v", err)
// 	}
// 	if entity == nil {
// 		t.Error("Expected entities with values, but not")
// 	}
// 	if entity.ID != id {
// 		t.Errorf("Expected entity with id=%s, but got %v", id, entity)
// 	}
// }

// func BenchmarkPGHealth(b *testing.B) {
// 	tr, err := NewRefPGRepo(testRefPGXCS, timeoutRef)
// 	if err != nil {
// 		b.Errorf("Expected success make NewRefPGRepo, but error, %v", err)
// 	}
// 	for i := 0; i < b.N; i++ {
// 		tr.Health()
// 	}
// }
