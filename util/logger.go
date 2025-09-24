package util

import (
	"log"
	"os"
)

func SetupLogging() *log.Logger {
	output := os.Stdout
	if logFile := os.Getenv("LOGFILE"); logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		output = f
		log.Printf("Logging to file: %s\n", logFile)
	} else {
		log.Println("Logging to Stdout")
	}

	return log.New(output, "", log.LstdFlags|log.Lshortfile)
}
