package workloadOptions

import "strings"

// For returns the item for a location name; stored items are keyed by the org link /org/<org>/location/<name>.
func (o LocalOptions) For(location string) *LocalOptionsItem {
	if location == "" {
		return nil
	}
	for i := range o {
		link := string(o[i].Location)
		if link == location || strings.HasSuffix(link, "/location/"+location) {
			return &o[i]
		}
	}
	return nil
}
