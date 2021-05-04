package call

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
)

type curl struct {
	args        []string
	tmpFileName string
}

func newCurlBackend(callData *data.Call) (*curl, error) {
	returnValue := &curl{}
	args := callData.BackendOptions

	if callData.Method != "" {
		args = append(args, "-X", strings.ToUpper(callData.Method))
	}

	for _, header := range callData.Headers {
		args = append(args, "-H", header)
	}

	if len(callData.Body) > 0 {
		tmpFile, err := callData.GetBodyAsTempFile()
		if err != nil {
			return nil, err
		}

		returnValue.tmpFileName = tmpFile.Name()
		args = append(args, "-d", "@"+tmpFile.Name())
	}

	args = append(args, callData.Host.String())

	returnValue.args = args

	return returnValue, nil
}

func (curl curl) runAsCmd(ctx context.Context) ([]byte, error) {
	curlCmd := exec.CommandContext(ctx, "curl", curl.args...)
	return curlCmd.CombinedOutput()
}

func (curl curl) cleanUp() error {
	if curl.tmpFileName != "" {
		return os.Remove(curl.tmpFileName)
	}

	return nil
}
