package gen

import (
	"strings"

	"github.com/go-openapi/inflect"
)

func pascalWords(words []string) string {
	for i, w := range words {
		upper := strings.ToUpper(w)
		if _, ok := acronyms[upper]; ok {
			words[i] = upper
		} else {
			words[i] = rules.Capitalize(w)
		}
	}
	return strings.Join(words, "")
}

var (
	rules    = ruleset()
	acronyms = make(map[string]struct{})
)

func isSeparator(r rune) bool {
	return r == '_' || r == '-'
}

// pascal converts the given name into a PascalCase.
//
//	user_info 	=> UserInfo
//	full_name 	=> FullName
//	user_id   	=> UserID
//	full-admin	=> FullAdmin
//
func pascal(s string) string {
	words := strings.FieldsFunc(s, isSeparator)
	return pascalWords(words)
}

// camel converts the given name into a camelCase.
//
//	user_info  => userInfo
//	full_name  => fullName
//	user_id    => userID
//	full-admin => fullAdmin
//
func camel(s string) string {
	words := strings.FieldsFunc(s, isSeparator)
	if len(words) == 1 {
		return strings.ToLower(words[0])
	}
	return strings.ToLower(words[0]) + pascalWords(words[1:])
}

func ruleset() *inflect.Ruleset {
	rules := inflect.NewDefaultRuleset()
	// Add common initialisms from golint and more.
	for _, w := range []string{
		"ACL", "API", "ASCII", "AWS", "CPU", "CSS", "DNS", "EOF", "GB", "GUID",
		"HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "KB", "LHS", "MAC", "MB",
		"QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SQL", "SSH", "SSO", "TCP",
		"TLS", "TTL", "UDP", "UI", "UID", "URI", "URL", "UTF8", "UUID", "VM",
		"XML", "XMPP", "XSRF", "XSS", "SMS",
	} {
		acronyms[w] = struct{}{}
		rules.AddAcronym(w)
	}
	return rules
}
