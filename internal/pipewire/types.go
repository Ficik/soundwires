package pipewire

import "encoding/json"

const (
	TypeNode   = "PipeWire:Interface:Node"
	TypePort   = "PipeWire:Interface:Port"
	TypeLink   = "PipeWire:Interface:Link"
	TypeDevice = "PipeWire:Interface:Device"
	TypeClient = "PipeWire:Interface:Client"
)

type PWObject struct {
	ID          int             `json:"id"`
	Type        string          `json:"type"`
	Permissions []string        `json:"permissions"`
	Info        json.RawMessage `json:"info,omitempty"`
	Props       json.RawMessage `json:"props,omitempty"`
}

// InfoNode is the parsed "info" block for Node objects.
type InfoNode struct {
	MaxInputPorts  int             `json:"max-input-ports"`
	MaxOutputPorts int             `json:"max-output-ports"`
	NInputPorts    int             `json:"n-input-ports"`
	NOutputPorts   int             `json:"n-output-ports"`
	State          string          `json:"state"`
	Error          *string         `json:"error"`
	Props          json.RawMessage `json:"props"`
}

// InfoPort is the parsed "info" block for Port objects.
type InfoPort struct {
	Direction string          `json:"direction"`
	Props     json.RawMessage `json:"props"`
}

// InfoLink is the parsed "info" block for Link objects.
type InfoLink struct {
	OutputNodeID int    `json:"output-node-id"`
	OutputPortID int    `json:"output-port-id"`
	InputNodeID  int    `json:"input-node-id"`
	InputPortID  int    `json:"input-port-id"`
	State        string `json:"state"`
}
