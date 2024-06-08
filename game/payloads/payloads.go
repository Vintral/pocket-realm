package payloads

import "github.com/google/uuid"

type Payload struct {
	Type string `json:"type"`
}

type ExplorePayload struct {
	Type   string `json:"type"`
	Energy int    `json:"energy"`
}

type ExploreSuccess struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type GatherPayload struct {
	Type     string `json:"type"`
	Energy   int    `json:"energy"`
	Resource string `json:"resource"`
}

type RoundDataPayload struct {
	Type  string    `json:"type"`
	Round uuid.UUID `json:"round"`
}

type RoundDataSuccess struct {
	Type  string `json:"type"`
	Round string `json:"round"`
}

type Response struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

type Error struct {
	Type    string `default:"ERROR" json:"type"`
	Class   string `json:"class"`
	Message string `json:"message"`
}
