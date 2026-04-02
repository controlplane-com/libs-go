package images

import (
	"errors"
	"fmt"
	"strings"
)

type TranslationContext struct {
	Org               string
	Project           string
	Image             string
	EnvironmentDomain string
}

func TranslateImage(context *TranslationContext) (string, error) {
	if strings.HasPrefix(context.Image, fmt.Sprintf("gcr.io/%s", context.Project)) {
		return "", errors.New("Invalid image path")
	}

	if !IsCplnImage(context.Image) {
		return context.Image, nil
	}
	_, a, _ := strings.Cut(context.Image, "/image/")

	return fmt.Sprintf("%s.registry.%s/%s", context.Org, context.EnvironmentDomain, a), nil
}

func IsCplnImage(image string) bool {
	return strings.HasPrefix(image, "/org") || strings.HasPrefix(image, "/image/")
}
