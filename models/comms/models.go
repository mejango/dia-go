package comms

import coremodel "github.com/mejango/dia-go/models/core"

// CommsState retains encrypted provider input for explicit delivery retries.
type CommsState struct {
	Delivery    string        `json:"delivery"`
	Subject     coremodel.Ref `json:"subject"`
	RequestHash string        `json:"requestHash"`
}
