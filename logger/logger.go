/*
Package logger provide the log rotate feature.
*/
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var lastTime time.Time
var logQueue = make(chan string, 1000)
var debugMode = true

func doRotate(fPrefix string) {
	preFileName := fPrefix + ".log"
	curFileName := fPrefix + "-" + lastTime.Format("01-02") + ".log"
	_, err := os.Stat(preFileName)
	if err == nil {
		fmt.Println("rename : ", preFileName, " => ", curFileName)
		err = os.Rename(preFileName, curFileName)
		if err != nil {
			fmt.Println(os.Getwd())
			fmt.Println(err)
		}
	}

}

func newLogger(fPrefix string) (*log.Logger, *os.File) {
	lastTime = time.Now()
	var logger *log.Logger
	fileName := fmt.Sprintf("%s.log", fPrefix)
	fmt.Println("fileName :", fileName)
	logFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Println("open file error!")
	} else {
		if debugMode == true {
			logger = log.New(io.MultiWriter(os.Stdout, logFile), "", log.Ldate|log.Ltime|log.Lshortfile)
		}
		logger = log.New(logFile, "", log.Ldate|log.Ltime)
	}
	return logger, logFile
}

func checkCondition() bool {
	currentTime := time.Now()
	if currentTime.Day() != lastTime.Day() {
		fmt.Printf("%d != %d", currentTime.Day(), lastTime.Day())
		return false
	}
	return true
}

func logWorker(msgQueue <-chan string) {
	fPrefix := "ChatServer"
	logger, logFile := newLogger(fPrefix)
	for msg := range msgQueue {
		_, err := logFile.Stat()
		if err == nil {
			if checkCondition() == false {
				logFile.Close()
				doRotate(fPrefix)
				logger, logFile = newLogger(fPrefix)
			}
		}
		logger.Print(msg)
	}
	logFile.Close()
}

// DebugLog : Print a log with DEBUG tag
func DebugLog(v ...interface{}) {
	logQueue <- "[DEBUG]: " + fmt.Sprintln(v...)
}

// InfoLog : Print a log with INFO tag
func InfoLog(v ...interface{}) {
	logQueue <- "[INFO]: " + fmt.Sprintln(v...)
}

// WarningLog : Print a log with WARNING tag
func WarningLog(v ...interface{}) {
	logQueue <- "[WARNING]: " + fmt.Sprintln(v...)
}

// ErrorLog : Print a log with ERROR tag
func ErrorLog(v ...interface{}) {
	logQueue <- "[ERROR]: " + fmt.Sprintln(v...)
}

// Init : Start the tool system
func Init() {
	go logWorker(logQueue)
	InfoLog("Log system start!")
}

// UnInit : Stop the tool system
func UnInit() {
	close(logQueue)
	InfoLog("Log system stop!")
}
