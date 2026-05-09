package notes

import (
	"bytes"
	"html/template"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

const MaxSleeveNoteLength = 10_000

type AlbumNoteDTO struct {
	ID        string
	UserID    string
	AlbumID   string
	Content   string
	UpdatedAt time.Time
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.Linkify),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

// RenderMarkdown converts markdown content to safe HTML.
// All links are given target="_blank" rel="noopener noreferrer".
func RenderMarkdown(content string) template.HTML {
	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(content))
	}
	out := strings.ReplaceAll(buf.String(), "<a ", `<a target="_blank" rel="noopener noreferrer" `)
	return template.HTML(out)
}
