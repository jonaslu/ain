package assemble

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/disk"
	"github.com/jonaslu/ain/internal/pkg/parse"
)

func mergeParsedTemplate(dest, merge *data.ParsedTemplate) {
	dest.Host = append(dest.Host, merge.Host...)

	dest.Query = append(dest.Query, merge.Query...)

	if len(merge.Body) != 0 {
		dest.Body = merge.Body
	}

	if merge.Method != "" {
		dest.Method = merge.Method
	}

	dest.Headers = append(dest.Headers, merge.Headers...)

	if merge.Backend != "" {
		dest.Backend = merge.Backend
	}

	dest.BackendOptions = append(dest.BackendOptions, merge.BackendOptions...)

	if merge.Config.Timeout != data.TimeoutNotSet {
		dest.Config.Timeout = merge.Config.Timeout
	}

	if merge.Config.QueryDelim != nil {
		dest.Config.QueryDelim = merge.Config.QueryDelim
	}
}

func newBackendInput(parsedTemplate *data.ParsedTemplate) (*data.BackendInput, []string) {
	backendInput := data.BackendInput{}
	fatals := []string{}

	if len(parsedTemplate.Host) == 0 {
		fatals = append(fatals, "No mandatory [Host] section found")
	} else {
		hostStr := strings.Join(parsedTemplate.Host, "")
		host, err := url.Parse(hostStr)

		if err != nil {
			fatals = append(fatals, fmt.Sprintf("[Host] has illegal url: %s, error: %v", hostStr, err))
		} else {
			addQueryString(host, parsedTemplate)
			backendInput.Host = host
		}
	}

	if parsedTemplate.Backend == "" {
		fatals = append(fatals, "No mandatory [Backend] section found")
	}

	backendInput.Body = parsedTemplate.Body
	backendInput.Method = parsedTemplate.Method
	backendInput.Headers = parsedTemplate.Headers
	backendInput.Backend = parsedTemplate.Backend
	backendInput.BackendOptions = parsedTemplate.BackendOptions
	backendInput.Config = parsedTemplate.Config

	return &backendInput, fatals
}

func appendFatalMessages(fatalMessage, filename string, fatals []string) string {
	if fatalMessage != "" {
		fatalMessage = fatalMessage + "\n\n"
	}

	if filename != "" {
		fatalMessage = fatalMessage + "Fatal error"
		if len(fatals) > 1 {
			fatalMessage = fatalMessage + "s"
		}

		fatalMessage = fatalMessage + " in file: " + filename + "\n"
	}

	return fatalMessage + strings.Join(fatals, "\n")
}

func Assemble(ctx context.Context, filenames []string) (*data.BackendInput, string, error) {
	fatals := ""

	parsedTemplate := &data.ParsedTemplate{}
	parsedTemplate.Config.Timeout = data.TimeoutNotSet

	for _, filename := range filenames {
		rawTemplateString, err := disk.ReadRawTemplateString(filename)
		if err != nil {
			return nil, "", err
		}

		template, fileFatals := parse.ParseTemplate(ctx, rawTemplateString)
		if len(fileFatals) > 0 {
			fatals = appendFatalMessages(fatals, filename, fileFatals)
		}

		if fatals == "" {
			mergeParsedTemplate(parsedTemplate, template)
		}
	}

	if fatals != "" {
		return nil, fatals, nil
	}

	backendInput, validationFatals := newBackendInput(parsedTemplate)
	if len(validationFatals) > 0 {
		fatals = appendFatalMessages(fatals, "", validationFatals)
	}

	return backendInput, fatals, nil
}
