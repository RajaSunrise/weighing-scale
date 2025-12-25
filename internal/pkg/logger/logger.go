package logger

import (
	"io"
	"log"
	"os"
)

var LogFile *os.File

func Init() {
	// Create logs directory
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}

	// Open file
	file, err := os.OpenFile("logs/system.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	LogFile = file

	// MultiWriter: Write to both file and stdout
	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)

	log.Println("Logger initialized")
}
