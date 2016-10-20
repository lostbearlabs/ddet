package util

import (
	"github.com/juju/loggo"
)

func SetLogInfo() {
	if err := loggo.ConfigureLoggers("<root>=INFO"); err != nil {
		panic(err)
	}
}

func SetLogTrace() {
	if err := loggo.ConfigureLoggers("<root>=TRACE"); err != nil {
		panic(err)
	}
}
