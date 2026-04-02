package data_service

import (
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
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
	return GetLastLinkPart(href)
}

func GetLastLinkPart(link string) string {
	if i := strings.LastIndex(link, "/"); i >= 0 {
		link = link[i+1:]
	}
	return link
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
