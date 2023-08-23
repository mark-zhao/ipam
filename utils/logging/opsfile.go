package logging

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Log struct {
	LogDir     string
	LogFile    string
	LogFileExt string
	TimeFormat string
}

func (L *Log) getLogFilePath() string {
	return L.LogDir
}

func (L *Log) getLogFileFullPath() string {
	prefixPath := L.getLogFilePath()
	suffixPath := fmt.Sprintf("%s%s.%s", L.LogFile, time.Now().Format(L.TimeFormat), L.LogFileExt)
	return fmt.Sprintf("%s%s", prefixPath, suffixPath)
}

func (L *Log) openLogFile(filePath string) *os.File {
	_, err := os.Stat(filePath)
	switch {
	case os.IsNotExist(err):
		L.mkDir()
	case os.IsPermission(err):
		log.Fatalf("Permission :%v", err)
	}

	handle, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Fail to OpenFile :%v", err)
	}

	return handle
}

func (L *Log) mkDir() {
	//dir, _ := os.Getwd()
	//err := os.MkdirAll(dir+"/"+self.getLogFilePath(), os.ModePerm)
	err := os.MkdirAll(L.getLogFilePath(), os.ModePerm)
	if err != nil {
		panic(err)
	}
}
