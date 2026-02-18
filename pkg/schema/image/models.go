/* auto-generated */

package image

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type ImageTags map[string]any

type ImageManifest map[string]any

type Image struct {
	Id           string        `json:"id,omitempty"`
	Name         string        `json:"name,omitempty"`
	Kind         base.Kind     `json:"kind,omitempty"`
	Version      float32       `json:"version"`
	Created      string        `json:"created,omitempty"`
	LastModified string        `json:"lastModified,omitempty"`
	Tags         ImageTags     `json:"tags,omitempty"`
	Links        base.Links    `json:"links,omitempty"`
	Repository   string        `json:"repository,omitempty"`
	Tag          string        `json:"tag,omitempty"`
	Manifest     ImageManifest `json:"manifest,omitempty"`
	Digest       string        `json:"digest,omitempty"`
}

type ImageSummaryRepositories struct {
	Name      string  `json:"name,omitempty"`
	TagsCount float32 `json:"tagsCount"`
}

type ImageSummary struct {
	Kind         base.Kind                  `json:"kind,omitempty"`
	Links        base.Links                 `json:"links,omitempty"`
	Repositories []ImageSummaryRepositories `json:"repositories,omitempty"`
}
