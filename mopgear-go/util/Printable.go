package util

import (
	"fmt"
	"io"
)

type Printable interface {
	Println0()
	Println(str string)
	PrintlnFromBuild(strBuild StringBuild2)
	Printf(format string, args ...any)
	PrintBytes(bytes []byte)
}

var _ Printable = PrintRecorder_Nop()

type filePrintable struct {
	write io.Writer
}

func FilePrintableMake(write io.Writer) Printable {
	return &filePrintable{write}
}

func (fp *filePrintable) Println0() {
	_, err := fp.write.Write([]byte{'\n'})
	if err != nil {
		panic(err)
	}
}

func (fp *filePrintable) Println(str string) {
	_, err := io.WriteString(fp.write, str)
	if err != nil {
		panic(err)
	}
	_, err = fp.write.Write([]byte{'\n'})
	if err != nil {
		panic(err)
	}
}

func (fp *filePrintable) PrintlnFromBuild(strBuild StringBuild2) {
	_, err := fp.write.Write(strBuild)
	if err != nil {
		panic(err)
	}
	_, err = fp.write.Write([]byte{'\n'})
	if err != nil {
		panic(err)
	}
}

func (fp *filePrintable) Printf(format string, args ...any) {
	_, err := fmt.Fprintf(fp.write, format, args...)
	if err != nil {
		panic(err)
	}
}

func (fp *filePrintable) PrintBytes(bytes []byte) {
	_, err := fp.write.Write(bytes)
	if err != nil {
		panic(err)
	}
}
