package main

import (
	"bytes"
	"fmt"
	"log/syslog"
	"time"

	"github.com/dumacp/go-fareCollection/internal/logstrans"
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

func newLogHtml(logger *logstrans.Logger, dir, prefixFile string) error {

	funcWrite := func(output []byte) []byte {
		varSplit := bytes.SplitAfterN(output, []byte(":"), 2)
		title := make([]byte, 0)
		dataMsg := output
		if len(varSplit) > 1 {
			title = varSplit[0]
			dataMsg = bytes.TrimSpace(varSplit[1])
		}
		dataMsg = bytes.ReplaceAll(dataMsg, []byte("\n"), []byte(""))
		tmpl := `{"timestamp": %d, "type": "%s" , "title": "%s", "message": %s}
`
		if !bytes.HasPrefix(dataMsg, []byte("{")) {
			tmpl = `{"timestamp": %d, "type": "%s" , "title": "%s", "message": %q}
`
		}
		// log.Printf("splt: %s\n", varSplit)
		data := []byte(fmt.Sprintf(tmpl,
			time.Now().UnixNano()/1_000_000, logger.Type, title, dataMsg))
		return data

	}
	logg, err := logs.NewRotateWithFuncWriter(funcWrite, dir, prefixFile, 1024*1024*2, 20, 0)
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
	newLog(logs.LogWarn, "[ warn ] ", 0, 4)
	newLog(logs.LogInfo, "[ info ] ", 0, 6)
	newLog(logs.LogBuild, "[ build ] ", 0, 7)
	newLog(logs.LogError, "[ error ] ", 0, 3)

}

func initLogsTransactional(dir, prefixFile string, verbose int, logStd bool) {
	defer func() {
		if verbose < 4 {
			logstrans.LogBuild.Disable()
		}
		if verbose < 3 {
			logstrans.LogInfo.Disable()
		}
		if verbose < 2 {
			logstrans.LogWarn.Disable()
		}
		if verbose < 1 {
			logstrans.LogError.Disable()
		}
	}()
	if logStd {
		return
	}

	newLogHtml(logstrans.LogWarn, dir, prefixFile)
	newLogHtml(logstrans.LogInfo, dir, prefixFile)
	newLogHtml(logstrans.LogBuild, dir, prefixFile)
	newLogHtml(logstrans.LogError, dir, prefixFile)
}
