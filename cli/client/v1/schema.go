package v1

import (
	"net/http"
	"net/url"
	"path"

	"github.com/parnurzeal/gorequest"
)

// ClientMetadataAndSchema API
type ClientMetadataAndSchema struct {
	SchemaAndMetadataAPIEndpoint url.URL
	Headers                      map[string]string
}

// NewClientMetadataAndSchema returns  a pointer to the respective struct
func NewClientMetadataAndSchema(hasuraAPIEndpoint url.URL, headers map[string]string) *ClientMetadataAndSchema {
	client := new(ClientMetadataAndSchema)

	const schemaAndMetadataAPIEndpoint = "v1/query"
	client.SchemaAndMetadataAPIEndpoint = hasuraAPIEndpoint
	client.SchemaAndMetadataAPIEndpoint.Scheme = hasuraAPIEndpoint.Scheme
	client.SchemaAndMetadataAPIEndpoint.Path = path.Join(hasuraAPIEndpoint.Path, schemaAndMetadataAPIEndpoint)
	return client
}

// SendQuery does what the name implies
func (client *ClientMetadataAndSchema) SendQuery(m interface{}) (resp *http.Response, body []byte, err error) {
	request := gorequest.New()

	request = request.Post(client.SchemaAndMetadataAPIEndpoint.String()).Send(m)

	for headerName, headerValue := range client.Headers {
		request.Set(headerName, headerValue)
	}

	resp, body, errs := request.EndBytes()

	if len(errs) == 0 {
		err = nil
	} else {
		err = errs[0]
	}

	return resp, body, err
}
