package call

import (
	"context"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/data"
)

type curl struct {
	args []string
}

func newCurlBackend(callData *data.Call) (*curl, error) {
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

		// defer os.Remove(tmpFile.Name())

		args = append(args, "-d", "@"+tmpFile.Name())
	}

	args = append(args, callData.Host.String())

	return &curl{args: args}, nil
}

func (curl curl) runAsCmd(ctx context.Context) ([]byte, error) {
	curlCmd := exec.CommandContext(ctx, "curl", curl.args...)
	return curlCmd.CombinedOutput()
}
