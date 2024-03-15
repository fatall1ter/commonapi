package repos

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"git.countmax.ru/countmax/commonapi/domain"

	_ "github.com/denisenkom/go-mssqldb" // go-mssqldb mssql import
)

const (
	repoType string = "postgres://"
)

// NewRefRepo fabric for ReferenceRepo
func NewRefRepo(connString string, timeout time.Duration) (domain.RefRepo, error) {
	if strings.Contains(connString, repoType) {
		return NewRefPGRepo(connString, timeout)
	}
	return NewRefSQLRepo(connString, timeout)
}

type RefSQLRepo struct {
	connString string
	timeout    time.Duration
	db         *sql.DB
}

// NewRefSQLRepo returns instance of SQLRepo with connected evolution DB,
// format connection string: "server=study-app;user id=transport;password=transport;port=1433;database=CM_Transport523;"
func NewRefSQLRepo(connString string, timeout time.Duration) (*RefSQLRepo, error) {
	rer := &RefSQLRepo{
		connString: connString,
		timeout:    timeout,
	}
	err := rer.connDB()
	if err != nil {
		return nil, err
	}
	return rer, nil
}

// connDB - make connection to database
func (rer *RefSQLRepo) connDB() error {
	db, err := sql.Open("mssql", rer.connString)
	if err != nil {
		return nil
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return err
	}
	db.SetMaxOpenConns(dbMaxOpenConnections)
	db.SetMaxIdleConns(dbMaxIdleConnections)
	rer.db = db
	return nil
}

// Close - close connection to database
func (rer *RefSQLRepo) Close() error {
	return rer.db.Close()
}

// Health - check connection to DB
func (rer *RefSQLRepo) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), rer.timeout)
	defer cancel()
	row := rer.db.QueryRowContext(ctx, "select @@version as [v]")
	var tmp string
	return row.Scan(&tmp)
}

// FindEntities returns slice of Entity from evolution database
func (rer *RefSQLRepo) FindEntities(offset, limit int64) (domain.Entities, int64, error) {
	query := `SELECT  [id]
				,[description]
			FROM [dbo].[entities]
			ORDER BY [id]
			OFFSET ? ROWS 
			FETCH NEXT ? ROWS ONLY;
			SELECT count(*) FROM [dbo].[entities]`
	ctx, cancel := context.WithTimeout(context.Background(), rer.timeout)
	defer cancel()
	rows, err := rer.db.QueryContext(ctx, query, offset, limit)
	if err != nil {
		return nil, -1, fmt.Errorf("query [%s], error %v", query, err)
	}
	defer rows.Close()
	result := make(domain.Entities, 0, limit)
	for rows.Next() {
		item := domain.Entity{}
		err = rows.Scan(&item.ID, &item.Description)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, -1, nil
			}
			return nil, -1, fmt.Errorf("query [%s], error %v", query, err)
		}
		result = append(result, item)
	}
	var count int64
	if rows.NextResultSet() {
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				if err == sql.ErrNoRows {
					return result, -1, nil
				}
				return nil, -1, fmt.Errorf("query [%s], error %v", query, err)
			}

		}
	}
	return result, count, nil
}

// FindEntityByID returns single entity with specified id from evolution database
func (rer *RefSQLRepo) FindEntityByID(id string) (*domain.Entity, error) {
	query := `SELECT  [id]
				,[description]
			FROM [dbo].[entities] where id=?`
	ctx, cancel := context.WithTimeout(context.Background(), rer.timeout)
	defer cancel()
	item := domain.Entity{}
	err := rer.db.QueryRowContext(ctx, query, id).Scan(&item.ID, &item.Description)
	if err != nil {
		return nil, fmt.Errorf("query [%s], error %v", query, err)
	}
	return &item, nil
}
