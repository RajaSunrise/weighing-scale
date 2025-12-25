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
		// Try to create, but don't fail if we can't
		_ = os.Mkdir("logs", 0755)
	}

	// Open file
	file, err := os.OpenFile("logs/system.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		// If we can't open the log file (e.g. permission denied), fall back to stdout
		log.Println("WARNING: Could not open logs/system.log:", err)
		log.Println("Falling back to standard output logging.")
		log.SetOutput(os.Stdout)
		return
	}
	LogFile = file

	// MultiWriter: Write to both file and stdout
	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)

	log.Println("Logger initialized")
}
