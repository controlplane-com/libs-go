/*
Command seed-embedded-postgres fills the embedded-postgres binary cache at image build time.

github.com/fergusstrange/embedded-postgres downloads a postgres build from Maven Central the
first time a test starts a server, and caches it under $HOME/.embedded-postgres-go. CI containers
start with that cache empty, so every test run fetched the jar again, and any non-200 from the
repository failed the suite — reported as "no version found matching <version>", which the library
returns for any bad status, not just a missing artifact. Seeding the cache while the image is built
means the tests themselves never reach the network.

Run it with the postgres versions a service's tests ask for:

	go run .../seed-embedded-postgres 14.13.0

Only the standard library is used, so it runs with `go run <file>` from anywhere, with no module
context and nothing to download first.
*/
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const repository = "https://repo1.maven.org/maven2"

// Retries cover the transient failures that made this worth doing at all. Cheap: on the happy path
// the first attempt succeeds.
const (
	attempts = 5
	backoff  = 3 * time.Second
)

func main() {
	versions := os.Args[1:]
	if len(versions) == 0 {
		fatal("usage: seed-embedded-postgres <version>...")
	}
	cacheDir, err := cacheDirectory()
	if err != nil {
		fatal("cannot determine the cache directory: %v", err)
	}
	for _, version := range versions {
		if err := seed(cacheDir, version); err != nil {
			fatal("%v", err)
		}
	}
}

// seed writes one version's binaries into the cache, under the exact name the library's
// CacheLocator looks for. A version already present is left alone.
func seed(cacheDir, version string) error {
	goos, arch := platform()
	name := fmt.Sprintf("embedded-postgres-binaries-%s-%s-%s.txz", goos, arch, version)
	target := filepath.Join(cacheDir, name)
	if _, err := os.Stat(target); err == nil {
		fmt.Printf("seed-embedded-postgres: %s already cached\n", name)
		return nil
	}

	url := fmt.Sprintf("%s/io/zonky/test/postgres/embedded-postgres-binaries-%s-%s/%s/embedded-postgres-binaries-%s-%s-%s.jar",
		repository, goos, arch, version, goos, arch, version)

	jar, err := download(url)
	if err != nil {
		return err
	}
	binaries, err := extractTarXz(jar, url)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return fmt.Errorf("cannot create %s: %w", cacheDir, err)
	}
	// Write beside the target and rename, so a build interrupted mid-write cannot leave a
	// truncated archive that every later run would happily treat as a cache hit.
	tmp, err := os.CreateTemp(cacheDir, name+".partial-*")
	if err != nil {
		return fmt.Errorf("cannot create a temporary file in %s: %w", cacheDir, err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(binaries); err != nil {
		tmp.Close()
		return fmt.Errorf("cannot write %s: %w", tmp.Name(), err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cannot close %s: %w", tmp.Name(), err)
	}
	if err := os.Rename(tmp.Name(), target); err != nil {
		return fmt.Errorf("cannot move the archive into place: %w", err)
	}
	fmt.Printf("seed-embedded-postgres: cached %s (%d bytes)\n", name, len(binaries))
	return nil
}

func download(url string) ([]byte, error) {
	var last error
	for attempt := 1; attempt <= attempts; attempt++ {
		body, err := get(url)
		if err == nil {
			return body, nil
		}
		last = err
		if attempt < attempts {
			fmt.Fprintf(os.Stderr, "seed-embedded-postgres: attempt %d/%d failed: %v\n", attempt, attempts, err)
			time.Sleep(time.Duration(attempt) * backoff)
		}
	}
	return nil, fmt.Errorf("giving up on %s after %d attempts: %w", url, attempts, last)
}

func get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// extractTarXz pulls the binaries out of the jar. The jar is a zip holding one .txz, which is what
// the library caches — it stores the inner archive, not the jar.
func extractTarXz(jar []byte, url string) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(jar), int64(len(jar)))
	if err != nil {
		return nil, fmt.Errorf("cannot read the archive from %s: %w", url, err)
	}
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || !strings.HasSuffix(file.Name, ".txz") {
			continue
		}
		opened, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("cannot open %s in the archive: %w", file.Name, err)
		}
		defer opened.Close()
		return io.ReadAll(opened)
	}
	return nil, fmt.Errorf("no .txz found in the archive retrieved from %s", url)
}

// cacheDirectory mirrors the library's default: $HOME/.embedded-postgres-go, falling back to a
// relative path when there is no home directory. Anything else and the tests would look somewhere
// this never wrote to.
func cacheDirectory() (string, error) {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".embedded-postgres-go"), nil
	}
	return ".embedded-postgres-go", nil
}

// platform mirrors the library's naming for the zonky binaries, which differs from Go's for arm.
func platform() (string, string) {
	goos, arch := runtime.GOOS, runtime.GOARCH
	if goos == "linux" && arch == "arm64" {
		arch += "v8"
	}
	return goos, arch
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "seed-embedded-postgres: "+format+"\n", args...)
	os.Exit(1)
}
