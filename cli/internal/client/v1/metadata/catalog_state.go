package metadata

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hasura/graphql-engine/cli/internal/client"
)

// GetCatalogStateQuery
//
type GetCatalogStateQuery struct {
	GetCatalogStateRequest
	GetCatalogStateHTTP
	*GetCatalogStateResponseBody
	*ErrGetCatalogStateQuery
}

func NewGetCatalogStateQuery(apiPath string) *GetCatalogStateQuery {
	return &GetCatalogStateQuery{
		GetCatalogStateRequest:      GetCatalogStateRequest{},
		GetCatalogStateHTTP:         NewGetCatalogStateHTTP(apiPath),
		GetCatalogStateResponseBody: new(GetCatalogStateResponseBody),
		ErrGetCatalogStateQuery:     new(ErrGetCatalogStateQuery),
	}
}

func (getCatalogStateQuery *GetCatalogStateQuery) Send(client *client.Client) error {
	queryType := NewQueryType(
		getCatalogStateQuery.GetCatalogStateRequest,
		getCatalogStateQuery.GetCatalogStateResponseBody,
		getCatalogStateQuery.ErrGetCatalogStateQuery,
		getCatalogStateQuery.GetCatalogStateHTTP,
	)
	return SendMetadataQuery(client, queryType)
}

type ErrGetCatalogStateQuery map[string]interface{}

func (errGetCatalogStateQuery ErrGetCatalogStateQuery) Error() string {
	var error string
	for k, v := range errGetCatalogStateQuery {
		error += fmt.Sprintf("%v: %v\n", k, v)
	}
	return error
}

func (errGetCatalogStateQuery *ErrGetCatalogStateQuery) UnmarshalResponseErrorJSON(b []byte) error {
	return json.Unmarshal(b, errGetCatalogStateQuery)
}

type GetCatalogStateHTTP struct {
	path string
}

func NewGetCatalogStateHTTP(path string) GetCatalogStateHTTP {
	return GetCatalogStateHTTP{
		path: path,
	}
}

func (getCatalogStateHTTP GetCatalogStateHTTP) Method() string {
	return http.MethodPost
}

func (getCatalogStateHTTP GetCatalogStateHTTP) Path() string {
	if getCatalogStateHTTP.path == "" {
		return defaultMetadataAPIPath
	}
	return getCatalogStateHTTP.path
}

type CatalogStateBackend string

const CatalogStateCLIBackend CatalogStateBackend = "cli"
const CatalogStateConsoleBackend CatalogStateBackend = "console"

// request
//
type GetCatalogStateRequest struct {
	args map[string]interface{}
}

func (s GetCatalogStateRequest) Type() string {
	return "get_catalog_state"
}

func (s GetCatalogStateRequest) Args() interface{} {
	return s.args
}

// response
//
type GetCatalogStateResponseBody struct {
	ID           string                  `json:"id" mapstructure:"id,omitempty"`
	CLIState     *map[string]interface{} `json:"cli_state,omitempty" mapstructure:"cli_state,omitempty"`
	ConsoleState *map[string]interface{} `json:"console_state,omitempty" mapstructure:"console_state,omitempty"`
}

func (getCatalogStateResponseBody *GetCatalogStateResponseBody) UnmarshalResponseBodyJSON(b []byte) error {
	return json.Unmarshal(b, getCatalogStateResponseBody)
}

// set catalog state
//
// request
type SetCatalogStateRequest struct {
	args *SetCatalogStateRequestArgs
}

func NewSetCatalogStateRequest(args *SetCatalogStateRequestArgs) *SetCatalogStateRequest {
	return &SetCatalogStateRequest{args}
}

type SetCatalogStateRequestArgs struct {
	Type  CatalogStateBackend    `json:"type" mapstructure:"type,omitempty"`
	State map[string]interface{} `json:"state" mapstructure:"state,omitempty"`
}

func (s *SetCatalogStateRequest) Type() string {
	return "set_catalog_state"
}

func (s *SetCatalogStateRequest) Args() interface{} {
	return s.args
}

// response
//
type SetCatalogStateResponseBody struct {
	Message string `json:"message" mapstructure:"message,omitempty"`
}
