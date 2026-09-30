package metering

import (
	"time"

	timeUtils "github.com/controlplane-com/libs-go/pkg/time-utils"
)

// InvoiceRetentionMonths is how far back billing can still generate an invoice. Metering keeps every
// granularity at least this long, so an invoice it accepts is never priced from expired rows.
const InvoiceRetentionMonths = 12

// RetentionCutoff is the earliest instant still retained when data is kept for the given number of whole months.
func RetentionCutoff(now time.Time, months int) time.Time {
	return timeUtils.AddMonths(timeUtils.FirstDayOfTheMonth(now.UTC()), -months)
}

func InvoiceRetentionCutoff(now time.Time) time.Time {
	return RetentionCutoff(now, InvoiceRetentionMonths)
}
