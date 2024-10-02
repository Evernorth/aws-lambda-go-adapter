package util

import "os"

func IsLambdaRuntime() bool {
	return os.Getenv("LAMBDA_TASK_ROOT") != ""
}
