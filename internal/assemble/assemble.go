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

	if merge.Config.Timeout != -1 {
		dest.Config.Timeout = merge.Config.Timeout
	}
}

func getCallData(parse *data.Parse) (*data.Call, []string) {
	callData := data.Call{}
	fatals := []string{}

	if len(parse.Host) == 0 {
		fatals = append(fatals, "No mandatory [Host] section found")
	} else {
		hostStr := strings.Join(parse.Host, "")
		host, err := url.Parse(hostStr)
		if err != nil {
			fatals = append(fatals, fmt.Sprintf("[Host] has illegal url: %s, error: %v", hostStr, err))
		}

		callData.Host = host
	}

	if parse.Backend == "" {
		fatals = append(fatals, "No mandatory [Backend] section found")
	}

	if len(fatals) != 0 {
		return nil, fatals
	}

	callData.Body = parse.Body
	callData.Method = parse.Method
	callData.Headers = parse.Headers
	callData.Backend = parse.Backend
	callData.BackendOptions = parse.BackendOptions
	callData.Config = parse.Config

	return &callData, fatals
}

func appendErrorMessages(errorMessage, filename string, fatals []string) string {
	if errorMessage != "" {
		errorMessage = errorMessage + "\n"
	}

	if filename != "" {
		errorMessage = errorMessage + `Error in file: ` + filename + "\n"
	}

	return errorMessage + strings.Join(fatals, "\n") + "\n"
}

func Assemble(ctx context.Context, filenames []string, execute bool) (*data.Call, string, error) {
	errors := ""

	parseData := &data.Parse{}
	parseData.Config.Timeout = -1

	for _, filename := range filenames {
		template, err := disk.ReadTemplate(filename, execute)
		if err != nil {
			return nil, "", err
		}

		fileCallData, fatals := parse.ParseTemplate(ctx, template)
		if len(fatals) > 0 {
			errors = appendErrorMessages(errors, filename, fatals)
		}

		if errors == "" {
			mergeCallData(parseData, fileCallData)
		}
	}

	if errors != "" {
		return nil, errors, nil
	}

	callData, validationErrors := getCallData(parseData)
	if len(validationErrors) > 0 {
		errors = appendErrorMessages(errors, "", validationErrors)
	}

	return callData, errors, nil
}
