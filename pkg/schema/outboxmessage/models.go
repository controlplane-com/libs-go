/* auto-generated */

package outboxmessage

type OutboxMessage struct {
	Id        string `json:"id,omitzero"`
	Created   string `json:"created,omitzero"`
	Delivered string `json:"delivered,omitzero" validate:"date"`
	Payload   string `json:"payload,omitzero"`
}
