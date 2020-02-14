package v1

import (
	"net/http"
	"net/url"
	"path"

	"github.com/parnurzeal/gorequest"
)

// ClientPGDump API
type ClientPGDump struct {
	PGDumpAPIEndpoint url.URL
	Headers           map[string]string
}

// NewClientPGDump returns  a pointer to the respective struct
func NewClientPGDump(hasuraAPIEndpoint url.URL, headers map[string]string) *ClientPGDump {
	client := new(ClientPGDump)
	const pgDumpAPIEndpoint = "/v1alpha1/pg_dump"
	client.PGDumpAPIEndpoint = hasuraAPIEndpoint
	client.PGDumpAPIEndpoint.Scheme = hasuraAPIEndpoint.Scheme
	client.PGDumpAPIEndpoint.Path = path.Join(hasuraAPIEndpoint.Path, pgDumpAPIEndpoint)
	return client
}

// SendPGDumpQuery --
func (client *ClientPGDump) SendPGDumpQuery(m interface{}) (resp *http.Response, body []byte, err error) {
	request := gorequest.New()

	request = request.Post(client.PGDumpAPIEndpoint.String()).Send(m)

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
