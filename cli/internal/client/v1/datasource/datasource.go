package datasource

import (
	"github.com/hasura/graphql-engine/cli/internal/client/v1/datasource/mysql"
	"github.com/hasura/graphql-engine/cli/internal/client/v1/datasource/pg"
)

type Datasource interface {
}

type Backend string

const (
	PG    Backend = "pg"
	MySQL Backend = "mysql"
)

func New(backend Backend) Datasource {
	switch backend {
	case PG:
		return pg.New()
	case MySQL:
		return mysql.New()
	default:
		return nil
	}
}
