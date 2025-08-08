package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogLevel represents the severity of a log message
type LogLevel int

// Log levels
const (
	DebugLevel LogLevel = iota
	InfoLevel
	ErrorLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a logger instance
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	errorLogger *log.Logger
	verbose     bool
}

// defaultLogger is the package-level logger instance
var defaultLogger *Logger

// checkDirWritable checks if a directory exists and is writable
func checkDirWritable(dir string) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %w", err)
		}
		return fmt.Errorf("error checking directory: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	// Check if we can write to it by creating a temporary file
	testFile := filepath.Join(dir, ".ipsec-vpn-write-test-"+time.Now().Format("20060102150405"))
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	file.Close()
	os.Remove(testFile) // Clean up

	return nil
}

// Init initializes the default logger
func Init(verbose bool) error {
	// Try to get log directory from config
	logDir := viper.GetString("log.directory")
	
	// If not specified in config, determine the best location
	if logDir == "" {
		// First try to use /var/log/ipsec-vpn (production default)
		logDir = "/var/log/ipsec-vpn"
		
		// Check if we can write to /var/log
		if err := checkDirWritable("/var/log"); err != nil {
			// Fall back to a local logs directory
			execDir, err := os.Getwd()
			if err == nil {
				logDir = filepath.Join(execDir, "logs")
				fmt.Fprintf(os.Stderr, "Using local log directory: %s\n", logDir)
			} else {
				// Last resort: use temp directory
				logDir = filepath.Join(os.TempDir(), "ipsec-vpn-logs")
				fmt.Fprintf(os.Stderr, "Using temporary log directory: %s\n", logDir)
			}
		}
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// If we can't create the directory, try using a temporary directory
		tmpLogDir := filepath.Join(os.TempDir(), "ipsec-vpn-logs")
		fmt.Fprintf(os.Stderr, "Failed to create log directory %s, falling back to %s\n", logDir, tmpLogDir)
		
		logDir = tmpLogDir
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Configure log rotation
	logFile := filepath.Join(logDir, "ipsec-vpn.log")
	rotatingLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    viper.GetInt("log.max_size"),     // megabytes
		MaxBackups: viper.GetInt("log.max_backups"),  // number of backups
		MaxAge:     viper.GetInt("log.max_age"),      // days
		Compress:   viper.GetBool("log.compress"),    // compress rotated files
	}

	// Set defaults if not specified in config
	if rotatingLogger.MaxSize == 0 {
		rotatingLogger.MaxSize = 10 // 10 MB
	}
	if rotatingLogger.MaxBackups == 0 {
		rotatingLogger.MaxBackups = 5
	}
	if rotatingLogger.MaxAge == 0 {
		rotatingLogger.MaxAge = 30 // 30 days
	}

	// Create multi-writer for console and file
	var debugWriter, infoWriter, errorWriter io.Writer

	// If verbose, write debug logs to both stdout and file
	if verbose {
		debugWriter = io.MultiWriter(os.Stdout, rotatingLogger)
		infoWriter = io.MultiWriter(os.Stdout, rotatingLogger)
		// Print log file location in verbose mode
		fmt.Printf("Logging to file: %s\n", logFile)
	} else {
		// In non-verbose mode, debug logs go only to file
		debugWriter = rotatingLogger
		infoWriter = io.MultiWriter(os.Stdout, rotatingLogger)
	}

	// Error logs always go to stderr and file
	errorWriter = io.MultiWriter(os.Stderr, rotatingLogger)

	// Create the logger
	defaultLogger = &Logger{
		debugLogger: log.New(debugWriter, "DEBUG: ", log.Ldate|log.Ltime),
		infoLogger:  log.New(infoWriter, "INFO: ", log.Ldate|log.Ltime),
		errorLogger: log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime),
		verbose:     verbose,
	}

	return nil
}

// New creates a new logger instance
func New(verbose bool) (*Logger, error) {
	logDir := viper.GetString("log.directory")
	if logDir == "" {
		// Default to /var/log/ipsec-vpn if not specified
		logDir = "/var/log/ipsec-vpn"
		
		// For development, use a local logs directory if /var/log is not writable
		if _, err := os.Stat("/var/log"); os.IsNotExist(err) || os.IsPermission(err) {
			execDir, err := os.Getwd()
			if err == nil {
				logDir = filepath.Join(execDir, "logs")
			}
		}
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Configure log rotation
	logFile := filepath.Join(logDir, "ipsec-vpn.log")
	rotatingLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    viper.GetInt("log.max_size"),     // megabytes
		MaxBackups: viper.GetInt("log.max_backups"),  // number of backups
		MaxAge:     viper.GetInt("log.max_age"),      // days
		Compress:   viper.GetBool("log.compress"),    // compress rotated files
	}

	// Set defaults if not specified in config
	if rotatingLogger.MaxSize == 0 {
		rotatingLogger.MaxSize = 10 // 10 MB
	}
	if rotatingLogger.MaxBackups == 0 {
		rotatingLogger.MaxBackups = 5
	}
	if rotatingLogger.MaxAge == 0 {
		rotatingLogger.MaxAge = 30 // 30 days
	}

	// Create multi-writer for console and file
	var debugWriter, infoWriter, errorWriter io.Writer

	// If verbose, write debug logs to both stdout and file
	if verbose {
		debugWriter = io.MultiWriter(os.Stdout, rotatingLogger)
		infoWriter = io.MultiWriter(os.Stdout, rotatingLogger)
	} else {
		// In non-verbose mode, debug logs go only to file
		debugWriter = rotatingLogger
		infoWriter = io.MultiWriter(os.Stdout, rotatingLogger)
	}

	// Error logs always go to stderr and file
	errorWriter = io.MultiWriter(os.Stderr, rotatingLogger)

	// Create the logger
	return &Logger{
		debugLogger: log.New(debugWriter, "DEBUG: ", log.Ldate|log.Ltime),
		infoLogger:  log.New(infoWriter, "INFO: ", log.Ldate|log.Ltime),
		errorLogger: log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime),
		verbose:     verbose,
	}, nil
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

// Debug logs a debug message using the default logger
func Debug(format string, v ...interface{}) {
	if defaultLogger == nil {
		// Initialize with default settings if not initialized
		if err := Init(false); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
			return
		}
	}
	defaultLogger.Debug(format, v...)
}

// Info logs an info message using the default logger
func Info(format string, v ...interface{}) {
	if defaultLogger == nil {
		// Initialize with default settings if not initialized
		if err := Init(false); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
			return
		}
	}
	defaultLogger.Info(format, v...)
}

// Error logs an error message using the default logger
func Error(format string, v ...interface{}) {
	if defaultLogger == nil {
		// Initialize with default settings if not initialized
		if err := Init(false); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
			return
		}
	}
	defaultLogger.Error(format, v...)
}

// SetVerbose sets the verbose mode for the default logger
func SetVerbose(verbose bool) {
	if defaultLogger != nil {
		defaultLogger.verbose = verbose
	}
}

// GetTimestamp returns a formatted timestamp for logging
func GetTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}