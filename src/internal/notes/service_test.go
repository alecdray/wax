package notes

import (
	"strings"
	"testing"
)

// --- RenderMarkdown ---

func TestRenderMarkdown(t *testing.T) {
	t.Run("renders a bare URL as a clickable link", func(t *testing.T) {
		out := string(RenderMarkdown("https://example.com"))
		if !strings.Contains(out, `href="https://example.com"`) {
			t.Fatalf("expected href link, got: %s", out)
		}
	})

	t.Run("renders inline markdown link", func(t *testing.T) {
		out := string(RenderMarkdown("[Pitchfork](https://pitchfork.com)"))
		if !strings.Contains(out, `href="https://pitchfork.com"`) {
			t.Fatalf("expected href link, got: %s", out)
		}
		if !strings.Contains(out, "Pitchfork") {
			t.Fatalf("expected link text, got: %s", out)
		}
	})

	t.Run("links open in new tab with noreferrer", func(t *testing.T) {
		out := string(RenderMarkdown("[Pitchfork](https://pitchfork.com)"))
		if !strings.Contains(out, `target="_blank"`) {
			t.Fatalf("expected target=_blank, got: %s", out)
		}
		if !strings.Contains(out, `rel="noopener noreferrer"`) {
			t.Fatalf("expected rel=noopener noreferrer, got: %s", out)
		}
	})

	t.Run("renders plain text unchanged", func(t *testing.T) {
		out := string(RenderMarkdown("just some notes"))
		if !strings.Contains(out, "just some notes") {
			t.Fatalf("expected plain text in output, got: %s", out)
		}
	})

	t.Run("renders empty string as empty", func(t *testing.T) {
		out := strings.TrimSpace(string(RenderMarkdown("")))
		if out != "" {
			t.Fatalf("expected empty output, got: %q", out)
		}
	})

	t.Run("does not execute HTML in content", func(t *testing.T) {
		out := string(RenderMarkdown("<script>alert(1)</script>"))
		if strings.Contains(out, "<script>") {
			t.Fatalf("expected script to be escaped, got: %s", out)
		}
	})
}
