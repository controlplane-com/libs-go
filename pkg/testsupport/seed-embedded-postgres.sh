#!/bin/sh
#
# Cache the embedded-postgres binaries a service's tests need, at image build time.
#
# Run from inside the service's module, with the version alias its tests use:
#
#     /home/nonroot/go-libs/testsupport/seed-embedded-postgres.sh V14
#
# The concrete version is read from the embedded-postgres release the service actually depends on
# rather than written down here, because the constants move between releases -- V14 is 14.13.0 in
# v1.30.0 and 14.18.0 in v1.32.0. A hardcoded version would still build, cache the wrong file, and
# quietly put the download back into every test run.
set -e

if [ -z "$1" ]; then
    echo "usage: seed-embedded-postgres.sh <version alias, e.g. V14>" >&2
    exit 1
fi

dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

cat > "${tmp}/version.go" <<GO
package main

import (
	"fmt"

	epg "github.com/fergusstrange/embedded-postgres"
)

func main() { fmt.Print(epg.$1) }
GO

# Resolved through the current module, so this fails the build if the alias is not one the linked
# release defines.
version=$(go run "${tmp}/version.go")

go run "${dir}/cmd/seed-embedded-postgres/main.go" "${version}"
