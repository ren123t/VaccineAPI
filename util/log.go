package util

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type LoggerStruct struct {
	*logrus.Logger
}

var Logger LoggerStruct

func SetLogWriter() {
	var logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.Out = os.Stdout
	Logger = LoggerStruct{logger}
}

//can add levels and customized fields as needed with future iterations. breaks
//log fields down into json for easy ingestion by apis and maybe ELK stacks
func (logger LoggerStruct) Log(logValues map[string]interface{}) {

	file, err := os.OpenFile("./logs/logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		logrus.SetLevel(logrus.WarnLevel)
		logpath := "./logs/log_error.log"
		f, _ := os.OpenFile(logpath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		logrus.SetOutput(f)
		defer f.Close()
		logrus.Warnf(err.Error())
	} else {
		logger.Out = file
	}

	if logValues == nil {
		logValues = make(map[string]interface{})
	}
	entry := logger.WithFields(logValues)
	entry.Time = time.Unix(1, 0)
	defer file.Close()
	entry.Errorf("")
}
