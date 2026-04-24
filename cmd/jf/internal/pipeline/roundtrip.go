package pipeline

import (
	"bytes"
	"fmt"
	"strings"
)

// CheckRoundtrip tests whether content survives md → ADF → md locally.
// Returns true if NormalizeMarkdown(original) == NormalizeMarkdown(roundtripped).
// Input should be frontmatter-stripped. Requires Node.js.
func CheckRoundtrip(content []byte) (bool, error) {
	adf, err := CompileMarkdown(content)
	if err != nil {
		return false, err
	}
	roundtripped, err := ConvertADF(adf)
	if err != nil {
		return false, err
	}
	return bytes.Equal(
		NormalizeMarkdown(content),
		NormalizeMarkdown(roundtripped),
	), nil
}

// FirstDivergence re-runs the roundtrip and returns a human-readable hint
// showing the first line where the original and roundtripped content differ.
// Returns empty string if roundtrip is clean or on error.
func FirstDivergence(content []byte) string {
	adf, err := CompileMarkdown(content)
	if err != nil {
		return fmt.Sprintf("compile error: %s", err)
	}
	roundtripped, err := ConvertADF(adf)
	if err != nil {
		return fmt.Sprintf("convert error: %s", err)
	}

	origLines := strings.Split(string(NormalizeMarkdown(content)), "\n")
	rtLines := strings.Split(string(NormalizeMarkdown(roundtripped)), "\n")

	maxLen := len(origLines)
	if len(rtLines) > maxLen {
		maxLen = len(rtLines)
	}

	for i := 0; i < maxLen; i++ {
		var orig, rt string
		if i < len(origLines) {
			orig = origLines[i]
		}
		if i < len(rtLines) {
			rt = rtLines[i]
		}
		if orig != rt {
			return fmt.Sprintf("line %d: %q vs %q", i+1, truncate(orig, 60), truncate(rt, 60))
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
