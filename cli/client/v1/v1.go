package v1

import (
	"net/url"
	"path"
	"strings"
)

// Client will implement the hasura V1 API
type Client struct {
	SchemaAndMetadataAPIEndpoint url.URL
	PGDumpAPIEndpoint            url.URL
	Headers                      map[string]string
}

// NewClient when provided a URL will return a provider with
// default config
func NewClient(hasuraAPIEndpoint string) (*Client, error) {
	parsedHasuraAPIEndpoint, err := url.Parse(hasuraAPIEndpoint)
	if err != nil {
		return nil, nil
	}

	client := new(Client)

	params := parsedHasuraAPIEndpoint.Query()
	headers := make(map[string]string)
	if queryHeaders, ok := params["headers"]; ok {
		for _, header := range queryHeaders {
			headerValue := strings.SplitN(header, ":", 2)
			if len(headerValue) == 2 && headerValue[1] != "" {
				headers[headerValue[0]] = headerValue[1]
			}
		}
	}

	client.Headers = headers

	// Use sslMode query param to set Scheme
	var scheme string
	sslMode := params.Get("sslmode")
	if sslMode == "enable" {
		scheme = "https"
	} else {
		scheme = "http"
	}

	const schemaAndMetadataAPIEndpoint = "v1/query"
	client.SchemaAndMetadataAPIEndpoint = *parsedHasuraAPIEndpoint
	client.SchemaAndMetadataAPIEndpoint.Scheme = scheme
	client.SchemaAndMetadataAPIEndpoint.Path = path.Join(parsedHasuraAPIEndpoint.Path, schemaAndMetadataAPIEndpoint)

	const pgDumpAPIEndpoint = "v1/pg_dump"
	client.PGDumpAPIEndpoint = *parsedHasuraAPIEndpoint
	client.PGDumpAPIEndpoint.Scheme = scheme
	client.PGDumpAPIEndpoint.Path = path.Join(parsedHasuraAPIEndpoint.Path, pgDumpAPIEndpoint)

	return client, nil
}
