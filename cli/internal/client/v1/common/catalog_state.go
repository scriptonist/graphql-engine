package common

import (
	"net/http"

	v1 "github.com/hasura/graphql-engine/cli/internal/client/v1"

	"github.com/hasura/graphql-engine/cli/internal/client"
)

type catalogState struct {
	Get func(client *client.Client, request getCatalogStateRequest) (GetCatalogStateResponse, error)
	Set func(client *client.Client, request setCatalogStateRequest) (SetCatalogStateResponse, error)
}

func newCatalogState() *catalogState {
	gcs := getCatalogState{v1.DefaultMetadataAPIPath, http.MethodPost}
	scs := setCatalogState{v1.DefaultMetadataAPIPath, http.MethodPost}
	return &catalogState{gcs.GetCatalogState, scs.SetCatalogState}
}

type getCatalogState struct {
	// api path
	path string
	// http method
	method string
}

type getCatalogStateRequest struct {
	Type string
	Args map[string]interface{}
}

func NewGetCatalogStateRequest() getCatalogStateRequest {
	return getCatalogStateRequest{
		Type: "get_catalog_state",
		Args: map[string]interface{}{},
	}
}

type GetCatalogStateResponse struct {
	ID           string                  `json:"id" mapstructure:"id,omitempty"`
	CLIState     *map[string]interface{} `json:"cli_state,omitempty" mapstructure:"cli_state,omitempty"`
	ConsoleState *map[string]interface{} `json:"console_state,omitempty" mapstructure:"console_state,omitempty"`
}

func (gcs *getCatalogState) GetCatalogState(client *client.Client, request getCatalogStateRequest) (GetCatalogStateResponse, error) {
	return GetCatalogStateResponse{}, nil
}

type setCatalogState struct {
	// api path
	path string
	// http method
	method string
}

func (gcs *setCatalogState) SetCatalogState(client *client.Client, request setCatalogStateRequest) (SetCatalogStateResponse, error) {
	return SetCatalogStateResponse{}, nil
}

type CatalogStateBackend string

type SetCatalogStateArgs struct {
	Type  CatalogStateBackend    `json:"type" mapstructure:"type,omitempty"`
	State map[string]interface{} `json:"state" mapstructure:"state,omitempty"`
}

type setCatalogStateRequest struct {
	Type string
	Args SetCatalogStateArgs
}

type SetCatalogStateResponse struct {
	Message string `json:"message" mapstructure:"message,omitempty"`
}

func NewSetCatalogStateRequest(args SetCatalogStateArgs) setCatalogStateRequest {
	return setCatalogStateRequest{
		Type: "set_catalog_state",
		Args: args,
	}
}
