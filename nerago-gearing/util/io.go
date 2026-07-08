package util

import "os"

func WriteStringToFile(filename, content string) {
	bytes := []byte(content)
	err := os.WriteFile(filename, bytes, 0666)
	if err != nil {
		panic(err)
	}
}
