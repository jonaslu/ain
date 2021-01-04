package sections

import (
	"strings"
	"testing"

	"github.com/jonaslu/ain/template"
)

func TestCaptureSeveralOfSameSection(t *testing.T) {
	sectionName := "Host"

	templateStr := `
[` + sectionName + `]
[` + sectionName + `]`

	templat := template.TokenizeTemplate(templateStr)

	_, err := captureSection(sectionName, templat)

	if err == nil {
		t.Errorf("Expected no warnings")
	}

	if !strings.Contains(err.Message, "Several ["+sectionName+"] sections found") {
		t.Error("Expected warning on several [" + sectionName + "]")
	}
}
