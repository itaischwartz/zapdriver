package zapdriver

import (
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type fakeErr struct{}

// manually set the frames to allow asserting stacktraces
func (fakeErr) StackTrace() errors.StackTrace {
	pc1, _, _, _ := runtime.Caller(0)
	pc2, _, _, _ := runtime.Caller(0)
	return []errors.Frame{
		errors.Frame(pc1),
		errors.Frame(pc2),
	}
}
func (fakeErr) Error() string {
	return "fake error: underlying error"
}

/*
func TestFmtStack(t *testing.T) {
	stacktrace := stackdriverFmtError{fakeErr{}}.Error()
	assert.Equal(t, `fake error: underlying error

goroutine 1 [running]:
github.com/blendle/zapdriver.fakeErr.StackTrace()
	/error_test.go:18 +0x1337
github.com/blendle/zapdriver.fakeErr.StackTrace()
	/error_test.go:19 +0x1337`, makeStackTraceStable(stacktrace))
}
*/

// cleanup local paths & local function pointers
func makeStackTraceStable(str string) string {
	re := regexp.MustCompile(`(?m)^\t.+(\/\S+:\d+) \+0x.+$`)
	str = re.ReplaceAllString(str, "\t${1} +0x1337")
	dir, _ := os.Getwd()
	str = strings.ReplaceAll(str, dir, "")
	return str
}

func ExampleSkipFmtStackTraces() {
	logger, _ := NewProduction()
	logger.Error("with exception", zap.Error(errors.New("internal error")), ErrorReport(runtime.Caller(0)))

	logger, _ = NewProduction(WrapCore(ServiceName("service"), ReportAllErrors(true)))
	logger.Error("with exception", zap.Error(errors.New("internal error")))

	logger, _ = NewProduction(WrapCore(ServiceName("service"), SkipFmtStackTraces(true)))
	logger.Error("without exception", zap.Error(errors.New("internal error")))

	// Output:
}

func TestStackTrace(t *testing.T) {
	config := NewProductionConfig()
	logger, err := config.Build()
	require.NoError(t, err)
	core, logs := observer.New(zap.InfoLevel)
	logger = logger.WithOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return core
	}), WrapCore(ReportAllErrors(true)))

	err = errors.New("internal error")
	logger.Error("error", zap.Error(err))
	logger.Sugar().With(zap.Error(err)).Error("error2")
	logger.Sugar().Info("test3")
	logger.Sugar().Error("test4")

	require.NotEmpty(t, logs.All()[0].ContextMap()["exception"])
	require.NotEmpty(t, logs.All()[1].ContextMap()["exception"])
	require.Empty(t, logs.All()[2].ContextMap()["exception"])
	require.Empty(t, logs.All()[3].ContextMap()["exception"])
}
