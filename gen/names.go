package gen

import (
	"strings"
	"unicode"

	"github.com/go-openapi/inflect"
)

func pascalWords(words []string) string {
	for i, w := range words {
		upper := strings.ToUpper(w)
		if alias, ok := aliases[upper]; ok {
			words[i] = alias
			continue
		}

		if _, ok := acronyms[upper]; ok {
			words[i] = upper
			continue
		}

		// check for acronym + any letter: IDs, IPs, UIDs
		if _, ok := acronyms[upper[:len(upper)-1]]; ok {
			words[i] = upper[:len(upper)-1] + strings.ToLower(upper[len(upper)-1:])
			continue
		}

		words[i] = rules.Capitalize(w)
	}

	return strings.Join(words, "")
}

var (
	rules    = ruleset()
	acronyms = make(map[string]struct{})
	aliases  = make(map[string]string)
)

// CodeMD5Test -> cODEmd5tEST

// splitByWords split name into words by separators and capital letters
func splitByWords(s string) []string {
	if s == "" {
		return []string{s}
	}

	if strings.ContainsAny(s, "_-") {
		return strings.FieldsFunc(s, func(r rune) bool {
			return r == '_' || r == '-'
		})
	}

	words := make([]string, 0)
	word := make([]rune, 0, len(s))
	var prev rune

	for i, current := range s {
		if i == 0 {
			prev = current
			continue
		}

		switch {
		case unicode.IsNumber(prev) || unicode.IsNumber(current):
			word = append(word, prev)
		case unicode.IsUpper(prev) && unicode.IsLower(current): // Ab
			if len(word) != 0 {
				words = append(words, string(word))
				word = word[:0]
			}

			word = append(word, prev)
		case unicode.IsLower(prev) && unicode.IsUpper(current): // aB
			words = append(words, string(append(word, prev)))
			word = word[:0]
		default:
			word = append(word, prev)
		}

		prev = current
	}

	words = append(words, string(append(word, prev)))

	return words
}

// pascal converts the given name into a PascalCase.
//
//	user_info 	 => UserInfo
//	full_name 	 => FullName
//	user_id   	 => UserID
//	full-admin	 => FullAdmin
//	cdnConfig    => CDNConfig
//	cdn_1_config => CDN1Config
func pascal(s string) string {
	words := splitByWords(s)
	return pascalWords(words)
}

// camel converts the given name into a camelCase.
//
//	user_info  => userInfo
//	full_name  => fullName
//	user_id    => userID
//	full-admin => fullAdmin
func camel(s string) string {
	words := splitByWords(s)
	if len(words) == 1 {
		return strings.ToLower(words[0])
	}
	return strings.ToLower(words[0]) + pascalWords(words[1:])
}

func ruleset() *inflect.Ruleset {
	r := inflect.NewDefaultRuleset()
	// Add common initialisms from golint and more.
	//
	// You can commit this with following message:
	//   chore(gen): update acronym list
	for _, w := range []string{
		"ACL", "API", "ASCII", "AWS", "CPU", "CSS", "DNS", "EOF", "GB", "GUID",
		"HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "KB", "LHS", "MAC", "MB",
		"QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SQL", "SSH", "SSO", "TCP",
		"TLS", "TTL", "UDP", "UI", "UID", "URI", "URL", "UTF8", "UUID", "VM",
		"XML", "XMPP", "XSRF", "XSS", "SMS", "CDN", "TCP", "UDP", "DC", "PFS",
		"P2P", "SHA256", "SHA1", "MD5", "SRP", "2FA", "ISO",
	} {
		acronyms[w] = struct{}{}
		r.AddAcronym(w)
	}

	aliases = map[string]string{
		"TCPO":    "TCPObfuscated",
		"SMSJOBS": "SMSJobs",
	}

	return r
}
