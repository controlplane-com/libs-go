package delivery

import (
	"testing"
	"time"
)

func ptrTime(t time.Time) *time.Time { return &t }

func ptrString(s string) *string { return &s }

func TestCompose_UndeliveredUpstreamWins(t *testing.T) {
	up := State{Status: StatusFailed, LastErrorType: ptrString("transient")}
	got := Compose(up, State{Status: StatusDelivered})
	if got.Status != StatusFailed || *got.LastErrorType != "transient" {
		t.Fatalf("got %+v, want the upstream state", got)
	}
}

func TestCompose_NoDownstreamIsUpstream(t *testing.T) {
	up := State{Status: StatusDelivered, DeliveredAt: ptrTime(time.Unix(10, 0))}
	got := Compose(up)
	if got.Status != StatusDelivered || !got.DeliveredAt.Equal(time.Unix(10, 0)) {
		t.Fatalf("got %+v", got)
	}
}

func TestCompose_WorstDownstreamWins(t *testing.T) {
	cases := []struct {
		name       string
		downstream []string
		want       string
	}{
		{"all delivered", []string{StatusDelivered, StatusDelivered}, StatusDelivered},
		{"pending beats delivered", []string{StatusDelivered, StatusPending}, StatusPending},
		{"in progress beats pending", []string{StatusPending, StatusInProgress}, StatusInProgress},
		{"failed beats in progress", []string{StatusInProgress, StatusFailed}, StatusFailed},
		{"permanently failed beats all", []string{StatusFailed, StatusPermanentlyFailed, StatusDelivered}, StatusPermanentlyFailed},
		{"unknown status is never masked", []string{StatusDelivered, "bogus"}, "bogus"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ds []State
			for _, s := range tc.downstream {
				ds = append(ds, State{Status: s})
			}
			got := Compose(State{Status: StatusDelivered}, ds...)
			if got.Status != tc.want {
				t.Fatalf("status = %q, want %q", got.Status, tc.want)
			}
		})
	}
}

func TestCompose_DeliveredAtIsLatestDownstreamOnlyWhenAllDelivered(t *testing.T) {
	up := State{Status: StatusDelivered, DeliveredAt: ptrTime(time.Unix(1, 0))}
	got := Compose(up,
		State{Status: StatusDelivered, DeliveredAt: ptrTime(time.Unix(5, 0))},
		State{Status: StatusDelivered, DeliveredAt: ptrTime(time.Unix(9, 0))},
	)
	if got.DeliveredAt == nil || !got.DeliveredAt.Equal(time.Unix(9, 0)) {
		t.Fatalf("deliveredAt = %v, want the latest downstream", got.DeliveredAt)
	}

	got = Compose(up,
		State{Status: StatusDelivered, DeliveredAt: ptrTime(time.Unix(5, 0))},
		State{Status: StatusFailed},
	)
	if got.DeliveredAt != nil {
		t.Fatalf("deliveredAt = %v, want nil while a downstream is undelivered", got.DeliveredAt)
	}
}

func TestCompose_MergesErrorHistoryAndTakesWorstError(t *testing.T) {
	up := State{Status: StatusDelivered, ErrorMessages: []ErrorEntry{{Message: "up"}}}
	down := State{
		Status:        StatusPermanentlyFailed,
		LastErrorType: ptrString("permanent"),
		ErrorMessages: []ErrorEntry{{Message: "down", Detail: "sendgrid returned 400"}},
	}
	got := Compose(up, down)
	if len(got.ErrorMessages) != 2 || got.ErrorMessages[1].Detail != "sendgrid returned 400" {
		t.Fatalf("errorMessages = %+v", got.ErrorMessages)
	}
	if got.LastErrorType == nil || *got.LastErrorType != "permanent" {
		t.Fatalf("lastErrorType = %v", got.LastErrorType)
	}
	if len(up.ErrorMessages) != 1 {
		t.Fatal("Compose must not mutate the upstream's error history")
	}
}
