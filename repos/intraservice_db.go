package repos

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"git.countmax.ru/countmax/commonapi/domain"

	_ "github.com/denisenkom/go-mssqldb" // go-mssqldb mssql import
)

const (
	// dbMaxOpenConnections -  maximum number of open connections to the database
	dbMaxOpenConnections = 2
	// dbMAXIdleConnections -  maximum number of connections in the idle connection pool
	dbMaxIdleConnections = 2
)

type AssetSQLRepo struct {
	connString string
	timeout    time.Duration
	db         *sql.DB
}

// NewAssetSQLRepo returns instance of SQLRepo with connected Intraservice DB,
// format connection string: "server=study-app;user id=transport;password=transport;port=1433;database=CM_Transport523;"
func NewAssetSQLRepo(connString string, timeout time.Duration) (*AssetSQLRepo, error) {
	asr := &AssetSQLRepo{
		connString: connString,
		timeout:    timeout,
	}
	err := asr.connDB()
	if err != nil {
		return nil, err
	}
	return asr, nil
}

// connDB - make connection to database
func (asr *AssetSQLRepo) connDB() error {
	db, err := sql.Open("mssql", asr.connString)
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
	asr.db = db
	return nil
}

// CloseCMAX - close connection to database
func (asr *AssetSQLRepo) Close() error {
	return asr.db.Close()
}

// Health - check connection to DB
func (asr *AssetSQLRepo) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), asr.timeout)
	defer cancel()
	row := asr.db.QueryRowContext(ctx, "select @@version as [v]")
	var tmp string
	return row.Scan(&tmp)
}

// FindAll returns slice of assets from Intraservice database
func (asr *AssetSQLRepo) FindAll(offset, limit int64) (domain.Assets, int64, error) {
	query := `SELECT Data.value('(/data/field[@id=62])[1]', 'int') as ID, 
					Name, 
					ParentId, 
					Changed, 
					Id as ServiceDeskID 
			FROM [dbo].[Asset]
			WHERE Data.value('(/data/field[@id=62])[1]', 'int') IS NOT NULL
			AND ParentId IS NOT NULL
			ORDER BY Id
			OFFSET ? ROWS 
			FETCH NEXT ? ROWS ONLY;
			SELECT count(*) 
			FROM [dbo].[Asset]
			WHERE Data.value('(/data/field[@id=62])[1]', 'int') IS NOT NULL
			AND ParentId IS NOT NULL`
	ctx, cancel := context.WithTimeout(context.Background(), asr.timeout)
	defer cancel()
	rows, err := asr.db.QueryContext(ctx, query, offset, limit)
	if err != nil {
		return nil, -1, fmt.Errorf("query [%s], error %v", query, err)
	}
	defer rows.Close()
	result := make([]domain.Asset, 0, limit)
	for rows.Next() {
		item := domain.Asset{}
		err = rows.Scan(&item.ID, &item.Name, &item.ServiceDeskParentID, &item.Changed, &item.ServiceDeskID)
		if err != nil {
			if err == sql.ErrNoRows {
				return result, -1, nil
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

// FindByID returns single Asset with specified id from Intraservice database
func (asr *AssetSQLRepo) FindByID(id int64) (*domain.Asset, error) {
	query := `SELECT Data.value('(/data/field[@id=62])[1]', 'int') as ID,
					Name,
					ParentId as ServiceDeskParentID,
					Changed,
					Id as ServiceDeskID 
			FROM [dbo].[Asset]
			WHERE Data.value('(/data/field[@id=62])[1]', 'int') = ?`
	ctx, cancel := context.WithTimeout(context.Background(), asr.timeout)
	defer cancel()
	item := domain.Asset{}
	err := asr.db.QueryRowContext(ctx, query, id).
		Scan(&item.ID, &item.Name, &item.ServiceDeskParentID, &item.Changed, &item.ServiceDeskID)
	if err != nil {
		return nil, fmt.Errorf("query [%s], error %v", query, err)
	}
	return &item, nil
}
