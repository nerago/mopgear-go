package util_async

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var modkernel32 = windows.NewLazySystemDLL("kernel32.dll")
var procReadConsoleInput = modkernel32.NewProc("ReadConsoleInputW")

type INPUT_RECORD struct {
	EventType uint16
	KeyEvent  KEY_EVENT_RECORD
}

type KEY_EVENT_RECORD struct {
	bKeyDown          int32
	wRepeatCount      uint16
	wVirtualKeyCode   uint16
	wVirtualScanCode  uint16
	UnicodeChar       uint16
	dwControlKeyState uint32
}

func readConsoleInput(console windows.Handle, buf *INPUT_RECORD, length uint32, numRead *uint32) (err error) {
	r1, _, e1 := syscall.SyscallN(procReadConsoleInput.Addr(), uintptr(console), uintptr(unsafe.Pointer(buf)), uintptr(length), uintptr(unsafe.Pointer(numRead)))
	if r1 == 0 {
		err = e1
	}
	return
}

func waitForKeyPressCancellable(cancel CancelSignal) (bool, error) {
	stdinHandle, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		return false, err
	}
	err = windows.SetConsoleMode(stdinHandle, 0)
	if err != nil {
		return false, err
	}
	err = windows.FlushConsoleInputBuffer(stdinHandle)
	if err != nil {
		return false, err
	}

	for cancel.ShouldContinue() {
		waitResult, err := windows.WaitForSingleObject(stdinHandle, 1000)
		if err != nil {
			return false, err
		}

		if waitResult == 0 {
			inputRecord := INPUT_RECORD{}
			numRead := uint32(0)
			err = readConsoleInput(stdinHandle, &inputRecord, 1, &numRead)
			if err != nil {
				return false, err
			}

			if numRead > 0 && inputRecord.EventType == windows.KEY_EVENT {
				return true, nil
			}
		}
	}

	return false, nil
}
