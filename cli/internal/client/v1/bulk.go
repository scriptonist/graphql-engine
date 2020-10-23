package v1

import "net/http"

type bulk struct {
	path   string
	method string
}

func NewBulk() *bulk {
	return &bulk{
		path:   DefaultMetadataAPIPath,
		method: http.MethodPost,
	}
}
func (b *bulk) Send(request bulkRequest) (*BulkResponse, error) {
	return &BulkResponse{}, nil
}

type BulkRequestArgs []interface{}

type bulkRequest struct {
	Type string
	Args BulkRequestArgs
}

func NewBulkRequest(args BulkRequestArgs) bulkRequest {
	return bulkRequest{
		Type: "bulk",
		Args: args,
	}
}

type BulkResponse []interface{}
