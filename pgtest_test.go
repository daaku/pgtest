package pgtest

import (
	"database/sql"
	"testing"

	"github.com/facebookgo/ensure"
	_ "github.com/lib/pq"
)

func TestRun(t *testing.T) {
	pg := Start()
	defer pg.Stop()

	db, err := sql.Open("postgres", pg.URL)
	ensure.Nil(t, err)
	var n int
	err = db.QueryRow("SELECT 1").Scan(&n)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, n, 1)
}
