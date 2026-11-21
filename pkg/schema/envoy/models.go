/* auto-generated */

package envoy

import "github.com/controlplane-com/libs-go/pkg/schema/envoyAccessLog"
import "github.com/controlplane-com/libs-go/pkg/schema/envoyCluster"
import "github.com/controlplane-com/libs-go/pkg/schema/envoyExcExtAuth"
import "github.com/controlplane-com/libs-go/pkg/schema/envoyHttp"
import "github.com/controlplane-com/libs-go/pkg/schema/volumeSpec"

type EnvoyFilters struct {
	AccessLog            []envoyAccessLog.AccessLog          `json:"accessLog,omitzero"`
	Clusters             []envoyCluster.Cluster              `json:"clusters,omitzero"`
	ExcludedExternalAuth []envoyExcExtAuth.ExcExtAuth        `json:"excludedExternalAuth,omitzero"`
	ExcludedRateLimit    []envoyExcExtAuth.ExcludedRateLimit `json:"excludedRateLimit,omitzero"`
	Http                 []envoyHttp.HttpFilter              `json:"http,omitzero"`
	Network              []any                               `json:"network,omitzero"`
	Volumes              []volumeSpec.VolumeSpec             `json:"volumes,omitzero"`
}
