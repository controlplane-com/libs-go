package data_service

import (
	"strings"

	"github.com/controlplane-com/types-go/pkg/base"
)

func GetLinkMap(links []base.Link) map[string]string {
	m := make(map[string]string)
	if links != nil {
		for _, l := range links {
			m[l.Rel] = l.Href
		}
	}
	return m
}

func GetLinkedObjectName(rel string, links []base.Link) string {
	href := GetLinkByRel(rel, links)
	name := href
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return name
}

func GetLinkByRel(rel string, links []base.Link) string {
	if links == nil {
		return ""
	}
	for _, l := range links {
		if l.Rel != rel {
			continue
		}
		return l.Href
	}
	return ""
}
