package metadata

import (
	"net/http"

	"github.com/hasura/graphql-engine/cli/internal/client"
)

type Common interface {
	CatalogState
}

// RequestType is the interface which has to be satisfied by any type
// which is supposed to be a metadata request type supported by hasura
type RequestType interface {
	Type() string
	Args() interface{}
}

func NewMetadataHTTPRequest(c *client.Client, path, method string, body RequestType) (*http.Request, error) {
	var requestBody = struct {
		Type string
		Args interface{}
	}{
		body.Type(),
		body.Args(),
	}
	req, err := c.NewRequest(method, path, requestBody)
	if err != nil {
		return nil, err
	}

	return req, nil
}
