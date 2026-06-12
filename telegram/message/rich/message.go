package rich

import "github.com/gotd/td/tg"

// Message assembles a [tg.InputRichMessage] from blocks and attachments.
//
// The zero value is not usable; create one with [New]. Methods return the same
// pointer so they can be chained, and [Message.Input] produces the request
// value.
type Message struct {
	msg tg.InputRichMessage
}

// New creates a rich message from the given top-level blocks.
func New(blocks ...tg.PageBlockClass) *Message {
	return &Message{msg: tg.InputRichMessage{Blocks: blocks}}
}

// Block appends top-level blocks to the message.
func (m *Message) Block(blocks ...tg.PageBlockClass) *Message {
	m.msg.Blocks = append(m.msg.Blocks, blocks...)
	return m
}

// RTL marks the message as right-to-left.
func (m *Message) RTL() *Message {
	m.msg.Rtl = true
	return m
}

// NoAutoLink disables automatic detection of links, mentions and similar
// entities in the message text.
func (m *Message) NoAutoLink() *Message {
	m.msg.Noautolink = true
	return m
}

// Photos sets the photos referenced by the message blocks.
func (m *Message) Photos(photos ...tg.InputPhotoClass) *Message {
	m.msg.Photos = photos
	return m
}

// Documents sets the documents referenced by the message blocks.
func (m *Message) Documents(documents ...tg.InputDocumentClass) *Message {
	m.msg.Documents = documents
	return m
}

// Users sets the users referenced by the message blocks.
func (m *Message) Users(users ...tg.InputUserClass) *Message {
	m.msg.Users = users
	return m
}

// Input returns the assembled input rich message.
func (m *Message) Input() *tg.InputRichMessage {
	cp := m.msg
	return &cp
}

// Source describes an HTML or Markdown rich message source to be parsed by
// Telegram's servers, together with its attached files.
//
// The zero value is a valid empty source; configure it with the methods and
// finalize it with [Source.HTML] or [Source.Markdown].
type Source struct {
	rtl        bool
	noAutoLink bool
	files      []tg.InputRichFileClass
}

// Rich starts a server-parsed rich message source.
func Rich() *Source {
	return &Source{}
}

// RTL marks the message as right-to-left.
func (s *Source) RTL() *Source {
	s.rtl = true
	return s
}

// NoAutoLink disables automatic detection of links, mentions and similar
// entities in the message text.
func (s *Source) NoAutoLink() *Source {
	s.noAutoLink = true
	return s
}

// Photo attaches a photo referenced by the given id in the message source
// (inputRichFilePhoto).
func (s *Source) Photo(id string, photo tg.InputPhotoClass) *Source {
	s.files = append(s.files, &tg.InputRichFilePhoto{ID: id, Photo: photo})
	return s
}

// Document attaches a document referenced by the given id in the message source
// (inputRichFileDocument).
func (s *Source) Document(id string, document tg.InputDocumentClass) *Source {
	s.files = append(s.files, &tg.InputRichFileDocument{ID: id, Document: document})
	return s
}

// Files sets the files referenced by the message source.
func (s *Source) Files(files ...tg.InputRichFileClass) *Source {
	s.files = files
	return s
}

// HTML finalizes the source as an HTML rich message parsed by the server
// (inputRichMessageHTML).
func (s *Source) HTML(html string) *tg.InputRichMessageHTML {
	msg := &tg.InputRichMessageHTML{
		Rtl:        s.rtl,
		Noautolink: s.noAutoLink,
		HTML:       html,
	}
	if len(s.files) > 0 {
		msg.SetFiles(s.files)
	}
	return msg
}

// Markdown finalizes the source as a Markdown rich message parsed by the server
// (inputRichMessageMarkdown).
func (s *Source) Markdown(markdown string) *tg.InputRichMessageMarkdown {
	msg := &tg.InputRichMessageMarkdown{
		Rtl:        s.rtl,
		Noautolink: s.noAutoLink,
		Markdown:   markdown,
	}
	if len(s.files) > 0 {
		msg.SetFiles(s.files)
	}
	return msg
}

// HTML wraps an HTML source into an inputRichMessageHTML to be parsed by
// Telegram's servers. For attachments or flags, use [Rich] instead.
func HTML(html string) *tg.InputRichMessageHTML {
	return Rich().HTML(html)
}

// Markdown wraps a Markdown source into an inputRichMessageMarkdown to be parsed
// by Telegram's servers. For attachments or flags, use [Rich] instead.
func Markdown(markdown string) *tg.InputRichMessageMarkdown {
	return Rich().Markdown(markdown)
}
