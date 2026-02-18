/* auto-generated */

package crmevent

type CrmEventPayloadOrgCreated struct {
	Version   float32 `json:"version"`
	Name      string  `json:"name,omitempty"`
	AccountId string  `json:"accountId,omitempty"`
}

type CrmEventPayloadOrgUserCreated struct {
	Version   float32 `json:"version"`
	Org       string  `json:"org,omitempty"`
	AccountId string  `json:"accountId,omitempty"`
	Email     string  `json:"email,omitempty"`
}

type CrmEventPayloadOrgUserRemoved struct {
	Version   float32 `json:"version"`
	Org       string  `json:"org,omitempty"`
	AccountId string  `json:"accountId,omitempty"`
	Email     string  `json:"email,omitempty"`
}

type CrmEventPayloadAccountCreated struct {
	Version float32 `json:"version"`
	Name    string  `json:"name,omitempty"`
	Id      string  `json:"id,omitempty"`
}

type CrmEventPayloadAccountUserCreated struct {
	Version   float32 `json:"version"`
	AccountId string  `json:"accountId,omitempty"`
	Email     string  `json:"email,omitempty"`
}

type CrmEventPayloadAccountUserRemoved struct {
	Version   float32 `json:"version"`
	AccountId string  `json:"accountId,omitempty"`
	Email     string  `json:"email,omitempty"`
}

type CrmEventPayload struct {
	OrgCreated         *CrmEventPayloadOrgCreated         `json:"orgCreated,omitempty"`
	OrgUserCreated     *CrmEventPayloadOrgUserCreated     `json:"orgUserCreated,omitempty"`
	OrgUserRemoved     *CrmEventPayloadOrgUserRemoved     `json:"orgUserRemoved,omitempty"`
	AccountCreated     *CrmEventPayloadAccountCreated     `json:"accountCreated,omitempty"`
	AccountUserCreated *CrmEventPayloadAccountUserCreated `json:"accountUserCreated,omitempty"`
	AccountUserRemoved *CrmEventPayloadAccountUserRemoved `json:"accountUserRemoved,omitempty"`
}

type CrmEvent struct {
	Id        string           `json:"id,omitempty"`
	Created   string           `json:"created,omitempty"`
	Delivered string           `json:"delivered,omitempty"`
	Payload   *CrmEventPayload `json:"payload,omitempty"`
}
