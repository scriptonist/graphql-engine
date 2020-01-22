package v1

import (
	"net/http"

	"github.com/parnurzeal/gorequest"
)

// SendQuery does what the name implies
func (h *Client) SendQuery(m interface{}) (resp *http.Response, body []byte, err error) {
	request := gorequest.New()

	request = request.Post(h.SchemaAndMetadataAPIEndpoint.String()).Send(m)

	for headerName, headerValue := range h.Headers {
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
