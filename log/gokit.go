package log

import (
	"io"
	"os"
	"time"

	"github.com/go-kit/kit/log"
)

var logger = newLog(os.Stderr)

type kitlog struct{ ctx log.Logger }

func newLog(w io.Writer) *kitlog {

	return &kitlog{ctx: log.NewJSONLogger(w)}
}

func (l *kitlog) KV(k string, v interface{}) Log {
	switch s := v.(type) {
	case interface {
		String() string
	}:
		v = s.String()
	case interface {
		GoString() string
	}:
		v = s.GoString()
	}
	return &kitlog{ctx: log.With(l.ctx, k, v)}
}

func (l *kitlog) KVs(f F) Log {
	var out Log
	for k, v := range f {
		out = out.KV(k, v)
	}
	return out
}

func (l *kitlog) Err(err error) Log { return l.KV("err", err) }
func (l *kitlog) Info(msg string)   { l.log("info", msg) }
func (l *kitlog) Error(msg string)  { l.log("error", msg) }
func (l *kitlog) Warn(msg string)   { l.log("warn", msg) }
func (l *kitlog) Fatal(msg string)  { l.log("fatal", msg); os.Exit(1) }
func (l *kitlog) Panic(msg string)  { l.log("panic", msg); panic(msg) }

func (l *kitlog) log(lvl, msg string) {
	err := l.ctx.Log(
		"level", lvl,
		"msg", msg,
		"src", log.DefaultCaller(),
		"time", time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		panic(err)
	}
}

func Err(err error) Log              { return logger.Err(err) }
func Info(msg string)                { logger.Info(msg) }
func Error(msg string)               { logger.Error(msg) }
func Warn(msg string)                { logger.Warn(msg) }
func Fatal(msg string)               { logger.Fatal(msg) }
func Panic(msg string)               { logger.Panic(msg) }
func KV(k string, v interface{}) Log { return logger.KV(k, v) }
func KVs(f F) Log                    { return logger.KVs(f) }
