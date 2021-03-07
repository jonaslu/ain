package call

import (
	"context"
	"time"
)

// !! TODO !! Make this a global config
const backendTimeoutSeconds = 10

func (callData *Data) SetBackend(backendName string) bool {
	switch backendName {
	case "curl":
		callData.BackendFunc = runAsCurl
		return true
	case "httpie":
		callData.BackendFunc = runAsHttpie
		return true
	default:
		return false
	}
}

func CallBackend(ctx context.Context, callData *Data) (string, error) {
	backendTimeoutContext, _ := context.WithTimeout(ctx, backendTimeoutSeconds*time.Second)
	return callData.BackendFunc(backendTimeoutContext, callData)
}
