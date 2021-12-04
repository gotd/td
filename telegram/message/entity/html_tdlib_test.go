package entity

import "github.com/gotd/td/tg"

func tdlibHTMLTests() []htmlTestCase {
	entities := func(e ...tg.MessageEntityClass) func(msg string) []tg.MessageEntityClass {
		return func(msg string) []tg.MessageEntityClass {
			return e
		}
	}
	return []htmlTestCase{
		{"", "", nil, false, ""},
		{"➡️ ➡️", "➡️ ➡️", nil, false, ""},
		{
			"&lt;&gt;&amp;&quot;&laquo;&raquo;&#12345678;",
			"<>&\"&laquo;&raquo;&#12345678;",
			nil,
			false,
			"",
		},

		{
			"➡️ ➡️<i>➡️ ➡️</i>",
			"➡️ ➡️➡️ ➡️",
			entities(&tg.MessageEntityItalic{Offset: 5, Length: 5}),
			false,
			"",
		},
		{
			"➡️ ➡️<em>➡️ ➡️</em>", "➡️ ➡️➡️ ➡️",
			entities(&tg.MessageEntityItalic{Offset: 5, Length: 5}),
			false,
			"",
		},
		{
			"➡️ ➡️<b>➡️ ➡️</b>",
			"➡️ ➡️➡️ ➡️",
			entities(&tg.MessageEntityBold{Offset: 5, Length: 5}),
			false,
			"",
		},
		{
			"➡️ ➡️<strong>➡️ ➡️</strong>",
			"➡️ ➡️➡️ ➡️",
			entities(&tg.MessageEntityBold{Offset: 5, Length: 5}),
			false,
			"",
		},
		{
			"➡️ ➡️<u>➡️ ➡️</u>",
			"➡️ ➡️➡️ ➡️",
			entities(&tg.MessageEntityUnderline{Offset: 5, Length: 5}),
			false,
			"",
		},
		{
			"➡️ ➡️<ins>➡️ ➡️</ins>",
			"➡️ ➡️➡️ ➡️",
			entities(&tg.MessageEntityUnderline{Offset: 5, Length: 5}),
			false,
			"",
		},
		{
			"➡️ ➡️<s>➡️ ➡️</s>",
			"➡️ ➡️➡️ ➡️",
			entities(&tg.MessageEntityStrike{Offset: 5, Length: 5}),
			false,
			"",
		},
		{
			"➡️ ➡️<strike>➡️ ➡️</strike>",
			"➡️ ➡️➡️ ➡️",
			entities(&tg.MessageEntityStrike{Offset: 5, Length: 5}),
			false,
			"",
		},
		{
			"➡️ ➡️<del>➡️ ➡️</del>",
			"➡️ ➡️➡️ ➡️",
			entities(&tg.MessageEntityStrike{Offset: 5, Length: 5}),
			false,
			"",
		},
		{
			"➡️ ➡️<i>➡️ ➡️</i><b>➡️ ➡️</b>",
			"➡️ ➡️➡️ ➡️➡️ ➡️",
			entities(
				&tg.MessageEntityItalic{Offset: 5, Length: 5},
				&tg.MessageEntityBold{Offset: 10, Length: 5},
			),
			false,
			"",
		},

		{
			"🏟 🏟<i>🏟 &lt🏟</i>",
			"🏟 🏟🏟 <🏟",
			entities(&tg.MessageEntityItalic{Offset: 5, Length: 6}),
			false,
			"",
		},
		{
			"🏟 🏟<i>🏟 &gt;<b aba   =   caba>&lt🏟</b></i>",
			"🏟 🏟🏟 ><🏟",
			entities(
				&tg.MessageEntityItalic{Offset: 5, Length: 7},
				&tg.MessageEntityBold{Offset: 9, Length: 3},
			),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<i    aba  =  190azAz-.   >a</i>",
			"🏟 🏟<a",
			entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<i    aba  =  190azAz-.>a</i>",
			"🏟 🏟<a",
			entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<i    aba  =  \"&lt;&gt;&quot;\">a</i>",
			"🏟 🏟<a",
			entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<i    aba  =  '&lt;&gt;&quot;'>a</i>",
			"🏟 🏟<a",
			entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<i    aba  =  '&lt;&gt;&quot;'>a</>",
			"🏟 🏟<a",
			entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<i>🏟 🏟&lt;</>",
			"🏟 🏟<🏟 🏟<",
			entities(&tg.MessageEntityItalic{Offset: 6, Length: 6}),
			false,
			"",
		},

		{
			"🏟 🏟&lt;<i>a</    >",
			"🏟 🏟<a",
			entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<i>a</i   >",
			"🏟 🏟<a",
			entities(&tg.MessageEntityItalic{Offset: 6, Length: 1}),
			false,
			"",
		},
		// Empty entity.
		{
			"🏟 🏟&lt;<b></b>",
			"🏟 🏟<",
			nil,
			false,
			"",
		},
		// Space handling.
		{
			"<i>\t</i>",
			"\t",
			entities(&tg.MessageEntityItalic{Offset: 0, Length: 1}),
			false,
			"",
		},
		{
			"<i>\r</i>",
			"\r",
			entities(&tg.MessageEntityItalic{Offset: 0, Length: 1}),
			false,
			"",
		},
		{
			"<i>\n</i>",
			"\n",
			entities(&tg.MessageEntityItalic{Offset: 0, Length: 1}),
			false,
			"",
		},
		{
			"<a href=telegram.org>\t</a>",
			"\t",
			entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
			false,
			"",
		},
		{
			"<a href=telegram.org>\r</a>",
			"\r",
			entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
			false,
			"",
		},
		{
			"<a href=telegram.org>\n</a>",
			"\n",
			entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
			false,
			"",
		},
		{
			"<code><i><b> </b></i></code><i><b><code> </code></b></i>",
			"  ",
			entities(
				&tg.MessageEntityCode{Offset: 0, Length: 1},
				&tg.MessageEntityBold{Offset: 0, Length: 1},
				&tg.MessageEntityItalic{Offset: 0, Length: 1},
				&tg.MessageEntityCode{Offset: 1, Length: 1},
				&tg.MessageEntityBold{Offset: 1, Length: 1},
				&tg.MessageEntityItalic{Offset: 1, Length: 1}),
			false,
			"",
		},
		{
			"<i><b> </b> <code> </code></i>",
			"   ",
			entities(
				&tg.MessageEntityItalic{Offset: 0, Length: 3},
				&tg.MessageEntityBold{Offset: 0, Length: 1},
				&tg.MessageEntityCode{Offset: 2, Length: 1},
			),
			false,
			"",
		},
		{
			"<a href=telegram.org> </a>",
			" ",
			entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
			false,
			"",
		},
		{
			"<a href  =\"telegram.org\"   > </a>",
			" ",
			entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
			false,
			"",
		},
		{
			"<a   href=  'telegram.org'   > </a>",
			" ",
			entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/"}),
			false,
			"",
		},
		{
			"<a   href=  'telegram.org?&lt;'   > </a>",
			" ",
			entities(&tg.MessageEntityTextURL{Offset: 0, Length: 1, URL: "http://telegram.org/?<"}),
			false,
			"",
		},
		// URL handling
		{
			"<a>telegram.org </a>",
			"telegram.org ",
			nil,
			false,
			"",
		},
		{
			"<a>telegram.org</a>", "telegram.org",
			entities(&tg.MessageEntityTextURL{
				Offset: 0,
				Length: 12,
				URL:    "http://telegram.org/",
			}),
			false,
			"",
		},
		{
			"<a>https://telegram.org/asdsa?asdasdwe#12e3we</a>",
			"https://telegram.org/asdsa?asdasdwe#12e3we",
			entities(&tg.MessageEntityTextURL{
				Offset: 0,
				Length: 42,
				URL:    "https://telegram.org/asdsa?asdasdwe#12e3we",
			}),
			false,
			"",
		},
		// <pre> and <code> handling
		{
			"🏟 🏟&lt;<pre  >🏟 🏟&lt;</>",
			"🏟 🏟<🏟 🏟<",
			entities(&tg.MessageEntityPre{Offset: 6, Length: 6}),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<code >🏟 🏟&lt;</>",
			"🏟 🏟<🏟 🏟<",
			entities(&tg.MessageEntityCode{Offset: 6, Length: 6}),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<pre><code>🏟 🏟&lt;</code></>",
			"🏟 🏟<🏟 🏟<",
			entities(
				&tg.MessageEntityPre{Offset: 6, Length: 6},
				&tg.MessageEntityCode{Offset: 6, Length: 6},
			),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<pre><code class=\"language-\">🏟 🏟&lt;</code></>",
			"🏟 🏟<🏟 🏟<",
			entities(
				&tg.MessageEntityPre{Offset: 6, Length: 6},
				&tg.MessageEntityCode{Offset: 6, Length: 6},
			),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<pre><code class=\"language-fift\">🏟 🏟&lt;</></>",
			"🏟 🏟<🏟 🏟<",
			entities(&tg.MessageEntityPre{Offset: 6, Length: 6, Language: "fift"}),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<code class=\"language-fift\"><pre>🏟 🏟&lt;</></>",
			"🏟 🏟<🏟 🏟<",
			entities(&tg.MessageEntityPre{Offset: 6, Length: 6, Language: "fift"}),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<pre><code class=\"language-fift\">🏟 🏟&lt;</> </>",
			"🏟 🏟<🏟 🏟< ",
			entities(
				&tg.MessageEntityPre{Offset: 6, Length: 7},
				&tg.MessageEntityCode{Offset: 6, Length: 6},
			),
			false,
			"",
		},
		{
			"🏟 🏟&lt;<pre> <code class=\"language-fift\">🏟 🏟&lt;</></>",
			"🏟 🏟< 🏟 🏟<",
			entities(
				&tg.MessageEntityPre{Offset: 6, Length: 7},
				&tg.MessageEntityCode{Offset: 7, Length: 6},
			),
			false,
			"",
		},
	}
}
