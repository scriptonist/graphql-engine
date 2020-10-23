package v1

import (
	"github.com/hasura/graphql-engine/cli/internal/client/v1/common"
	"github.com/hasura/graphql-engine/cli/internal/client/v1/datasource"
)

const DefaultMetadataAPIPath = "v1/metadata"

type Metadata struct {
	*common.Common
	*bulk
	DatasourceAPI func(backend datasource.Backend) datasource.Datasource
}

func New() *Metadata {
	return &Metadata{
		common.New(),
		NewBulk(),
		datasource.New,
	}
}
