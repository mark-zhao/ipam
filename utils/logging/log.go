package logging

import (
	"fmt"
	"ipam/utils/options"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	FATAL
)

var (
	F *os.File

	DefaultPrefix      = ""
	DefaultCallerDepth = 2

	logger     *log.Logger
	logPrefix  = ""
	levelFlags = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	logLevel   Level // set default logLevel as INFO
)

func ConfigInit() {
	LOG := &Log{
		options.Conf.Log.LogDir,
		options.Conf.Log.LogFile,
		options.Conf.Log.LogFileExt,
		options.Conf.Log.TimeFormat,
	}
	filePath := LOG.getLogFileFullPath()
	F = LOG.openLogFile(filePath)
	switch options.Conf.Http.RunMode {
	case "debug":
		logLevel = DEBUG
	case "info":
		logLevel = INFO
	case "warning":
		logLevel = WARNING
	case "ERROR":
		logLevel = ERROR
	case "fatal":
		logLevel = FATAL
	default:
		logLevel = INFO
	}

	logger = log.New(F, DefaultPrefix, log.LstdFlags)
}

func Debug(v ...interface{}) {
	if logLevel <= DEBUG {
		setPrefix(DEBUG)
		logger.Println(v...)
	}
}

func Info(v ...interface{}) {
	if logLevel <= INFO {
		setPrefix(INFO)
		logger.Println(v...)
	}
}

func Warn(v ...interface{}) {
	if logLevel <= WARNING {
		setPrefix(WARNING)
		logger.Println(v...)
	}
}

func Error(v ...interface{}) {
	if logLevel <= FATAL {
		setPrefix(ERROR)
		logger.Printf("%+v", v)
	}
}

func Fatal(v ...interface{}) {
	setPrefix(FATAL)
	logger.Fatalln(v...)
}

func setPrefix(level Level) {
	_, file, line, ok := runtime.Caller(DefaultCallerDepth)
	if ok {
		logPrefix = fmt.Sprintf("[%s][%s:%d]", levelFlags[level], filepath.Base(file), line)
	} else {
		logPrefix = fmt.Sprintf("[%s]", levelFlags[level])
	}

	logger.SetPrefix(logPrefix)
}
