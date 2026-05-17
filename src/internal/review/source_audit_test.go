package review

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// These tests scan the source tree to enforce assembled-system invariants
// that are easy to violate in a future patch without breaking any single
// unit's behaviour. They run alongside the regular suite so a regression
// fails CI rather than silently flipping a system-wide guarantee.

const srcRootRelative = "../../../src"

// walkGoSources visits every .go file under src/ that isn't a test file or a
// sqlc-generated file, calling fn with the path and the file's contents.
func walkGoSources(t *testing.T, fn func(path string, content string)) {
	t.Helper()
	root, err := filepath.Abs(srcRootRelative)
	if err != nil {
		t.Fatalf("resolve src root: %v", err)
	}
	if err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		// sqlc-generated code is regenerated from .sql files — scanning it
		// would conflate generated literals with hand-written ones.
		if strings.Contains(path, "/core/db/sqlc/") {
			return nil
		}
		// templ-generated _templ.go files mirror their .templ counterparts;
		// scan the .templ source separately to avoid double-counting.
		if strings.HasSuffix(path, "_templ.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fn(path, string(data))
		return nil
	}); err != nil {
		t.Fatalf("walk src: %v", err)
	}
}

// walkTemplSources visits every .templ file under src/.
func walkTemplSources(t *testing.T, fn func(path string, content string)) {
	t.Helper()
	root, err := filepath.Abs(srcRootRelative)
	if err != nil {
		t.Fatalf("resolve src root: %v", err)
	}
	if err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || !strings.HasSuffix(path, ".templ") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fn(path, string(data))
		return nil
	}); err != nil {
		t.Fatalf("walk templ: %v", err)
	}
}

// --- No live write of 'stalled' to album_rating_state.state ---
//
// The retirement migration narrows album_rating_state.state's CHECK to
// {provisional, finalized}. Any application code path that produces the
// literal 'stalled' for that column is a bug. This guards the invariant at
// the source level — the DB CHECK guards it at the data level.

func TestNoLiveCodeWritesStalledToRatingState(t *testing.T) {
	// Match the literal in either quote style, e.g. `state = "stalled"`,
	// `State: "stalled"`, `'stalled'`. We then filter false positives by
	// inspecting the surrounding context: the constant is allowed inside
	// RatingStateLogLabel (read-path label lookup for historical log entries)
	// and inside comments.
	literal := regexp.MustCompile(`["']stalled["']`)
	commentLine := regexp.MustCompile(`^\s*//`)

	violations := []string{}
	walkGoSources(t, func(path, content string) {
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if !literal.MatchString(line) {
				continue
			}
			if commentLine.MatchString(line) {
				continue
			}
			// Allowed: the historical-label read-path. Recognised by the case
			// arm in review.RatingStateLogLabel — the only function in the
			// codebase that needs to mention the historical value.
			if strings.Contains(line, `case "stalled":`) || strings.Contains(line, `case 'stalled':`) {
				continue
			}
			violations = append(violations, path+":"+itoa(i+1)+": "+strings.TrimSpace(line))
		}
	})

	if len(violations) > 0 {
		t.Fatalf("found live references to 'stalled' that may write the historical value to album_rating_state.state:\n  %s",
			strings.Join(violations, "\n  "))
	}
}

// --- ReratePromptFrag is fully retired ---
//
// The pre-rework codebase exposed ReratePromptFrag as an alternate first view
// for the rating modal. PC4 requires the modal always opens on the
// score-entry form; any surviving reference to ReratePromptFrag or its DOM
// id would be a foothold for a regression that re-introduces an alternate
// entry point.

func TestReratePromptFragHasNoLiveReferences(t *testing.T) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\bReratePromptFrag\b`),
		regexp.MustCompile(`data-testid="rerate-prompt"`),
	}
	violations := []string{}
	walkGoSources(t, func(path, content string) {
		for _, p := range patterns {
			if p.MatchString(content) {
				violations = append(violations, path+": "+p.String())
			}
		}
	})
	walkTemplSources(t, func(path, content string) {
		for _, p := range patterns {
			if p.MatchString(content) {
				violations = append(violations, path+": "+p.String())
			}
		}
	})
	if len(violations) > 0 {
		t.Fatalf("ReratePromptFrag must be fully retired:\n  %s", strings.Join(violations, "\n  "))
	}
}

// --- Carousel section DOM id is stable ---
//
// The dashboard carousel section's HTMX swaps target id="carousel-section".
// The post-rework rename of the Rerate Due tab to Provisional did not change
// that id; PC9 says HTMX swaps must continue to land on the same element.
// This guards both halves of the contract: the id is declared on the
// section, and every hx-target in the codebase that mentions
// carousel-section uses that exact id.

func TestCarouselSectionDOMIdIsStable(t *testing.T) {
	idDecl := regexp.MustCompile(`id="carousel-section"`)
	hxTarget := regexp.MustCompile(`hx-target="#carousel-section"`)
	const carouselTempl = "carousel_section_frag.templ"

	var sawIDDecl, sawHxTarget bool
	walkTemplSources(t, func(path, content string) {
		if strings.HasSuffix(path, carouselTempl) {
			if idDecl.MatchString(content) {
				sawIDDecl = true
			}
			if hxTarget.MatchString(content) {
				sawHxTarget = true
			}
		}
	})
	if !sawIDDecl {
		t.Fatalf("expected id=%q on the carousel section templ", "carousel-section")
	}
	if !sawHxTarget {
		t.Fatalf("expected hx-target=%q wiring inside the carousel section templ", "#carousel-section")
	}
}

// itoa is a tiny stdlib-free int-to-string for line numbers in error messages.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := make([]byte, 0, 6)
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
