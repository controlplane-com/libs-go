/* auto-generated */

package userevent

type UserEventPayloadUserCreated struct {
	Version   float32 `json:"version"`
	AccountId string  `json:"accountId"`
	Email     string  `json:"email"`
}

type UserEventPayloadUserRemoved struct {
	Version   float32 `json:"version"`
	AccountId string  `json:"accountId"`
	Email     string  `json:"email"`
}

type UserEventPayload struct {
	UserCreated *UserEventPayloadUserCreated `json:"userCreated,omitzero"`
	UserRemoved *UserEventPayloadUserRemoved `json:"userRemoved,omitzero"`
}

type UserEvent struct {
	Id        string           `json:"id,omitzero"`
	Created   string           `json:"created,omitzero"`
	Delivered string           `json:"delivered,omitzero"`
	Payload   UserEventPayload `json:"payload"`
}
