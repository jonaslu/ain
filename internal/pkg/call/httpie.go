package call

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
)

type httpie struct {
	callData    *data.Call
	tmpFileName string
}

func prependIgnoreStdin(callData *data.Call) {
	var foundIgnoreStdin bool

	for _, backendOptionLine := range callData.BackendOptions {
		for _, backendOption := range backendOptionLine {
			if backendOption == "--ignore-stdin" {
				foundIgnoreStdin = true
				break
			}
		}
	}

	if !foundIgnoreStdin {
		callData.BackendOptions = append([][]string{{"--ignore-stdin"}}, callData.BackendOptions...)
	}
}

func newHttpieBackend(callData *data.Call) (*httpie, error) {
	prependIgnoreStdin(callData)
	return &httpie{callData: callData}, nil
}

func (httpie *httpie) getMethodArgument() string {
	return strings.ToUpper(httpie.callData.Method)
}

func (httpie *httpie) getBodyArgument() (string, error) {
	tmpFile, err := httpie.callData.GetBodyAsTempFile("")
	if err != nil {
		return "", err
	}

	httpie.tmpFileName = tmpFile.Name()
	return "@" + tmpFile.Name(), nil
}

func (httpie *httpie) runAsCmd(ctx context.Context) ([]byte, error) {
	args := []string{}
	for _, backendOpt := range httpie.callData.BackendOptions {
		args = append(args, backendOpt...)
	}

	if httpie.callData.Method != "" {
		args = append(args, httpie.getMethodArgument())
	}

	args = append(args, httpie.callData.Host.String())
	args = append(args, httpie.callData.Headers...)

	if len(httpie.callData.Body) > 0 {
		bodyArg, err := httpie.getBodyArgument()
		if err != nil {
			return nil, err
		}

		args = append(args, bodyArg)
	}

	httpCmd := exec.CommandContext(ctx, "http", args...)
	return httpCmd.CombinedOutput()
}

func (httpie *httpie) getAsString() (string, error) {
	return "httpie", nil
}

func (httpie *httpie) cleanUp() error {
	if httpie.tmpFileName != "" {
		return os.Remove(httpie.tmpFileName)
	}

	return nil
}
