package metadata

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hasura/graphql-engine/cli/internal/client"
)

type BulkQuery struct {
	BulkQueryRequest
	BulkQueryHTTP
	*BulkQueryResponse
	*ErrBulkQuery
}

func (b BulkQuery) Send(client *client.Client) error {
	queryType := NewQueryType(
		b.BulkQueryRequest,
		b.BulkQueryResponse,
		b.ErrBulkQuery,
		b.BulkQueryHTTP,
	)
	return SendMetadataQuery(client, queryType)
}

type ErrBulkQuery []ResponseError

func (bulkQueryErrors *ErrBulkQuery) UnmarshalResponseErrorJSON(b []byte) error {
	return json.Unmarshal(b, bulkQueryErrors)
}

func (bulkQueryErrors ErrBulkQuery) Error() string {
	errors := ""
	for _, e := range bulkQueryErrors {
		errors += fmt.Sprintf("%s\n", e.Error())
	}
	return errors
}

type BulkQueryRequest struct {
	args []RequestBody
}

func (b BulkQueryRequest) Type() string {
	return "bulk"
}

func (b BulkQueryRequest) Args() interface{} {
	return b.args
}

type BulkQueryResponse []ResponseBody

func (b *BulkQueryResponse) UnmarshalResponseBodyJSON(data []byte) error {
	return json.Unmarshal(data, b)
}

type BulkQueryHTTP struct {
}

func (b BulkQueryHTTP) Method() string {
	return http.MethodPost
}

func (b BulkQueryHTTP) Path() string {
	return defaultMetadataAPIPath
}
