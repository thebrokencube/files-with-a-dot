package pipeline

import "bytes"

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
