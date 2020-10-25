package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/hasura/graphql-engine/cli/internal/client"

	"github.com/pkg/errors"
)

const DefaultBulkMetadataAPIPath = "v1/metadata"

type bulk struct {
	path   string
	method string
}

func NewBulk() *bulk {
	return &bulk{
		path:   DefaultBulkMetadataAPIPath,
		method: http.MethodPost,
	}
}
func (b *bulk) Send(client *client.Client, request bulkRequest) (*bulkResponse, error) {
	req, err := client.NewRequest(b.method, b.path, request)
	if err != nil {
		return nil, errors.Wrap(err, "constructing get catalog request")
	}

	var respBody = new(bytes.Buffer)
	resp, err := client.Do(context.Background(), req, respBody)
	if err != nil {
		return nil, errors.Wrap(err, "making api request")
	}
	if resp.StatusCode != http.StatusOK {
		errors.New(respBody.String())
	}

	var bulkResponse = new(bulkResponse)
	if err := json.Unmarshal(respBody.Bytes(), bulkResponse); err != nil {
		return nil, errors.Wrap(err, "decoding api response")
	}
	return bulkResponse, nil
}

type BulkRequestArgs []interface{}

type bulkRequest struct {
	Type string          `json:"type"`
	Args BulkRequestArgs `json:"args"`
}

func NewBulkRequest(args BulkRequestArgs) bulkRequest {
	return bulkRequest{
		Type: "bulk",
		Args: args,
	}
}

type bulkResponse []interface{}
