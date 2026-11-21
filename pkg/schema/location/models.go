/* hand-maintained — canonical source of truth for the `location` resource schema.
 * `validate:` tags drive pkg/validate.FromType. See nodelibs/schema/src/location.ts.
 * Location = Base.keys({origin, provider, region, spec, status}). The createSchema
 * fork (name+provider required, description default-to-name, spec.enabled default)
 * is applied on POST by locationCreateRefine (schema_reflect_orggvc.go).
 */

package location

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// LocationTags aliases base.Tags for the reflection base-type registry.
type LocationTags = base.Tags

// IpRange is a CIDR-required leaf (String().Custom(cidrRequired)); see validate.go.
type IpRange string

// IpRangeList is status.ipRanges: an IpRange slice deduped in place (the old
// `custom=dedupeEmpty` tag). Its ValidationRule lives in validate.go.
type IpRangeList []IpRange

type LocationOrigin string

const (
	LocationOriginBuiltin LocationOrigin = "builtin"
	LocationOriginDefault LocationOrigin = "default"
	LocationOriginCustom  LocationOrigin = "custom"
)

type LocationProvider string

const (
	LocationProviderAws     LocationProvider = "aws"
	LocationProviderGcp     LocationProvider = "gcp"
	LocationProviderAzure   LocationProvider = "azure"
	LocationProviderByok    LocationProvider = "byok"
	LocationProviderLinode  LocationProvider = "linode"
	LocationProviderVultr   LocationProvider = "vultr"
	LocationProviderEquinix LocationProvider = "equinix"
	LocationProviderOci     LocationProvider = "oci"
	LocationProviderCpln    LocationProvider = "cpln"
)

type Location struct {
	Id           string           `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name        `json:"name,omitzero"`
	Kind         base.Kind        `json:"kind,omitzero"`
	Version      *float64         `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string           `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         LocationTags     `json:"tags,omitzero" validate:"default={}"`
	Created      string           `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string           `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links       `json:"links,omitzero" validate:"readOnly"`
	Origin       LocationOrigin   `json:"origin,omitzero" validate:"valid=builtin|default|custom,default=default,optional,readOnly"`
	Provider     LocationProvider `json:"provider,omitzero" validate:"valid=aws|gcp|azure|byok|linode|vultr|equinix|oci|cpln,unmodifiable"`
	Region       string           `json:"region,omitzero" validate:"unmodifiable"`
	Spec         *LocationSpec    `json:"spec,omitzero"`
	Status       *LocationStatus  `json:"status,omitzero"`
}

type LocationSpec struct {
	Enabled *bool `json:"enabled,omitzero"`
}

type LocationStatusGeo struct {
	Lat       *float64 `json:"lat,omitzero"`
	Lon       *float64 `json:"lon,omitzero"`
	Country   string   `json:"country,omitzero"`
	State     string   `json:"state,omitzero"`
	City      string   `json:"city,omitzero"`
	Continent string   `json:"continent,omitzero"`
}

type LocationStatus struct {
	Geo      *LocationStatusGeo `json:"geo,omitzero"`
	IpRanges IpRangeList        `json:"ipRanges,omitzero"`
}
