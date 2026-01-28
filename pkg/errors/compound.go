package errors

import (
	"fmt"
	"github.com/controlplane-com/libs-go/pkg/pipeline"
	"strings"
)

type CompoundError struct {
	summary string
	Errs    []error
}

func (c *CompoundError) Error() string {
	s, _ := pipeline.Map(c.Errs, func(e error) (string, error) {
		return e.Error(), nil
	})
	return fmt.Sprintf("%s: - %s", c.summary, strings.Join(s, "\n-"))
}

func NewCompoundError(summary string, errs ...error) error {
	return &CompoundError{summary: summary, Errs: errs}
}
