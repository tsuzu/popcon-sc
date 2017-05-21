package main

import (
	"github.com/cs3238-tsuzu/chan-utils"
)

var programExitedNotifier chanUtils.ExitedNotifier

func InitProgramExitedNotifier() {
	programExitedNotifier = chanUtils.NewExitedNotifier()
}
