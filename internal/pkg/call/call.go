package call

import (
	"context"
	"time"
)

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
	backendTimeoutContext := ctx
	if callData.Config.Timeout > -1 {
		backendTimeoutContext, _ = context.WithTimeout(ctx, time.Duration(callData.Config.Timeout)*time.Second)
	}

	return callData.BackendFunc(backendTimeoutContext, callData)
}
