package gen

import (
	"net/url"
	"path"
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
