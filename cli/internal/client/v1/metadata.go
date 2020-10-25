package v1

import (
	"github.com/hasura/graphql-engine/cli/internal/client/v1/common"
	"github.com/hasura/graphql-engine/cli/internal/client/v1/datasource"
)

type Metadata struct {
	// Datasource agnostic API's
	Common        *common.Common
	Bulk          *bulk
	DatasourceAPI func(backend datasource.Backend) datasource.Datasource
}

func New() *Metadata {
	return &Metadata{
		common.New(),
		NewBulk(),
		datasource.New,
	}
}
