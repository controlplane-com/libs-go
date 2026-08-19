package metering

import (
	"fmt"
	"net/http"
	"time"

	"github.com/controlplane-com/libs-go/pkg/web"
)

// InjectConsumptionRequest is the payload for the metering usage-injection
// endpoint (POST /org/{org}/-inject). It mirrors the server-side type in
// metering/pkg/types. StartTime must be hour-aligned and must not be in the
// future; Tags must be non-nil.
type InjectConsumptionRequest struct {
	ChargeableItem string         `json:"chargeableItem"`
	Value          float64        `json:"value"`
	StartTime      time.Time      `json:"startTime"`
	Tags           map[string]any `json:"tags"`
}

// InjectConsumptionResponse is the metering response for an injection.
type InjectConsumptionResponse struct {
	Success       bool    `json:"success"`
	Delta         float64 `json:"delta"`
	PreviousValue float64 `json:"previousValue"`
	NewValue      float64 `json:"newValue"`
}

// InjectConsumption injects a single hourly consumption datapoint for the given
// org. The metering service rolls the value up into its day/week/month
// aggregates. Authorization uses a Bearer token (typically the controller
// token) that is authorized for the injectConsumption action.
func (m *httpClient) InjectConsumption(token string, org string, req *InjectConsumptionRequest) (*InjectConsumptionResponse, error) {
	return web.DoJSONRequestWithBodyAndResult[*InjectConsumptionRequest, *InjectConsumptionResponse](
		m.httpClient,
		http.MethodPost,
		m.getFullUrl(fmt.Sprintf("/org/%s/-inject", org)),
		req,
		token,
	)
}
