package seed

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/parnurzeal/gorequest"
)

// HasuraV1APIProvider will satisfy the HasuraAPIProviderInterface
type HasuraV1APIProvider struct {
	v1URL   *url.URL
	Headers map[string]string
}

// NewHasuraV1APIProvider when provided a URL will return a provider with
// default config
func NewHasuraV1APIProvider(hasuraURL string) (*HasuraV1APIProvider, error) {
	parsedHasuraURL, err := url.Parse(hasuraURL)
	if err != nil {
		return nil, nil
	}

	hasuraV1APIProvider := new(HasuraV1APIProvider)
	hasuraV1APIProvider.v1URL = parsedHasuraURL

	params := parsedHasuraURL.Query()
	headers := make(map[string]string)
	if queryHeaders, ok := params["headers"]; ok {
		for _, header := range queryHeaders {
			headerValue := strings.SplitN(header, ":", 2)
			if len(headerValue) == 2 && headerValue[1] != "" {
				headers[headerValue[0]] = headerValue[1]
			}
		}
	}

	hasuraV1APIProvider.Headers = headers

	// Use sslMode query param to set Scheme
	var scheme string
	sslMode := params.Get("sslmode")
	if sslMode == "enable" {
		scheme = "https"
	} else {
		scheme = "http"
	}
	hasuraV1APIProvider.v1URL.Scheme = scheme
	hasuraV1APIProvider.v1URL.Path = path.Join(parsedHasuraURL.Path, "v1/query")
	return hasuraV1APIProvider, nil
}

// SendQuery does what the name implies
func (h *HasuraV1APIProvider) SendQuery(m interface{}) (resp *http.Response, body []byte, err error) {
	request := gorequest.New()

	request = request.Post(h.v1URL.String()).Send(m)

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
