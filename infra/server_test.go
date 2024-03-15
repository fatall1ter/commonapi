package infra

import (
	"testing"
)

const (
	sqlCS          string = "server=sql2-caravan.watcom.local;user id=commonapi;password=commonapi;port=1433;database=evolution;"
	sqlShadow      string = "[server=sql2-caravan.watcom.local user id=****** password=****** port=1433 database=evolution ]"
	postgresURL    string = "postgres://commonapi:commonapi@elk-01.watcom.local:15432/evolution?sslmode=disable&pool_max_conns=2"
	postgresShadow string = "[postgres //****** ******@elk-01.watcom.local 15432/evolution?sslmode=disable&pool_max_conns=2]"
)

type tcase struct {
	title    string
	connStr  string
	expected string
	iseq     bool
}

func TestShadowConnString(t *testing.T) {
	var tCases []tcase = []tcase{
		tcase{
			title:    "sql shadowed success",
			connStr:  sqlCS,
			expected: sqlShadow,
			iseq:     true,
		},
		tcase{
			title:    "url shadowed success",
			connStr:  postgresURL,
			expected: postgresShadow,
			iseq:     true,
		},
		tcase{
			title:    "sql shadowed wrong",
			connStr:  sqlCS,
			expected: postgresShadow,
			iseq:     false,
		},
	}
	// test cicle
	for _, tc := range tCases {
		t.Run(tc.title, func(t *testing.T) {
			shadow := shadowConnString(tc.connStr)
			if tc.iseq != (shadow == tc.expected) {
				t.Errorf("Expected shadow=%s, for %s, but got %s", tc.expected, tc.connStr, shadow)
			}
			//t.Logf("connStr=%s\nshadow=%s\n", tc.connStr, shadow)
		})
	}

}

func TestGetSrvPortDB(t *testing.T) {
	var tCases []tcase = []tcase{
		tcase{
			title:    "sql success",
			connStr:  sqlCS,
			expected: "[sql2-caravan.watcom.local].[1433].[evolution]",
			iseq:     true,
		},
		tcase{
			title:    "url success",
			connStr:  postgresURL,
			expected: "[elk-01.watcom.local].[15432].[evolution]",
			iseq:     true,
		},
		tcase{
			title:    "url wrong",
			connStr:  sqlCS,
			expected: "[elk-01.watcom.local].[15432].[evolution]",
			iseq:     false,
		},
	}
	// test cicle
	for _, tc := range tCases {
		t.Run(tc.title, func(t *testing.T) {
			srvPD := getSrvPortDB(tc.connStr)
			if tc.iseq != (srvPD == tc.expected) {
				t.Errorf("Expected srv.port.db=%s, for %s, but got %s", tc.expected, tc.connStr, srvPD)
			}
			//t.Logf("connStr=%s\nsrv.port.db=%s\n", tc.connStr, srvPD)
		})
	}

}
