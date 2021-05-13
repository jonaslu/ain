package call

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
)

type curl struct {
	callData    *data.Call
	tmpFileName string
}

func newCurlBackend(callData *data.Call) (*curl, error) {
	returnValue := &curl{callData: callData}
	return returnValue, nil
}

func (curl *curl) getHeaderArguments() [][]string {
	args := [][]string{}
	for _, header := range curl.callData.Headers {
		args = append(args, []string{"-H", header})
	}

	return args
}

func (curl *curl) getMethodArgument() []string {
	if curl.callData.Method != "" {
		methodCapitalized := strings.ToUpper(curl.callData.Method)
		return []string{"-X", methodCapitalized}
	}

	return []string{}
}

func (curl *curl) getBodyArgument() ([]string, error) {
	if len(curl.callData.Body) > 0 {
		tmpFile, err := curl.callData.GetBodyAsTempFile()

		if err != nil {
			return nil, err
		}

		curl.tmpFileName = tmpFile.Name()
		return []string{"-d", "@" + tmpFile.Name()}, nil
	}

	return []string{}, nil
}

func (curl *curl) runAsCmd(ctx context.Context) ([]byte, error) {
	args := []string{}
	for _, backendOpt := range curl.callData.BackendOptions {
		args = append(args, backendOpt...)
	}

	args = append(args, curl.getMethodArgument()...)
	for _, headerArgs := range curl.getHeaderArguments() {
		args = append(args, headerArgs...)
	}

	bodyArgs, err := curl.getBodyArgument()
	if err != nil {
		return nil, err
	}

	args = append(args, bodyArgs...)
	args = append(args, curl.callData.Host.String())

	curlCmd := exec.CommandContext(ctx, "curl", args...)

	return curlCmd.CombinedOutput()
}

func (curl *curl) getAsString() (string, error) {
	return "curl", nil
}

func (curl *curl) cleanUp() error {
	if curl.tmpFileName != "" {
		return os.Remove(curl.tmpFileName)
	}

	return nil
}
