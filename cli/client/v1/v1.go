package v1

import (
	"net/url"
	"strings"
)

// Client will implement the hasura V1 API
type Client struct {
	*ClientMetadataAndSchema
	*ClientPGDump
}

// NewClient when provided a URL will return a API Client with default config
func NewClient(hasuraAPIEndpoint string) (*Client, error) {
	parsedHasuraAPIEndpoint, headers, err := ParseHasuraAPIEndpoint(hasuraAPIEndpoint)
	if err != nil {
		return nil, err
	}
	client := new(Client)
	client.ClientMetadataAndSchema = NewClientMetadataAndSchema(*parsedHasuraAPIEndpoint, *headers)
	client.ClientPGDump = NewClientPGDump(*parsedHasuraAPIEndpoint, *headers)
	return client, nil
}

// ParseHasuraAPIEndpoint and return the parsed URL and Headers (extracted from query strings)
func ParseHasuraAPIEndpoint(hasuraAPIEndpoint string) (*url.URL, *map[string]string, error) {
	parsedHasuraAPIEndpoint, err := url.Parse(hasuraAPIEndpoint)
	if err != nil {
		return nil, nil, err
	}

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

	// Use sslMode query param to set Scheme
	var scheme string
	sslMode := params.Get("sslmode")
	if sslMode == "enable" {
		scheme = "https"
	} else {
		scheme = "http"
	}
	parsedHasuraAPIEndpoint.Scheme = scheme

	return parsedHasuraAPIEndpoint, &headers, nil
}
