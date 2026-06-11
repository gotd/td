package rich

import "github.com/gotd/td/tg"

// Title returns a title block (pageBlockTitle).
func Title(text tg.RichTextClass) *tg.PageBlockTitle {
	return &tg.PageBlockTitle{Text: text}
}

// Subtitle returns a subtitle block (pageBlockSubtitle).
func Subtitle(text tg.RichTextClass) *tg.PageBlockSubtitle {
	return &tg.PageBlockSubtitle{Text: text}
}

// Header returns a header block (pageBlockHeader).
func Header(text tg.RichTextClass) *tg.PageBlockHeader {
	return &tg.PageBlockHeader{Text: text}
}

// Subheader returns a subheader block (pageBlockSubheader).
func Subheader(text tg.RichTextClass) *tg.PageBlockSubheader {
	return &tg.PageBlockSubheader{Text: text}
}

// Kicker returns a kicker block (pageBlockKicker).
func Kicker(text tg.RichTextClass) *tg.PageBlockKicker {
	return &tg.PageBlockKicker{Text: text}
}

// Footer returns a footer block (pageBlockFooter).
func Footer(text tg.RichTextClass) *tg.PageBlockFooter {
	return &tg.PageBlockFooter{Text: text}
}

// Paragraph returns a paragraph block (pageBlockParagraph).
func Paragraph(text tg.RichTextClass) *tg.PageBlockParagraph {
	return &tg.PageBlockParagraph{Text: text}
}

// AuthorDate returns an author and publication date block (pageBlockAuthorDate).
func AuthorDate(author tg.RichTextClass, publishedDate int) *tg.PageBlockAuthorDate {
	return &tg.PageBlockAuthorDate{Author: author, PublishedDate: publishedDate}
}

// Heading1 returns a level-1 heading block (pageBlockHeading1).
func Heading1(text tg.RichTextClass) *tg.PageBlockHeading1 {
	return &tg.PageBlockHeading1{Text: text}
}

// Heading2 returns a level-2 heading block (pageBlockHeading2).
func Heading2(text tg.RichTextClass) *tg.PageBlockHeading2 {
	return &tg.PageBlockHeading2{Text: text}
}

// Heading3 returns a level-3 heading block (pageBlockHeading3).
func Heading3(text tg.RichTextClass) *tg.PageBlockHeading3 {
	return &tg.PageBlockHeading3{Text: text}
}

// Heading4 returns a level-4 heading block (pageBlockHeading4).
func Heading4(text tg.RichTextClass) *tg.PageBlockHeading4 {
	return &tg.PageBlockHeading4{Text: text}
}

// Heading5 returns a level-5 heading block (pageBlockHeading5).
func Heading5(text tg.RichTextClass) *tg.PageBlockHeading5 {
	return &tg.PageBlockHeading5{Text: text}
}

// Heading6 returns a level-6 heading block (pageBlockHeading6).
func Heading6(text tg.RichTextClass) *tg.PageBlockHeading6 {
	return &tg.PageBlockHeading6{Text: text}
}

// Heading returns a heading block for the given level (1-6), clamped to that
// range.
func Heading(level int, text tg.RichTextClass) tg.PageBlockClass {
	switch {
	case level <= 1:
		return Heading1(text)
	case level == 2:
		return Heading2(text)
	case level == 3:
		return Heading3(text)
	case level == 4:
		return Heading4(text)
	case level == 5:
		return Heading5(text)
	default:
		return Heading6(text)
	}
}

// Preformatted returns a preformatted (code) block with optional language
// (pageBlockPreformatted).
func Preformatted(text tg.RichTextClass, language string) *tg.PageBlockPreformatted {
	return &tg.PageBlockPreformatted{Text: text, Language: language}
}

// Divider returns a divider block (pageBlockDivider).
func Divider() *tg.PageBlockDivider {
	return &tg.PageBlockDivider{}
}

// AnchorBlock returns an anchor block with the given name (pageBlockAnchor),
// usable as the target of an [AnchorLink].
func AnchorBlock(name string) *tg.PageBlockAnchor {
	return &tg.PageBlockAnchor{Name: name}
}

// MathBlock returns a block-level mathematical expression with the given LaTeX
// source (pageBlockMath).
func MathBlock(source string) *tg.PageBlockMath {
	return &tg.PageBlockMath{Source: source}
}

// Thinking returns a thinking block (pageBlockThinking), used to show an AI
// model's reasoning.
func Thinking(text tg.RichTextClass) *tg.PageBlockThinking {
	return &tg.PageBlockThinking{Text: text}
}

// Blockquote returns a block quotation with an optional caption
// (pageBlockBlockquote). Use [BlockquoteBlocks] to quote multiple blocks.
func Blockquote(text, caption tg.RichTextClass) *tg.PageBlockBlockquote {
	return &tg.PageBlockBlockquote{Text: text, Caption: caption}
}

// BlockquoteBlocks returns a block quotation of multiple blocks with an
// optional caption (pageBlockBlockquoteBlocks).
func BlockquoteBlocks(caption tg.RichTextClass, blocks ...tg.PageBlockClass) *tg.PageBlockBlockquoteBlocks {
	return &tg.PageBlockBlockquoteBlocks{Blocks: blocks, Caption: caption}
}

// Pullquote returns a pull quotation with an optional caption
// (pageBlockPullquote).
func Pullquote(text, caption tg.RichTextClass) *tg.PageBlockPullquote {
	return &tg.PageBlockPullquote{Text: text, Caption: caption}
}

