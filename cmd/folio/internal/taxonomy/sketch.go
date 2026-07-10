package taxonomy

import (
	_ "embed"
	"strings"
)

//go:embed sketch.html
var sketchTemplateHTML string

// SketchTemplate returns the seed sketch HTML with {{TITLE}} filled from topic.
func SketchTemplate(topic string) string {
	title := strings.ReplaceAll(topic, "-", " ")
	title = strings.Title(title) //nolint:staticcheck
	return strings.ReplaceAll(sketchTemplateHTML, "{{TITLE}}", title)
}
