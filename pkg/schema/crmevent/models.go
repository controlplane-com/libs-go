/* auto-generated */

package crmevent

type CrmEventPayloadOrgCreated struct {
	Version   float64 `json:"version"`
	Name      string  `json:"name,omitzero" validate:"required"`
	AccountId string  `json:"accountId,omitzero" validate:"required"`
}

type CrmEventPayloadOrgUserCreated struct {
	Version   float64 `json:"version"`
	Org       string  `json:"org,omitzero" validate:"required"`
	AccountId string  `json:"accountId,omitzero" validate:"required"`
	Email     string  `json:"email,omitzero" validate:"required"`
}

type CrmEventPayloadOrgUserRemoved struct {
	Version   float64 `json:"version"`
	Org       string  `json:"org,omitzero" validate:"required"`
	AccountId string  `json:"accountId,omitzero" validate:"required"`
	Email     string  `json:"email,omitzero" validate:"required"`
}

type CrmEventPayloadAccountCreated struct {
	Version float64 `json:"version"`
	Name    string  `json:"name,omitzero" validate:"required"`
	Id      string  `json:"id,omitzero" validate:"required"`
}

type CrmEventPayloadAccountUserCreated struct {
	Version   float64 `json:"version"`
	AccountId string  `json:"accountId,omitzero" validate:"required"`
	Email     string  `json:"email,omitzero" validate:"required"`
}

type CrmEventPayloadAccountUserRemoved struct {
	Version   float64 `json:"version"`
	AccountId string  `json:"accountId,omitzero" validate:"required"`
	Email     string  `json:"email,omitzero" validate:"required"`
}

type CrmEventPayload struct {
	OrgCreated         *CrmEventPayloadOrgCreated         `json:"orgCreated,omitzero"`
	OrgUserCreated     *CrmEventPayloadOrgUserCreated     `json:"orgUserCreated,omitzero"`
	OrgUserRemoved     *CrmEventPayloadOrgUserRemoved     `json:"orgUserRemoved,omitzero"`
	AccountCreated     *CrmEventPayloadAccountCreated     `json:"accountCreated,omitzero"`
	AccountUserCreated *CrmEventPayloadAccountUserCreated `json:"accountUserCreated,omitzero"`
	AccountUserRemoved *CrmEventPayloadAccountUserRemoved `json:"accountUserRemoved,omitzero"`
}

type CrmEvent struct {
	Id        string           `json:"id,omitzero"`
	Created   string           `json:"created,omitzero"`
	Delivered string           `json:"delivered,omitzero" validate:"date"`
	Payload   *CrmEventPayload `json:"payload,omitzero"`
}
