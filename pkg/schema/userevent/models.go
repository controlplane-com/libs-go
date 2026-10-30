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
	UserCreated *UserEventPayloadUserCreated `json:"userCreated,omitempty"`
	UserRemoved *UserEventPayloadUserRemoved `json:"userRemoved,omitempty"`
}

type UserEvent struct {
	Id        string           `json:"id,omitempty"`
	Created   string           `json:"created,omitempty"`
	Delivered string           `json:"delivered,omitempty"`
	Payload   UserEventPayload `json:"payload,omitempty"`
}
