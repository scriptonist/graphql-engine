package common

const DefaultCommonMetadataAPIPath = "v1/metadata"

type Common struct {
	CatalogState *catalogState
}

func New() *Common {
	return &Common{
		newCatalogState(),
	}
}
