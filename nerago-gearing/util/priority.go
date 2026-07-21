package util

import (
	"os"

	"github.com/hekmon/processpriority"
)

func CurrentProcessLowerPriority() {
	pid := os.Getpid()
	newPriority := processpriority.BelowNormal
	err := processpriority.Set(pid, newPriority)
	if err != nil {
		panic(err)
	}
}
