/* auto-generated */

package accountevent

type AccountEventPayloadAccountCreated struct {
	Version float32 `json:"version"`
	Id      string  `json:"id"`
	Name    string  `json:"name"`
}

type AccountEventPayload struct {
	AccountCreated *AccountEventPayloadAccountCreated `json:"accountCreated,omitempty"`
}

type AccountEvent struct {
	Id        string              `json:"id,omitempty"`
	Created   string              `json:"created,omitempty"`
	Delivered string              `json:"delivered,omitempty"`
	Payload   AccountEventPayload `json:"payload,omitempty"`
}
