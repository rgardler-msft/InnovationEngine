package logging

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

const (
	DefaultLogFile  = "ie.log"
	maxLogSnapshots = 5
)

func Init(level Level, logPath string) {
	GlobalLogger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
		DisableQuote:  true,
	})

	GlobalLogger.SetReportCaller(false)
	GlobalLogger.SetLevel(level.Integer())

	writer, err := configureLogWriter(logPath)
	if err != nil {
		GlobalLogger.SetOutput(os.Stdout)
		GlobalLogger.Warnf("Failed to configure log file '%s', using stdout: %v", logPath, err)
	} else if writer != nil {
		GlobalLogger.SetOutput(writer)
	} else {
		GlobalLogger.SetOutput(os.Stdout)
	}

	// Add a hook to always echo warnings to the console in orange so they are visible
	GlobalLogger.AddHook(&warnConsoleHook{})
}

func configureLogWriter(logPath string) (*os.File, error) {
	path := strings.TrimSpace(logPath)
	if path == "" {
		return nil, nil
	}

	if err := ensureLogDirectory(path); err != nil {
		return nil, err
	}

	if err := rotateLogs(path, maxLogSnapshots); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o666)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func ensureLogDirectory(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func rotateLogs(basePath string, maxSnapshots int) error {
	if maxSnapshots <= 1 {
		return nil
	}

	oldest := fmt.Sprintf("%s.%d", basePath, maxSnapshots-1)
	if err := os.Remove(oldest); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	for i := maxSnapshots - 2; i >= 1; i-- {
		src := fmt.Sprintf("%s.%d", basePath, i)
		dst := fmt.Sprintf("%s.%d", basePath, i+1)
		if err := os.Rename(src, dst); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}
	}

	if err := os.Rename(basePath, fmt.Sprintf("%s.1", basePath)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	return nil
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
