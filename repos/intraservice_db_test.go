package repos

import (
	"time"
	//"git.countmax.ru/countmax/commonapi/domain"
	// "go.uber.org/zap"
	// "go.uber.org/zap/zapcore"
)

const (
	testCS                = "server=sd-test.watcom.ru;user id=commonapi;password=commonapi;port=1433;database=Intraservice;"
	timeout time.Duration = 5 * time.Second
)

// var (
// 	log *zap.SugaredLogger = makeLogger()
// )

// func TestHealth(t *testing.T) {
// 	tr, err := NewAssetSQLRepo(testCS, timeout)
// 	if err != nil {
// 		t.Errorf("Expected success make AssetSQLRepo, but error, %v", err)
// 	}
// 	err = tr.Health()
// 	if err != nil {
// 		t.Errorf("Expected success tr.Health(), but error %v", err)
// 	}
// }

// func BenchmarkHealth(b *testing.B) {
// 	tr, _ := NewAssetSQLRepo(testCS, timeout)
// 	for i := 0; i < b.N; i++ {
// 		tr.Health()
// 	}
// }
