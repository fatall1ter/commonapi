package repos

import (
	"context"
	"fmt"
	"time"

	"git.countmax.ru/countmax/commonapi/domain"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type RefPGSQLRepo struct {
	connString string
	timeout    time.Duration
	db         *pgxpool.Pool
}

// NewRefPGRepo returns instance of PGRepo with connected evolution DB,
// format connection string: "postgres://jack:secret@pg.example.com:5432/mydb?sslmode=verify-ca&pool_max_conns=2"
func NewRefPGRepo(connString string, timeout time.Duration) (*RefPGSQLRepo, error) {
	rpr := &RefPGSQLRepo{
		connString: connString,
		timeout:    timeout,
	}
	err := rpr.connDB()
	if err != nil {
		return nil, err
	}
	return rpr, nil
}

// connDB - make connection to database
func (rpr *RefPGSQLRepo) connDB() error {
	config, err := pgxpool.ParseConfig(rpr.connString)
	if err != nil {
		return err
	}
	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return err
	}
	rpr.db = pool
	return nil
}

// Close - close connection to database
func (rpr *RefPGSQLRepo) Close() {
	rpr.db.Close()
}

// Health - check connection to DB
func (rpr *RefPGSQLRepo) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), rpr.timeout)
	defer cancel()
	var tmp string
	return rpr.db.QueryRow(ctx, "SELECT version();").Scan(&tmp)
}

// FindEntities returns slice of Entity from evolution database
func (rpr *RefPGSQLRepo) FindEntities(offset, limit int64) (domain.Entities, int64, error) {
	query := "select entity_id, description from entities limit $1 offset $2;"
	batch := &pgx.Batch{}
	batch.Queue(query, limit, offset)
	batch.Queue("select count(*) from entities")
	ctx, cancel := context.WithTimeout(context.Background(), rpr.timeout)
	defer cancel()
	batchRes := rpr.db.SendBatch(ctx, batch)
	rows, err := batchRes.Query()
	if err != nil {
		return nil, 0, err
	}
	result := make(domain.Entities, 0, limit)
	for rows.Next() {
		item := domain.Entity{}
		err = rows.Scan(&item.ID, &item.Description)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, item)
	}
	var count int64
	err = batchRes.QueryRow().Scan(&count)
	if err != nil {
		return nil, 0, err
	}
	return result, count, nil
}

// FindEntityByID returns single entity with specified id from evolution database
func (rpr *RefPGSQLRepo) FindEntityByID(id string) (*domain.Entity, error) {
	query := `select entity_id, description from entities where entity_id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), rpr.timeout)
	defer cancel()
	item := domain.Entity{}
	err := rpr.db.QueryRow(ctx, query, id).Scan(&item.ID, &item.Description)
	if err != nil {
		return nil, fmt.Errorf("query [%s], error %v", query, err)
	}
	return &item, nil
}
