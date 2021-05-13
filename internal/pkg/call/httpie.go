package call

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
)

type httpie struct {
	args        []string
	tmpFileName string
}

func newHttpieBackend(data *data.Call) (*httpie, error) {
	returnValue := &httpie{}

	var args []string
	var foundIgnoreStdin bool
	for _, backendOptions := range data.BackendOptions {
		for _, backendOption := range backendOptions {
			if backendOption == "--ignore-stdin" {
				foundIgnoreStdin = true
			}

			args = append(args, backendOption)
		}
	}

	if !foundIgnoreStdin {
		args = append([]string{"--ignore-stdin"}, args...)
	}

	if data.Method != "" {
		args = append(args, strings.ToUpper(data.Method))
	}

	args = append(args, data.Host.String())

	for _, header := range data.Headers {
		args = append(args, header)
	}

	if len(data.Body) > 0 {
		tmpFile, err := data.GetBodyAsTempFile()
		if err != nil {
			return nil, err
		}

		returnValue.tmpFileName = tmpFile.Name()
		args = append(args, "@"+tmpFile.Name())
	}

	returnValue.args = args

	return returnValue, nil
}

func (httpie httpie) runAsCmd(ctx context.Context) ([]byte, error) {
	httpCmd := exec.CommandContext(ctx, "http", httpie.args...)
	return httpCmd.CombinedOutput()
}

func (httpie httpie) getAsString() (string, error) {
	return "httpie", nil
}

func (httpie httpie) cleanUp() error {
	if httpie.tmpFileName != "" {
		return os.Remove(httpie.tmpFileName)
	}

	return nil
}
