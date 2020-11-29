package sections

import "github.com/jonaslu/ain/template"

type Error struct {
	Message      string
	TemplateLine template.SourceMarker
}

func newError(message string, templateLine template.SourceMarker) *Error {
	return &Error{Message: message, TemplateLine: templateLine}
}

type Warning struct {
	Message      string
	TemplateLine template.SourceMarker
}

func newWarning(message string, templateLine template.SourceMarker) Warning {
	return Warning{Message: message, TemplateLine: templateLine}
}

type Warnings []Warning

func addWarning(warnings Warnings, message string, templateLine template.SourceMarker) Warnings {
	return append(warnings, newWarning(message, templateLine))
}
