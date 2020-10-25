package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/hasura/graphql-engine/cli/internal/client"
)

type catalogState struct {
	Get func(client *client.Client, request getCatalogStateRequest) (*GetCatalogStateResponse, error)
	Set func(client *client.Client, request setCatalogStateRequest) (*SetCatalogStateResponse, error)
}

func newCatalogState() *catalogState {
	gcs := getCatalogState{DefaultCommonMetadataAPIPath, http.MethodPost}
	scs := setCatalogState{DefaultCommonMetadataAPIPath, http.MethodPost}
	return &catalogState{gcs.GetCatalogState, scs.SetCatalogState}
}

type getCatalogState struct {
	// api path
	path string
	// http method
	method string
}

type getCatalogStateRequest struct {
	Type string                 `json:"type"`
	Args map[string]interface{} `json:"args"`
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

type errGetCatalogState map[string]interface{}

func (err errGetCatalogState) Error() string {
	var errors []string
	for k, v := range err {
		errors = append(errors, fmt.Sprintf("%s: %s", k, v))
	}
	return strings.Join(errors, "\n")
}

func (gcs *getCatalogState) GetCatalogState(client *client.Client, requestBody getCatalogStateRequest) (*GetCatalogStateResponse, error) {
	req, err := client.NewRequest(gcs.method, gcs.path, requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "constructing get catalog request")
	}

	var respBody = new(bytes.Buffer)
	resp, err := client.Do(context.Background(), req, respBody)
	if err != nil {
		return nil, errors.Wrap(err, "making api request")
	}
	if resp.StatusCode == http.StatusBadRequest {
		apiResponseErr := new(errGetCatalogState)
		err := json.Unmarshal(respBody.Bytes(), apiResponseErr)
		if err != nil {
			return nil, errors.Wrap(err, "decoding api response error")
		}
		return nil, apiResponseErr
	} else if resp.StatusCode != http.StatusOK {
		if s := respBody.String(); s != "" {
			return nil, fmt.Errorf("%s", s)
		}
	}

	var catalogStateResponse = new(GetCatalogStateResponse)
	if err := json.Unmarshal(respBody.Bytes(), catalogStateResponse); err != nil {
		return nil, errors.Wrap(err, "decoding api response")
	}
	return catalogStateResponse, nil
}

type setCatalogState struct {
	// api path
	path string
	// http method
	method string
}

func (gcs *setCatalogState) SetCatalogState(client *client.Client, request setCatalogStateRequest) (*SetCatalogStateResponse, error) {
	return &SetCatalogStateResponse{}, nil
}

type catalogStateBackend string

type SetCatalogStateArgs struct {
	Type  catalogStateBackend    `json:"type" mapstructure:"type,omitempty"`
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
