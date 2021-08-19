package main

import (
	"log"
	"log/syslog"

	"github.com/dumacp/go-logs/pkg/logs"
)

func newLog(logger *logs.Logger, prefix string, flags int, priority int) error {

	logg, err := syslog.NewLogger(syslog.Priority(priority), flags)
	if err != nil {
		return err
	}
	logger.SetLogError(logg)
	return nil
}

func newLogHtml(logger *logs.Logger, prefix string, flags int, priority int) error {

	logg, err := syslog.NewLogger(syslog.Priority(priority), flags)
	if err != nil {
		return err
	}
	logger.SetLogError(logg)
	return nil
}

func initLogs(verbose int, logStd bool) {
	defer func() {
		if verbose < 4 {
			logs.LogBuild.Disable()
		}
		if verbose < 3 {
			logs.LogInfo.Disable()
		}
		if verbose < 2 {
			logs.LogWarn.Disable()
		}
		if verbose < 1 {
			logs.LogError.Disable()
		}
	}()
	if logStd {
		return
	}
	newLog(logs.LogWarn, "[ warn ] ", log.LstdFlags, 4)
	newLog(logs.LogInfo, "[ info ] ", log.LstdFlags, 6)
	newLog(logs.LogBuild, "[ build ] ", log.LstdFlags, 7)
	newLog(logs.LogError, "[ error ] ", log.LstdFlags, 3)

}
