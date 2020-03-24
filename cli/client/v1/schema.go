package v1

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const schemaMetadataAPIEndpoint = "v1/query"

// SendQuery does what the name implies
func (c *Client) SendQuery(m interface{}) (*http.Response, []byte, *Error) {
	request, err := c.NewRequest("POST", schemaMetadataAPIEndpoint, m)
	if err != nil {
		return nil, nil, E(err)
	}

	resp, err := c.client.Do(request)
	if err != nil {
		return nil, nil, E(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, E(err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiError APIError
		err := json.Unmarshal(body, &apiError)
		if err != nil {
			return nil, nil, E(err)
		}
		return nil, nil, E(apiError)
	}
	return resp, body, nil
}
