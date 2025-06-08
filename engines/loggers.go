package engines

import (
	"io"
	"log"
	"os"
)

// NewLogger returns a file logger writing into the path
func NewLogger(path string) *log.Logger {
	var logFile io.Writer
	os.Remove(path)
	if file, err := os.Create(path); err != nil {
		panic(err)
	} else {
		logFile = io.MultiWriter(file)
	}

	return log.New(logFile, "", log.Ldate|log.Ltime|log.Llongfile)
}
