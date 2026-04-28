package util

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func CurrentProcessLowerPriority() {
	// NOTE go command mangles the double quote in priority if allowed to build command line
	pid := strconv.Itoa(os.Getpid())
	cmd := exec.Command(`C:\Windows\System32\wbem\wmic.exe`)
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `C:\\Windows\\System32\\wbem\\wmic.exe process where ProcessId=` + pid + ` CALL setpriority "below normal"`}
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
