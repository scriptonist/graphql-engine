package common

type Common struct {
	*catalogState
}

func New() *Common {
	return &Common{
		newCatalogState(),
	}
}
