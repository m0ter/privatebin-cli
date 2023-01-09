package utils

import (
	"bytes"
	"io"
	"os"
)

func ReadStdin() string {
	data, err := io.ReadAll(os.Stdin)

	if bytes.HasSuffix(data, []byte("\n")) {
		data = data[:len(data)-1]
	}

	if err != nil {
		panic(err)
	}

	return string(data)
}

func IsStdin() bool {
	stat, _ := os.Stdin.Stat()

	return (stat.Mode() & os.ModeCharDevice) == 0
}
