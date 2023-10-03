// Copyright (c) 2021 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Copyright (c) 2023 John Rowley
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package fxslog

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"go.uber.org/fx/fxevent"
)

var _ fxevent.Logger = &SlogLogger{}

type SlogLogger struct {
	Logger *slog.Logger

	ctx        context.Context
	logLevel   slog.Level
	errorLevel *slog.Level
}

// UseContext sets the context that will be used when logging to slog.
func (l *SlogLogger) UseContext(ctx context.Context) {
	l.ctx = ctx
}

// UseLogLevel sets the level of non-error logs emitted by Fx to level.
func (l *SlogLogger) UseLogLevel(level slog.Level) {
	l.logLevel = level
}

// UseErrorLevel sets the level of error logs emitted by Fx to level.
func (l *SlogLogger) UseErrorLevel(level slog.Level) {
	l.errorLevel = &level
}

func (l *SlogLogger) logEvent(msg string, fields ...any) {
	l.Logger.Log(l.ctx, l.logLevel, msg, fields...)
}

func (l *SlogLogger) logError(msg string, fields ...any) {
	lvl := slog.LevelError
	if l.errorLevel != nil {
		lvl = *l.errorLevel
	}

	l.Logger.Log(l.ctx, lvl, msg, fields...)
}

// LogEvent logs the given event to the provided Zap logger.
func (l *SlogLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.logEvent("OnStart hook executing",
			slog.String("callee", e.FunctionName),
			slog.String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.logError("OnStart hook failed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slogErr(e.Err),
			)
		} else {
			l.logEvent("OnStart hook executed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slog.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.OnStopExecuting:
		l.logEvent("OnStop hook executing",
			slog.String("callee", e.FunctionName),
			slog.String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.logError("OnStop hook failed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slogErr(e.Err),
			)
		} else {
			l.logEvent("OnStop hook executed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slog.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.logError("error encountered while applying options",
				slog.String("type", e.TypeName),
				slogStrings("stacktrace", e.StackTrace),
				moduleField(e.ModuleName),
				slogErr(e.Err))
		} else {
			l.logEvent("supplied",
				slog.String("type", e.TypeName),
				slogStrings("stacktrace", e.StackTrace),
				moduleField(e.ModuleName),
			)
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.logEvent("provided",
				slog.String("constructor", e.ConstructorName),
				slogStrings("stacktrace", e.StackTrace),
				moduleField(e.ModuleName),
				slog.String("type", rtype),
				maybeBool("private", e.Private),
			)
		}
		if e.Err != nil {
			l.logError("error encountered while applying options",
				moduleField(e.ModuleName),
				slogStrings("stacktrace", e.StackTrace),
				slogErr(e.Err))
		}
	case *fxevent.Replaced:
		for _, rtype := range e.OutputTypeNames {
			l.logEvent("replaced",
				slogStrings("stacktrace", e.StackTrace),
				moduleField(e.ModuleName),
				slog.String("type", rtype),
			)
		}
		if e.Err != nil {
			l.logError("error encountered while replacing",
				slogStrings("stacktrace", e.StackTrace),
				moduleField(e.ModuleName),
				slogErr(e.Err))
		}
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			l.logEvent("decorated",
				slog.String("decorator", e.DecoratorName),
				slogStrings("stacktrace", e.StackTrace),
				moduleField(e.ModuleName),
				slog.String("type", rtype),
			)
		}
		if e.Err != nil {
			l.logError("error encountered while applying options",
				slogStrings("stacktrace", e.StackTrace),
				moduleField(e.ModuleName),
				slogErr(e.Err))
		}
	case *fxevent.Run:
		if e.Err != nil {
			l.logError("error returned",
				slog.String("name", e.Name),
				slog.String("kind", e.Kind),
				moduleField(e.ModuleName),
				slogErr(e.Err),
			)
		} else {
			l.logEvent("run",
				slog.String("name", e.Name),
				slog.String("kind", e.Kind),
				moduleField(e.ModuleName),
			)
		}
	case *fxevent.Invoking:
		// Do not log stack as it will make logs hard to read.
		l.logEvent("invoking",
			slog.String("function", e.FunctionName),
			moduleField(e.ModuleName),
		)
	case *fxevent.Invoked:
		if e.Err != nil {
			l.logError("invoke failed",
				slogErr(e.Err),
				slog.String("stack", e.Trace),
				slog.String("function", e.FunctionName),
				moduleField(e.ModuleName),
			)
		}
	case *fxevent.Stopping:
		l.logEvent("received signal",
			slog.String("signal", strings.ToUpper(e.Signal.String())))
	case *fxevent.Stopped:
		if e.Err != nil {
			l.logError("stop failed", slogErr(e.Err))
		}
	case *fxevent.RollingBack:
		l.logError("start failed, rolling back", slogErr(e.StartErr))
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.logError("rollback failed", slogErr(e.Err))
		}
	case *fxevent.Started:
		if e.Err != nil {
			l.logError("start failed", slogErr(e.Err))
		} else {
			l.logEvent("started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.logError("custom logger initialization failed", slogErr(e.Err))
		} else {
			l.logEvent("initialized custom fxevent.Logger", slog.String("function", e.ConstructorName))
		}
	}
}

func moduleField(name string) slog.Attr {
	if len(name) == 0 {
		return slog.Group(name, []any{}...)
	}

	return slog.String(name, name)
}

func maybeBool(name string, b bool) slog.Attr {
	if !b {
		return slog.Group(name, []any{}...)
	}

	return slog.Bool(name, true)
}

func slogErr(err error) slog.Attr {
	return slog.String("err", fmt.Sprintf("%+v", err))
}

func slogStrings(key string, str []string) slog.Attr {
	var attrs []any
	for i, val := range str {
		attrs = append(attrs, slog.String(fmt.Sprintf("%d", i), val))
	}

	return slog.Group(key, attrs...)
}
