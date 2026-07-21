package util

import (
	"os"

	"github.com/hekmon/processpriority"
)

func CurrentProcessLowerPriority() {
	pid := os.Getpid()
	err := processpriority.Set(pid, newPriority)
	if err != nil {
		panic(err)
	}
}
