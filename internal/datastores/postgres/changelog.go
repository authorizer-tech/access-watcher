package postgres

import (
	"context"
	"time"

	watcher "github.com/authorizer-tech/access-watcher/internal"
	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type changelogDatastore struct {
	pool *pgxpool.Pool
}

func NewChangelogDatastore(p *pgxpool.Pool) (watcher.ChangelogDatastore, error) {

	c := &changelogDatastore{
		pool: p,
	}
	return c, nil
}

func (ds *changelogDatastore) GetRelationTupleChanges(ctx context.Context, namespaces []string, timestamp time.Time) (watcher.ChangelogIterator, error) {

	sqlbuilder := goqu.Dialect("postgres").From("changelog").Select(
		"namespace", "operation", "relationtuple", "timestamp",
	).Where(
		goqu.Ex{
			"namespace": namespaces,
			"timestamp": goqu.Op{"gte": timestamp},
		},
	).Order(goqu.C("timestamp").Asc())

	sql, params, err := sqlbuilder.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := ds.pool.Query(ctx, sql, params...)
	if err != nil {
		return nil, err
	}

	iter := &iterator{rows}
	return iter, nil
}

type iterator struct {
	rows pgx.Rows
}

func (i *iterator) Next() bool {
	return i.rows.Next()
}

func (i *iterator) Value() (*watcher.ChangelogEntry, error) {
	var namespace, operation, relationtuple string
	var timestamp time.Time
	if err := i.rows.Scan(&namespace, &operation, &relationtuple, &timestamp); err != nil {
		return nil, err
	}

	tuple, err := watcher.RelationTupleFromString(relationtuple)
	if err != nil {
		return nil, err
	}

	e := &watcher.ChangelogEntry{
		Namespace:     namespace,
		Operation:     operation,
		RelationTuple: tuple,
		Timestamp:     timestamp,
	}
	return e, nil
}
