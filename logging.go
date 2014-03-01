package logging

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type (
	Color int
	Level int
)

// Colors for different log levels.
const (
	BLACK Color = (iota + 30)
	RED
	GREEN
	YELLOW
	BLUE
	MAGENTA
	CYAN
	WHITE
)

// Logging levels.
const (
	CRITICAL Level = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

var LevelNames = map[Level]string{
	CRITICAL: "CRITICAL",
	ERROR:    "ERROR",
	WARNING:  "WARNING",
	NOTICE:   "NOTICE",
	INFO:     "INFO",
	DEBUG:    "DEBUG",
}

var LevelColors = map[Level]Color{
	CRITICAL: MAGENTA,
	ERROR:    RED,
	WARNING:  YELLOW,
	NOTICE:   GREEN,
	INFO:     WHITE,
	DEBUG:    CYAN,
}

var (
	DefaultLevel   = INFO
	DefaultHandler = StderrHandler
)

// Logger is the interface for outputing log messages in different levels.
// A new Logger can be created with NewLogger() function.
// You can changed the output handler with SetHandler() function.
type Logger interface {
	// SetLevel changes the level of the logger. Default is logging.Info.
	SetLevel(Level)

	// SetHandler replaces the current handler for output. Default is logging.StderrHandler.
	SetHandler(Handler)

	// Close handlers.
	Close()

	// Fatal is equivalent to l.Critical followed by a call to os.Exit(1).
	Fatal(format string, args ...interface{})

	// Panic is equivalent to l.Critical followed by a call to panic().
	Panic(format string, args ...interface{})

	// Critical logs a message using CRITICAL as log level.
	Critical(format string, args ...interface{})

	// Error logs a message using ERROR as log level.
	Error(format string, args ...interface{})

	// Warning logs a message using WARNING as log level.
	Warning(format string, args ...interface{})

	// Notice logs a message using NOTICE as log level.
	Notice(format string, args ...interface{})

	// Info logs a message using INFO as log level.
	Info(format string, args ...interface{})

	// Debug logs a message using DEBUG as log level.
	Debug(format string, args ...interface{})
}

// Handler is the main component of Logger that handles the output.
type Handler interface {
	// Log one message to output.
	Log(message string)

	// Close the handler.
	Close()
}

// Context contains information about a log message.
type Context struct {
	Name     string
	Level    Level
	Time     time.Time
	Filename string
	Line     int
}

///////////////////////////
//                       //
// Logger implementation //
//                       //
///////////////////////////

// logger is the default Logger implementation.
type logger struct {
	Name    string
	Level   Level
	Handler Handler
}

// NewLogger returns a new Logger implementation. Do not forget to close it at exit.
func NewLogger(name string) Logger {
	return &logger{
		Name:    name,
		Level:   DefaultLevel,
		Handler: DefaultHandler,
	}
}

func (l *logger) Close() {
	l.Handler.Close()
}

func (l *logger) SetLevel(level Level) {
	l.Level = level
}

func (l *logger) SetHandler(b Handler) {
	l.Handler = b
}

func (l *logger) log(level Level, format string, args ...interface{}) {
	// Add missing newline at the end.
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}

	ctx := &Context{
		Name:     l.Name,
		Level:    level,
		Time:     time.Now(),
		Filename: file,
		Line:     line,
	}

	l.Handler.Log(format, args, ctx)
}

func (l *logger) Fatal(format string, args ...interface{}) {
	l.Critical(format, args...)
	l.Close()
	os.Exit(1)
}

func (l *logger) Panic(format string, args ...interface{}) {
	l.Critical(format, args...)
	l.Close()
	panic(fmt.Sprintf(format, args...))
}

func (l *logger) Critical(format string, args ...interface{}) {
	if l.Level >= CRITICAL {
		l.log(CRITICAL, format, args...)
	}
}

func (l *logger) Error(format string, args ...interface{}) {
	if l.Level >= ERROR {
		l.log(ERROR, format, args...)
	}
}

func (l *logger) Warning(format string, args ...interface{}) {
	if l.Level >= WARNING {
		l.log(WARNING, format, args...)
	}
}

func (l *logger) Notice(format string, args ...interface{}) {
	if l.Level >= NOTICE {
		l.log(NOTICE, format, args...)
	}
}

func (l *logger) Info(format string, args ...interface{}) {
	if l.Level >= INFO {
		l.log(INFO, format, args...)
	}
}

