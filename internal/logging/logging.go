package logging

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type Level string

const (
	Trace Level = "trace"
	Debug Level = "debug"
	Info  Level = "info"
	Warn  Level = "warn"
	Error Level = "error"
	Fatal Level = "fatal"
)

// / Convert a logging level to a logrus level (uint32).
func (l Level) Integer() logrus.Level {
	switch l {
	case Trace:
		return logrus.TraceLevel
	case Debug:
		return logrus.DebugLevel
	case Info:
		return logrus.InfoLevel
	case Warn:
		return logrus.WarnLevel
	case Error:
		return logrus.ErrorLevel
	case Fatal:
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}

// / Convert a string to a logging level.
func LevelFromString(level string) Level {
	switch level {
	case string(Trace):
		return Trace
	case string(Debug):
		return Debug
	case string(Info):
		return Info
	case string(Warn):
		return Warn
	case string(Error):
		return Error
	case string(Fatal):
		return Fatal
	default:
		return Info
	}
}

var GlobalLogger = logrus.New()

func Init(level Level) {
	GlobalLogger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
		DisableQuote:  true,
	})

	GlobalLogger.SetReportCaller(false)
	GlobalLogger.SetLevel(level.Integer())

	file, err := os.OpenFile("ie.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err == nil {
		GlobalLogger.SetOutput(file)
	} else {
		GlobalLogger.SetOutput(os.Stdout)
		GlobalLogger.Warn("Failed to log to file, using default stderr")
	}

	// Add a hook to always echo warnings to the console in orange so they are visible
	GlobalLogger.AddHook(&warnConsoleHook{})
}

// warnConsoleHook duplicates warning messages to stderr with an orange color.
// This allows visibility of warnings even when primary log output is a file.
type warnConsoleHook struct{}

func (h *warnConsoleHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel}
}

func (h *warnConsoleHook) Fire(entry *logrus.Entry) error {
	color := "\x1b[33m" // fallback yellow
	reset := "\x1b[0m"
	// Prefer a 256-color orange if the terminal likely supports it.
	if supports256Color() {
		color = "\x1b[38;5;208m" // orange
	}

	// Reconstruct a simple field string (key=value) if present.
	var fieldParts []string
	for k, v := range entry.Data {
		fieldParts = append(fieldParts, k+"="+toString(v))
	}
	fields := ""
	if len(fieldParts) > 0 {
		fields = " (" + strings.Join(fieldParts, ", ") + ")"
	}

	// Write to stderr directly.
	fmtStr := color + "WARNING: " + entry.Message + fields + reset + "\n"
	_, _ = os.Stderr.WriteString(fmtStr)
	return nil
}

func supports256Color() bool {
	// Basic heuristic for 256-color support.
	term := os.Getenv("TERM")
	if strings.Contains(term, "256color") || strings.Contains(os.Getenv("COLORTERM"), "truecolor") {
		return true
	}
	return false
}

func toString(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", v)
}
