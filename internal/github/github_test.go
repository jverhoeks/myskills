package github

import (
	"testing"
)

func TestHasGH(t *testing.T) {
	// Just verify it doesn't panic
	_ = HasGH()
}
