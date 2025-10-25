package app

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v2/pkg/logger"
	"gopkg.in/natefinch/lumberjack.v2"
)

// getAppLogDir returns a writable directory appropriate for the platform.
func getAppLogDir(appName string) string {
	// Prefer user config dir (cross-platform)
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, appName, "logs")
	}

	// Fallbacks per OS
	switch runtime.GOOS {
	case "darwin":
		if hd, err := os.UserHomeDir(); err == nil {
			return filepath.Join(hd, "Library", "Logs", appName)
		}
	case "windows":
		if ad := os.Getenv("APPDATA"); ad != "" {
			return filepath.Join(ad, appName, "logs")
		}
	default: // linux/unix fallback
		if hd, err := os.UserHomeDir(); err == nil {
			return filepath.Join(hd, ".config", appName, "logs")
		}
	}

	// Last resort: current working directory
	if wd, err := os.Getwd(); err == nil {
		return filepath.Join(wd, "logs")
	}
	// If everything fails, use temp dir
	return filepath.Join(os.TempDir(), appName, "logs")
}

// InitLogger write log to both log file and stdout
func InitLogger(appName string) (logger.Logger, error) {
	dir := getAppLogDir(appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir log dir: %w", err)
	}
	logPath := filepath.Join(dir, appName+".log")
	log.Println("Log path:", logPath)
	// Try to create/open the log file first to catch any errors early
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("cannot create log file: %w", err)
	}
	file.Close()

	// Test lumberjack logger by writing a dummy log entry
	f := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}
	if _, err := f.Write([]byte("app is started\n")); err != nil {
		return nil, fmt.Errorf("lumberjack logger cannot write: %w", err)
	}
	// 同时输出到 stdout（在 wails dev 时可见）和文件（生产时持久化）
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	wailsLogger = &WailsAdapter{Out: f}
	return wailsLogger, nil
}

func disableStdout() {
	log.SetOutput(wailsLogger.(*WailsAdapter).Out)
}

// WailsAdapter 实现 wails.Logger，把日志转到标准 log 或任意 logger
type WailsAdapter struct {
	Out io.WriteCloser
}

var wailsLogger logger.Logger

// Print writes a log message.
func (w *WailsAdapter) Print(message string) { log.Println(message) }

// Trace writes a trace-level log message.
func (w *WailsAdapter) Trace(message string) { log.Println("TRA |", message) }

// Debug writes a debug-level log message.
func (w *WailsAdapter) Debug(message string) { log.Println("DEB |", message) }

// Info writes an info-level log message.
func (w *WailsAdapter) Info(message string) { log.Println("INF |", message) }

// Warning writes a warning-level log message.
func (w *WailsAdapter) Warning(message string) { log.Println("WAR |", message) }

// Error writes an error-level log message.
func (w *WailsAdapter) Error(message string) { log.Println("ERR |", message) }

// Fatal writes a fatal-level log message and exits the application.
func (w *WailsAdapter) Fatal(message string) { log.Println("FAT |", message); os.Exit(1) }
