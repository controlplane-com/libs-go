package delivery

// statusRanks orders statuses from best to worst for Compose.
var statusRanks = map[string]int{
	StatusDelivered:         0,
	StatusPending:           1,
	StatusInProgress:        2,
	StatusFailed:            3,
	StatusPermanentlyFailed: 4,
}

// statusRank ranks an unrecognized status worst so it can never mask a failure.
func statusRank(status string) int {
	if r, ok := statusRanks[status]; ok {
		return r
	}
	return len(statusRanks)
}

// Compose returns the effective state of a delivery that handed off to downstream
// deliveries: the upstream's own state until it is delivered, then the worst downstream state.
func Compose(upstream State, downstream ...State) State {
	if upstream.Status != StatusDelivered || len(downstream) == 0 {
		return upstream
	}
	out := upstream
	out.ErrorMessages = append([]ErrorEntry(nil), upstream.ErrorMessages...)
	worst := downstream[0]
	for _, d := range downstream {
		if statusRank(d.Status) > statusRank(worst.Status) {
			worst = d
		}
		out.ErrorMessages = append(out.ErrorMessages, d.ErrorMessages...)
		if d.LastAttemptAt != nil && (out.LastAttemptAt == nil || d.LastAttemptAt.After(*out.LastAttemptAt)) {
			out.LastAttemptAt = d.LastAttemptAt
		}
	}
	out.Status = worst.Status
	out.LastErrorType = worst.LastErrorType
	out.NextRetryAt = worst.NextRetryAt
	out.DeliveredAt = nil
	if worst.Status != StatusDelivered {
		return out
	}
	for _, d := range downstream {
		if d.DeliveredAt != nil && (out.DeliveredAt == nil || d.DeliveredAt.After(*out.DeliveredAt)) {
			out.DeliveredAt = d.DeliveredAt
		}
	}
	return out
}
