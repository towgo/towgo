package jsonrpc

import (
	"encoding/json"
)

// ContextKey is the key type used by the JSON-RPC protocol ctx field.
type ContextKey string

const (
	// SESSION is the protocol ctx key used for session values.
	SESSION = "jsonrpc.session"
	// JSON_RPC_CONNECTION_CONTEXT_KEY is the context key used by legacy integrations.
	JSON_RPC_CONNECTION_CONTEXT_KEY = "json_rpc_connection"
	// DEFAULT_ERROR_MSG is the fallback network error message.
	DEFAULT_ERROR_MSG = "network error"
)

// JsonrpcRoute carries optional routing metadata inside a JSON-RPC message.
type JsonrpcRoute struct {
	// TTL limits how many route hops a message may take.
	TTL int64 `json:"TTL,omitempty"`
	// DestAddr is the desired destination node or address.
	DestAddr string `json:"DestAddr,omitempty"`
	// SourceAddr is the origin node or address.
	SourceAddr string `json:"SourceAddr,omitempty"`
}

// RpcRoute is an alias for JsonrpcRoute.
type RpcRoute = JsonrpcRoute

// Jsonrpcrequest is the JSON-RPC request payload.
//
// Runtime context.Context values intentionally live on Request or
// JsonRpcConnection, not on this protocol struct.
type Jsonrpcrequest struct {
	// Jsonrpc is the JSON-RPC protocol version.
	Jsonrpc string `json:"jsonrpc"`
	// Method is the target JSON-RPC method path.
	Method string `json:"method"`
	// DataType preserves special payload typing for compatibility.
	DataType string `json:"DataType"`
	// Params carries request parameters.
	Params any `json:"params"`
	// Id correlates requests and responses.
	Id string `json:"id"`
	// Ctx carries protocol-level business context.
	Ctx map[ContextKey]any `json:"ctx"`
	// Session carries a protocol session value.
	Session string `json:"session"`
	// Timestampin records the inbound request timestamp.
	Timestampin string `json:"timestampin"`
	// Timestampout records the outbound response timestamp.
	Timestampout string `json:"timestampout"`
	// Isencryption keeps compatibility with older encrypted payloads.
	Isencryption bool `json:"-"`
	// Route carries optional protocol routing metadata.
	Route JsonrpcRoute `json:"route"`
}

// Jsonrpcresponse is the JSON-RPC response payload.
type Jsonrpcresponse struct {
	// Jsonrpc is the JSON-RPC protocol version.
	Jsonrpc string `json:"jsonrpc"`
	// DataType preserves special payload typing for compatibility.
	DataType string `json:"DataType"`
	// Result carries the successful response value.
	Result any `json:"result"`
	// Id correlates responses with requests.
	Id string `json:"id"`
	// Ctx carries protocol-level business context.
	Ctx map[ContextKey]any `json:"ctx"`
	// Timestampin mirrors the inbound request timestamp.
	Timestampin string `json:"timestampin"`
	// Timestampout records the outbound response timestamp.
	Timestampout string `json:"timestampout"`
	// Route carries optional protocol routing metadata.
	Route JsonrpcRoute `json:"route"`
	// Error carries the JSON-RPC error state.
	Error Error `json:"error"`
}

// Jsonrpcresponseclient is a loose response shape for clients with dynamic errors.
type Jsonrpcresponseclient struct {
	// Jsonrpc is the JSON-RPC protocol version.
	Jsonrpc string `json:"jsonrpc"`
	// Result carries the successful response value.
	Result any `json:"result"`
	// Id correlates responses with requests.
	Id string `json:"id"`
	// Ctx carries loose client-side protocol context.
	Ctx map[string]any `json:"ctx"`
	// Timestampin mirrors the inbound request timestamp.
	Timestampin string `json:"timestampin"`
	// Timestampout records the outbound response timestamp.
	Timestampout string `json:"timestampout"`
	// Route carries optional protocol routing metadata.
	Route JsonrpcRoute `json:"route"`
	// Error carries dynamic client-side error data.
	Error any `json:"error"`
}

// NewJsonrpcresponse creates a successful JSON-RPC response shell.
func NewJsonrpcresponse() *Jsonrpcresponse {
	return &Jsonrpcresponse{
		Jsonrpc: "2.0",
		Error: Error{
			Code:    JSONRPC_200_OK,
			Message: "ok",
		},
	}
}

// NewJsonrpcrequest creates a JSON-RPC request shell.
func NewJsonrpcrequest() *Jsonrpcrequest {
	return &Jsonrpcrequest{
		Jsonrpc: "2.0",
	}
}

// ToJsonrpcrequest parses a JSON string into a JSON-RPC request.
func ToJsonrpcrequest(s string) (*Jsonrpcrequest, error) {
	jsonrpcrequest := NewJsonrpcrequest()
	if err := json.Unmarshal([]byte(s), jsonrpcrequest); err != nil {
		return jsonrpcrequest, err
	}
	if jsonrpcrequest.Jsonrpc == "" {
		jsonrpcrequest.Jsonrpc = "2.0"
	}
	if jsonrpcrequest.DataType != "" {
		jsonrpcrequest.Params = ProcessSourceData(jsonrpcrequest.DataType, s, true)
	}
	return jsonrpcrequest, nil
}

// ToJsonrpcresponse parses a JSON string into a JSON-RPC response.
func ToJsonrpcresponse(s string) (Jsonrpcresponse, error) {
	jsonrpcresponse := *NewJsonrpcresponse()
	if err := json.Unmarshal([]byte(s), &jsonrpcresponse); err != nil {
		return jsonrpcresponse, err
	}
	if jsonrpcresponse.Jsonrpc == "" {
		jsonrpcresponse.Jsonrpc = "2.0"
	}
	if jsonrpcresponse.DataType != "" {
		jsonrpcresponse.Result = ProcessSourceData(jsonrpcresponse.DataType, s, false)
	}
	return jsonrpcresponse, nil
}

// ToJsonrpcresponseclient parses a JSON string into a loose client response.
func ToJsonrpcresponseclient(s string) (Jsonrpcresponseclient, error) {
	var jsonrpcresponse Jsonrpcresponseclient
	err := json.Unmarshal([]byte(s), &jsonrpcresponse)
	return jsonrpcresponse, err
}

// ProcessSourceData extracts typed params or result data from a raw JSON message.
func ProcessSourceData(dataType string, jsonStr string, isRequest bool) any {
	var data struct {
		Params []byte `json:"params"`
		Result []byte `json:"result"`
	}
	switch dataType {
	case "[]byte":
		_ = json.Unmarshal([]byte(jsonStr), &data)
	}
	if isRequest {
		return data.Params
	}
	return data.Result
}

// ReadResult unmarshals the response result into destParams.
func (jrr *Jsonrpcresponse) ReadResult(destParams any) error {
	paramsBytes, err := json.Marshal(jrr.Result)
	if err != nil {
		return err
	}
	return json.Unmarshal(paramsBytes, destParams)
}

// SetSecretkey is reserved for compatibility with earlier encrypted JSON-RPC APIs.
func SetSecretkey(key string, iv string) {
}
