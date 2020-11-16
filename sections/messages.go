package sections

import "github.com/jonaslu/ain/template"

type Warning struct {
	Message      string
	TemplateLine template.SourceMarker
}

func newWarning(message string, templateLine template.SourceMarker) Warning {
	return Warning{Message: message, TemplateLine: templateLine}
}

type Warnings []Warning

func (parseResult *ParseResult) addWarning(message string, templateLine template.SourceMarker) {
	parseResult.warnings = append(parseResult.warnings, newWarning(message, templateLine))
}

type Error struct {
	Message      string
	TemplateLine template.SourceMarker
}

func newError(message string, templateLine template.SourceMarker) Error {
	return Error{Message: message, TemplateLine: templateLine}
}

type Errors []Error

func (parseResult *ParseResult) addError(message string, templateLine template.SourceMarker) {
	parseResult.errors = append(parseResult.errors, newError(message, templateLine))
}

type ParseResult struct {
	warnings Warnings
	errors   Errors
}
