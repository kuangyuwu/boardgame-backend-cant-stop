package clog

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	ansi "github.com/kuangyuwu/boardgame-backend-cant-stop/internal/ansi"
)

type CLogger struct {
	l     *log.Logger
	out   io.Writer
	outMu *sync.Mutex
}

func New(out io.Writer, prefix string, flag int) *CLogger {
	cl := CLogger{log.New(io.Discard, prefix, flag), out, new(sync.Mutex)}
	return &cl
}

func (cl *CLogger) SetOutput(w io.Writer) {
	cl.outMu.Lock()
	defer cl.outMu.Unlock()
	cl.out = w
}

func (cl *CLogger) SetPrefix(prefix string) {
	cl.l.SetPrefix(prefix)
}

func (cl *CLogger) SetFlags(flag int) {
	cl.l.SetFlags(flag)
}

const defaultFlag = log.Ldate | log.Ltime | log.Lshortfile

var defaultCLogger = New(os.Stdout, "", defaultFlag)

func (cl *CLogger) output(c ansi.Color, tag string, msg string, calldepth int) {
	cl.outMu.Lock()
	defer cl.outMu.Unlock()

	head := ansi.SetFont(ansi.DefaultColor, c, ansi.Bold)(fmt.Sprintf(" %s ", tag))
	b := &strings.Builder{}
	cl.l.SetOutput(b)
	cl.l.Output(calldepth, msg)
	body := ansi.SetFont(c, ansi.DefaultColor)(b.String())
	toWrite := fmt.Sprintf("%s %s", head, body)
	cl.out.Write([]byte(toWrite))
}

// func (cl *CLogger) Clog(c ansi.Color, level string, msg string) {
// 	cl.output(c, level, msg, 3)
// }

type level struct {
	color ansi.Color
	tag   string
}

var (
	debugLevel = level{ansi.Color8Bit(39), "DEBUG"}
	infoLevel  = level{ansi.Color8Bit(40), "INFO"}
	warnLevel  = level{ansi.Color8Bit(220), "WARN"}
	errorLevel = level{ansi.Color8Bit(197), "ERROR"}
)

func (cl *CLogger) clog(l level, a ...any) {
	msg := fmt.Sprint(a...)
	defaultCLogger.output(l.color, l.tag, msg, 4)
}

func (cl *CLogger) clogf(l level, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	defaultCLogger.output(l.color, l.tag, msg, 4)
}

func (cl *CLogger) clogln(l level, a ...any) {
	msg := fmt.Sprintln(a...)
	defaultCLogger.output(l.color, l.tag, msg, 4)
}

// func (cl *CLogger) Debug(msg string) {
// 	cl.output(debugColor, "DEBUG", msg, 3)
// }

// func (cl *CLogger) Info(msg string) {
// 	cl.output(infoColor, "INFO", msg, 3)
// }

// func (cl *CLogger) Warn(msg string) {
// 	cl.output(warnColor, "WARN", msg, 3)
// }

// func (cl *CLogger) Error(msg string) {
// 	cl.output(errorColor, "ERROR", msg, 3)
// }

// func Clog(c ansi.Color, level string, msg string) {
// 	defaultCLogger.output(c, level, msg, 3)
// }

func Debug(a ...any) {
	defaultCLogger.clog(debugLevel, a...)
}

func Debugf(format string, a ...any) {
	defaultCLogger.clogf(debugLevel, format, a...)
}

func Debugln(a ...any) {
	defaultCLogger.clogln(debugLevel, a...)
}

func Info(a ...any) {
	defaultCLogger.clog(infoLevel, a...)
}

func Infof(format string, a ...any) {
	defaultCLogger.clogf(infoLevel, format, a...)
}

func Infoln(a ...any) {
	defaultCLogger.clogln(infoLevel, a...)
}

func Warn(a ...any) {
	defaultCLogger.clog(warnLevel, a...)
}

func Warnf(format string, a ...any) {
	defaultCLogger.clogf(warnLevel, format, a...)
}

func Warnln(a ...any) {
	defaultCLogger.clogln(warnLevel, a...)
}

func Error(a ...any) {
	defaultCLogger.clog(errorLevel, a...)
}

func Errorf(format string, a ...any) {
	defaultCLogger.clogf(errorLevel, format, a...)
}

func Errorln(a ...any) {
	defaultCLogger.clogln(errorLevel, a...)
}

// func Info(msg string) {
// 	defaultCLogger.output(infoColor, " INFO ", msg, 3)
// }

// func Warn(msg string) {
// 	defaultCLogger.output(warnColor, " WARN ", msg, 3)
// }

// func Error(msg string) {
// 	defaultCLogger.output(errorColor, " ERROR ", msg, 3)
// }
