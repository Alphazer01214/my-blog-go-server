package logs

import (
	"fmt"
	"time"
)

type Log struct {
	Time    time.Time
	Module  string
	Type    string // info, warning, error
	Message string
	Level   int
}

type Logman struct {
	Logs   []Log
	Mode   string // dev, prod
	Module string
	Path   string

	PrintLevel int
}

func NewLogman(module string, mode string, level int) *Logman {
	return &Logman{
		Logs:       []Log{},
		Mode:       mode,
		PrintLevel: level,
		Module:     module,
	}
}

func (l *Logman) Log(m string, t string, msg string, level int) {
	log := Log{
		Time:    time.Now(),
		Module:  m,
		Type:    t,
		Message: msg,
		Level:   level,
	}
	//print(log)
	l.Logs = append(l.Logs, log)
}

func (l *Logman) Info(msg string) {
	log := Log{
		Time:    time.Now(),
		Module:  l.Module,
		Type:    "info",
		Message: msg,
		Level:   0,
	}
	//print(log)
	l.Logs = append(l.Logs, log)
}

func (l *Logman) Warn(msg string) {
	log := Log{
		Time:    time.Now(),
		Module:  l.Module,
		Type:    "warning",
		Message: msg,
		Level:   1,
	}
	//print(log)
	l.Logs = append(l.Logs, log)
}

func (l *Logman) Error(msg error) {
	log := Log{
		Time:    time.Now(),
		Module:  l.Module,
		Type:    "error",
		Message: fmt.Sprintf("%v", msg),
		Level:   2,
	}
	//print(log)
	l.Logs = append(l.Logs, log)
}

func (l *Logman) print(log Log) {
	if log.Level >= l.PrintLevel {
		// [time] [type] module: message
		fmt.Printf("[%v] [%v] %v : %v", log.Time, log.Type, log.Module, log.Message)
	}
}
