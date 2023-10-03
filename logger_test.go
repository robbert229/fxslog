package fxslog_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/robbert229/fxslog"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func TestLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	loggerConstructor := func() *slog.Logger {
		return slog.New(slog.NewJSONHandler(buf, nil))
	}
	app := fx.New(
		fx.Provide(
			loggerConstructor,
		),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return &fxslog.SlogLogger{
				Logger: logger,
			}
		}),
	)

	go func() {
		time.Sleep(time.Second * 3)
		app.Stop(context.Background())
	}()
	err := app.Start(context.Background())
	if err != nil {
		t.Fatalf("failed to start app: %+v", err)
	}

	strings.Contains(buf.String(), `"msg":"started"`)
	fmt.Println(buf.String())
}
