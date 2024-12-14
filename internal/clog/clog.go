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

func (cl *CLogger) clog(c ansi.Color, level string, msg string) {
	head := ansi.SetFont(ansi.DefaultColor, c, ansi.Bold)(level)
	b := &strings.Builder{}
	cl.l.SetOutput(b)
	cl.l.Output(3, msg)
	body := ansi.SetFont(c, ansi.DefaultColor)(b.String())
	toWrite := fmt.Sprintf("%s %s", head, body)
	cl.outMu.Lock()
	defer cl.outMu.Unlock()
	cl.out.Write([]byte(toWrite))
}

func (cl *CLogger) Clog(c ansi.Color, level string, msg string) {
	cl.clog(c, level, msg)
}

const (
	debugColor = ansi.Color8Bit(39)
	infoColor  = ansi.Color8Bit(40)
	warnColor  = ansi.Color8Bit(220)
	errorColor = ansi.Color8Bit(197)
)

func (cl *CLogger) Debug(msg string) {
	cl.clog(debugColor, " DEBUG ", msg)
}

func (cl *CLogger) Info(msg string) {
	cl.clog(infoColor, " INFO ", msg)
}

func (cl *CLogger) Warn(msg string) {
	cl.clog(warnColor, " WARN ", msg)
}

func (cl *CLogger) Error(msg string) {
	cl.clog(errorColor, " ERROR ", msg)
}

func Clog(c ansi.Color, level string, msg string) {
	defaultCLogger.clog(c, level, msg)
}

func Debug(msg string) {
	defaultCLogger.clog(debugColor, " DEBUG ", msg)
}

func Info(msg string) {
	defaultCLogger.clog(infoColor, " INFO ", msg)
}

func Warn(msg string) {
	defaultCLogger.clog(warnColor, " WARN ", msg)
}

func Error(msg string) {
	defaultCLogger.clog(errorColor, " ERROR ", msg)
}
