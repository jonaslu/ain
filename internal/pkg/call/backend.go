package call

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// !! TODO !! Make this a global config
const backendTimeoutSeconds = 10

func CallBackend(ctx context.Context, callData *Data, backend string) (string, error) {
	backendTimeoutContext, _ := context.WithTimeout(ctx, backendTimeoutSeconds*time.Second)

	switch backend {
	case "curl":
		return callData.runAsCurl(backendTimeoutContext)
	case "httpie":
		return callData.runAsHttpie(backendTimeoutContext)
	}

	return "", errors.Errorf("Unknown backend: %s", backend)
}
