package base

import (
	"strings"
)

func (l Links) Org() string {
	href := l.FindHref("org")
	return LastPart(href)
}

func (l Links) Gvc() string {
	href := l.FindHref("gvc")
	return LastPart(href)
}

func (l Links) FindHref(rel string) string {
	var matches []Link
	for _, link := range l {
		if link.Rel == rel {
			matches = append(matches, link)
		}
	}
	if len(matches) == 0 {
		return ""
	}
	return matches[0].Href
}

func LastPart(href string) string {
	i := strings.LastIndex(href, "/")
	if i < 0 {
		return href
	}
	if i == len(href)-1 {
		return ""
	}
	return href[i+1:]
}

func GetPart(href string, part int) string {
	if len(href) == 0 {
		return ""
	}
	p := strings.Split(href, "/")
	index := part
	if href[0] == '/' {
		index++
	}
	if index < len(p) {
		return p[index]
	}
	return ""
}
