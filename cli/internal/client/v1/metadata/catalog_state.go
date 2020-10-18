package metadata

import "fmt"

type CatalogState interface {
	GetCatalogState() (*GetCatalogStateResponseBody, error)
	SetCatalogState(args *SetCatalogStateRequestArgs) (*SetCatalogStateResponseBody, error)
}

type CatalogStateErr map[string]interface{}

func (catalogStateError CatalogStateErr) Error() string {
	var error string
	for k, v := range catalogStateError {
		error += fmt.Sprintf("%v: %v\n", k, v)
	}
	return error
}

type CatalogStateBackend string

const CatalogStateCLIBackend CatalogStateBackend = "cli"
const CatalogStateConsoleBackend CatalogStateBackend = "console"

// get catalog state
//
// request
type GetCatalogStateRequest struct {
	args map[string]interface{}
}

func (s *GetCatalogStateRequest) Type() string {
	return "get_catalog_state"
}

func (s *GetCatalogStateRequest) Args() interface{} {
	return s.args
}

// response
//
type GetCatalogStateResponseBody struct {
	ID           string                  `json:"id" mapstructure:"id,omitempty"`
	CLIState     *map[string]interface{} `json:"cli_state,omitempty" mapstructure:"cli_state,omitempty"`
	ConsoleState *map[string]interface{} `json:"console_state,omitempty" mapstructure:"console_state,omitempty"`
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
