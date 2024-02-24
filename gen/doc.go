// Package gen implements code generation from TL schema.
package gen

import (
	"net/url"
	"path"
	"strings"
	"unicode"
)

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	u2 := new(url.URL)
	*u2 = *u
	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}
	return u2
}

func (g *Generator) docURL(parts ...string) string {
	if g.docBase == nil {
		return ""
	}

	u := cloneURL(g.docBase)
	u.Path = path.Join(append([]string{u.Path}, parts...)...)

	return u.String()
}

func splitLine(s string, limit int) (r []string) {
	for {
		if len(s) < limit {
			r = append(r, s)
			return
		}

		idx := strings.LastIndexFunc(s[:limit], func(r rune) bool {
			return unicode.IsSpace(r) || r == '.' || r == ','
		})
		if idx < 0 || len(s)-1 == idx {
			r = append(r, s)
			return
		}

		r = append(r, s[:idx])
		s = s[idx+1:]
	}
}

func splitLines(s []string, limit int) []string {
	r := make([]string, 0, len(s))

	for _, line := range s {
		r = append(r, splitLine(line, limit)...)
	}

	return r
}