// Details returns a collapsible details block (pageBlockDetails). When open is
// true the block is expanded by default.
func Details(open bool, title tg.RichTextClass, blocks ...tg.PageBlockClass) *tg.PageBlockDetails {
	return &tg.PageBlockDetails{Open: open, Title: title, Blocks: blocks}
}

// List returns an unordered list block (pageBlockList).
func List(items ...tg.PageListItemClass) *tg.PageBlockList {
	return &tg.PageBlockList{Items: items}
}

// OrderedList returns an ordered list block (pageBlockOrderedList).
func OrderedList(items ...tg.PageListOrderedItemClass) *tg.PageBlockOrderedList {
	return &tg.PageBlockOrderedList{Items: items}
}

// Collage returns a collage of blocks with an optional caption
// (pageBlockCollage).
func Collage(caption tg.PageCaption, items ...tg.PageBlockClass) *tg.PageBlockCollage {
	return &tg.PageBlockCollage{Items: items, Caption: caption}
}

// Slideshow returns a slideshow of blocks with an optional caption
// (pageBlockSlideshow).
func Slideshow(caption tg.PageCaption, items ...tg.PageBlockClass) *tg.PageBlockSlideshow {
	return &tg.PageBlockSlideshow{Items: items, Caption: caption}
}

// Cover wraps a block as a cover (pageBlockCover).
func Cover(cover tg.PageBlockClass) *tg.PageBlockCover {
	return &tg.PageBlockCover{Cover: cover}
}

// Photo returns a photo block referencing a photo by ID with an optional
// caption (pageBlockPhoto).
func Photo(photoID int64, caption tg.PageCaption) *tg.PageBlockPhoto {
	return &tg.PageBlockPhoto{PhotoID: photoID, Caption: caption}
}

// Video returns a video block referencing a document by ID with an optional
// caption (pageBlockVideo).
func Video(videoID int64, caption tg.PageCaption) *tg.PageBlockVideo {
	return &tg.PageBlockVideo{VideoID: videoID, Caption: caption}
}

// Audio returns an audio block referencing a document by ID with an optional
// caption (pageBlockAudio).
func Audio(audioID int64, caption tg.PageCaption) *tg.PageBlockAudio {
	return &tg.PageBlockAudio{AudioID: audioID, Caption: caption}
}

// Map returns a map block centered on the given input location
// (inputPageBlockMap), with the given zoom level, size in pixels and an
// optional caption.
func Map(geo tg.InputGeoPointClass, zoom, w, h int, caption tg.PageCaption) *tg.InputPageBlockMap {
	return &tg.InputPageBlockMap{Geo: geo, Zoom: zoom, W: w, H: h, Caption: caption}
}

// Table returns a table block with the given title (may be [Empty]) and rows
// (pageBlockTable). Set Bordered and Striped on the result to control styling.
func Table(title tg.RichTextClass, rows ...tg.PageTableRow) *tg.PageBlockTable {
	return &tg.PageBlockTable{Title: title, Rows: rows}
}

// Row returns a table row of the given cells (pageTableRow).
func Row(cells ...tg.PageTableCell) tg.PageTableRow {
	return tg.PageTableRow{Cells: cells}
}

// Cell returns a table cell with the given text (pageTableCell). Set Header,
// alignment, Colspan and Rowspan on the result as needed.
func Cell(text tg.RichTextClass) tg.PageTableCell {
	return tg.PageTableCell{Text: text}
}

// HeaderCell returns a header table cell with the given text (pageTableCell).
func HeaderCell(text tg.RichTextClass) tg.PageTableCell {
	c := Cell(text)
	c.Header = true
	return c
}

// RelatedArticles returns a related-articles block (pageBlockRelatedArticles).
func RelatedArticles(title tg.RichTextClass, articles ...tg.PageRelatedArticle) *tg.PageBlockRelatedArticles {
	return &tg.PageBlockRelatedArticles{Title: title, Articles: articles}
}

// Caption returns a page caption with optional credit (pageCaption). Pass
// [Empty] for an absent text or credit.
func Caption(text, credit tg.RichTextClass) tg.PageCaption {
	return tg.PageCaption{Text: text, Credit: credit}
}

// ListItem returns an unordered list item holding inline text
// (pageListItemText).
func ListItem(text tg.RichTextClass) *tg.PageListItemText {
	return &tg.PageListItemText{Text: text}
}

// CheckListItem returns an unordered task-list item holding inline text, with a
// checkbox in the given checked state (pageListItemText).
func CheckListItem(checked bool, text tg.RichTextClass) *tg.PageListItemText {
	return &tg.PageListItemText{Checkbox: true, Checked: checked, Text: text}
}

// ListItemBlocks returns an unordered list item holding blocks
// (pageListItemBlocks).
func ListItemBlocks(blocks ...tg.PageBlockClass) *tg.PageListItemBlocks {
	return &tg.PageListItemBlocks{Blocks: blocks}
}

// OrderedListItem returns an ordered list item holding inline text, labelled
// with num (pageListOrderedItemText).
func OrderedListItem(num string, text tg.RichTextClass) *tg.PageListOrderedItemText {
	return &tg.PageListOrderedItemText{Num: num, Text: text}
}

// OrderedListItemBlocks returns an ordered list item holding blocks, labelled
// with num (pageListOrderedItemBlocks).
func OrderedListItemBlocks(num string, blocks ...tg.PageBlockClass) *tg.PageListOrderedItemBlocks {
	return &tg.PageListOrderedItemBlocks{Num: num, Blocks: blocks}
}
