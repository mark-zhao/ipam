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

func (self *Log) getLogFilePath() string {
	return fmt.Sprintf("%s", self.LogDir)
}

func (self *Log) getLogFileFullPath() string {
	prefixPath := self.getLogFilePath()
	suffixPath := fmt.Sprintf("%s%s.%s", self.LogFile, time.Now().Format(self.TimeFormat), self.LogFileExt)
	return fmt.Sprintf("%s%s", prefixPath, suffixPath)
}

func (self *Log) openLogFile(filePath string) *os.File {
	_, err := os.Stat(filePath)
	switch {
	case os.IsNotExist(err):
		self.mkDir()
	case os.IsPermission(err):
		log.Fatalf("Permission :%v", err)
	}

	handle, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Fail to OpenFile :%v", err)
	}

	return handle
}

func (self *Log) mkDir() {
	//dir, _ := os.Getwd()
	//err := os.MkdirAll(dir+"/"+self.getLogFilePath(), os.ModePerm)
	err := os.MkdirAll(self.getLogFilePath(), os.ModePerm)
	if err != nil {
		panic(err)
	}
}
