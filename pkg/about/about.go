package about

import (
	"encoding/json"
	"net/http"
)

// Build information stamped at link time. Each service's Makefile/Dockerfile passes
// `-ldflags "-X github.com/controlplane-com/libs-go/pkg/about.<Var>=..."`; the
// -X flag works for any package that is part of the build, so one shared package
// serves every binary.
var (
	// Version is the current version of the app, generated at build time
	Version   = "dev"
	Epoch     = "-1"
	Timestamp = "dev"
	Build     = "dev"
)

type Ab struct {
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Epoch     string `json:"epoch"`
	Build     string `json:"build"`
}

var About Ab

func init() {
	About = Ab{
		Epoch:     Epoch,
		Timestamp: Timestamp,
		Version:   Version,
		Build:     Build,
	}
}

// ServeAbout writes the build information as JSON. Usable directly as an http.HandlerFunc.
func ServeAbout(writer http.ResponseWriter, _ *http.Request) {
	aboutBytes, _ := json.Marshal(About)
	_, _ = writer.Write(aboutBytes)
}
