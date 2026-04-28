package util

import (
	"os"
	"os/exec"
	"strconv"
)

func CurrentProcessLowerPriority() {
	pid := strconv.Itoa(os.Getpid())
	cmd := exec.Command(`renice -n 10 -p ` + pid)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
