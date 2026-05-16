package templates

import "testing"

func TestIconStyleSuffix(t *testing.T) {
	t.Run("returns empty string for outline style", func(t *testing.T) {
		got := iconStyleSuffix(IconStyleOutline)
		if got != "" {
			t.Errorf("iconStyleSuffix(IconStyleOutline) = %q; want %q", got, "")
		}
	})

	t.Run("returns -fill suffix for fill style", func(t *testing.T) {
		got := iconStyleSuffix(IconStyleFill)
		if got != "-fill" {
			t.Errorf("iconStyleSuffix(IconStyleFill) = %q; want %q", got, "-fill")
		}
	})
}