func (l *logger) Debug(format string, args ...interface{}) {
	if l.Level >= DEBUG {
		l.log(DEBUG, format, args...)
	}
}

///////////////////
//               //
// DefaultLogger //
//               //
///////////////////

var DefaultLogger = NewLogger("")

func Fatal(format string, args ...interface{}) {
	DefaultLogger.Fatal(format, args...)
}

func Panic(format string, args ...interface{}) {
	DefaultLogger.Panic(format, args...)
}

func Critical(format string, args ...interface{}) {
	DefaultLogger.Critical(format, args...)
}

func Error(format string, args ...interface{}) {
	DefaultLogger.Error(format, args...)
}

func Warning(format string, args ...interface{}) {
	DefaultLogger.Warning(format, args...)
}

func Notice(format string, args ...interface{}) {
	DefaultLogger.Notice(format, args...)
}

func Info(format string, args ...interface{}) {
	DefaultLogger.Info(format, args...)
}

func Debug(format string, args ...interface{}) {
	DefaultLogger.Debug(format, args...)
}

///////////////////
//               //
// WriterHandler //
//               //
///////////////////

// WriterHandler is a handler implementation that writes the logging output to a io.Writer.
type WriterHandler struct {
	w io.Writer
}

func NewWriterHandler(w io.Writer) *WriterHandler {
	return &WriterHandler{w: w}
}

func (b *WriterHandler) Log(format string, args []interface{}, c *Context) {
	fmt.Fprint(b.w, prefix(c)+fmt.Sprintf(format, args...))
}

func (b *WriterHandler) Close() {}

func prefix(c *Context) string {
	return fmt.Sprintf("%s %s %-8s ", fmt.Sprint(c.Time)[:19], c.Name, LevelNames[c.Level])
}

////////////////////
//                //
// ConsoleHandler //
//                //
////////////////////

type ConsoleHandler struct {
	wb *WriterHandler
}

func NewConsoleHandler(w io.Writer) *ConsoleHandler {
	return &ConsoleHandler{wb: NewWriterHandler(w)}
}

func (b *ConsoleHandler) Log(format string, args []interface{}, c *Context) {
	b.wb.w.Write([]byte(fmt.Sprintf("\033[%dm", LevelColors[c.Level])))
	b.wb.Log(format, args, c)
	b.wb.w.Write([]byte("\033[0m")) // reset color

}

func (b *ConsoleHandler) Close() {}

var StderrHandler = NewConsoleHandler(os.Stderr)
var StdoutHandler = NewConsoleHandler(os.Stdout)

///////////////////
//               //
// SyslogHandler //
//               //
///////////////////

// SyslogHandler sends the logging output to syslog.
type SyslogHandler struct {
	w *syslog.Writer
}

func NewSyslogHandler(tag string) (*SyslogHandler, error) {
	// Priority in New constructor is not important here because we
	// do not use w.Write() directly.
	w, err := syslog.New(syslog.LOG_INFO|syslog.LOG_USER, tag)
	if err != nil {
		return nil, err
	}
	return &SyslogHandler{w: w}, nil
}

func (b *SyslogHandler) Log(format string, args []interface{}, c *Context) {
	var fn func(string) error
	switch c.Level {
	case CRITICAL:
		fn = b.w.Crit
	case ERROR:
		fn = b.w.Err
	case WARNING:
		fn = b.w.Warning
	case NOTICE:
		fn = b.w.Notice
	case INFO:
		fn = b.w.Info
	case DEBUG:
		fn = b.w.Debug
	}
	fn(fmt.Sprintf(format, args...))
}

func (b *SyslogHandler) Close() {
	b.w.Close()
}

//////////////////
//              //
// MultiHandler //
//              //
//////////////////

// MultiHandler sends the log output to multiple handlers concurrently.
type MultiHandler struct {
	handlers []Handler
}

func NewMultiHandler(handlers ...Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (b *MultiHandler) Log(format string, args []interface{}, ctx *Context) {
	wg := sync.WaitGroup{}
	wg.Add(len(b.handlers))
	for _, handler := range b.handlers {
		go func(handler Handler) {
			handler.Log(format, args, ctx)
			wg.Done()
		}(handler)
	}
	wg.Wait()
}

func (b *MultiHandler) Close() {
	wg := sync.WaitGroup{}
	wg.Add(len(b.handlers))
	for _, handler := range b.handlers {
		go func(handler Handler) {
			handler.Close()
			wg.Done()
		}(handler)
	}
	wg.Wait()
}
