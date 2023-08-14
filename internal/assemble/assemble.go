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

func mergeCallData(dest, merge *data.Parse) {
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

func newBackendInput(parse *data.Parse) (*data.BackendInput, []string) {
	backendInput := data.BackendInput{}
	fatals := []string{}

	if len(parse.Host) == 0 {
		fatals = append(fatals, "No mandatory [Host] section found")
	} else {
		hostStr := strings.Join(parse.Host, "")
		host, err := url.Parse(hostStr)

		if err != nil {
			fatals = append(fatals, fmt.Sprintf("[Host] has illegal url: %s, error: %v", hostStr, err))
		} else {
			addQueryString(host, parse)
			backendInput.Host = host
		}
	}

	if parse.Backend == "" {
		fatals = append(fatals, "No mandatory [Backend] section found")
	}

	backendInput.Body = parse.Body
	backendInput.Method = parse.Method
	backendInput.Headers = parse.Headers
	backendInput.Backend = parse.Backend
	backendInput.BackendOptions = parse.BackendOptions
	backendInput.Config = parse.Config

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

	parseData := &data.Parse{}
	parseData.Config.Timeout = data.TimeoutNotSet

	for _, filename := range filenames {
		rawTemplateString, err := disk.ReadRawTemplateString(filename)
		if err != nil {
			return nil, "", err
		}

		fileCallData, fileFatals := parse.ParseTemplate(ctx, rawTemplateString)
		if len(fileFatals) > 0 {
			fatals = appendFatalMessages(fatals, filename, fileFatals)
		}

		if fatals == "" {
			mergeCallData(parseData, fileCallData)
		}
	}

	if fatals != "" {
		return nil, fatals, nil
	}

	backendInput, validationFatals := newBackendInput(parseData)
	if len(validationFatals) > 0 {
		fatals = appendFatalMessages(fatals, "", validationFatals)
	}

	return backendInput, fatals, nil
}
