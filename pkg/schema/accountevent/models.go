/* auto-generated */

package accountevent

type AccountEventPayloadAccountCreated struct {
	Version float32 `json:"version"`
	Id      string  `json:"id"`
	Name    string  `json:"name"`
}

type AccountEventPayload struct {
	AccountCreated *AccountEventPayloadAccountCreated `json:"accountCreated,omitzero"`
}

type AccountEvent struct {
	Id        string              `json:"id,omitzero"`
	Created   string              `json:"created,omitzero"`
	Delivered string              `json:"delivered,omitzero"`
	Payload   AccountEventPayload `json:"payload"`
}
