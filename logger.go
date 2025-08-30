package logger

import (
	"sync"
)

// ILogger is the interface for the logger
type ILogger interface {
	Debug(message string, fields ...interface{})
	Infof(message string, fields ...interface{})
	Info(message string, fields ...interface{})
	Warn(message string, fields ...interface{})
	Error(message string, fields ...interface{})
	Fatal(message string, fields ...interface{})
	Write(p []byte) (n int, err error)
}

// LoggerConfig is the config for the logger
type LoggerConfig struct {
	ServiceName           string // to set the service name (default: "")
	LogMode               string // INFO, DEBUG, WARN, ERROR, FATAL (default: INFO)
	JsonEncoderDisabled   bool   // to disable the json encoding of logs (default: false)
	ConsoleSyncerDisabled bool   // to disable the std out based logging of logs (default: false)
	FileSyncerDisabled    bool   // to disable file based logging of logs (default: false)
	SocketLoggingEnabled  bool   // to enable socket logging of logs (default: false)
	SocketTimeout         int    // to set the timeout for the socket connection (default: 10)
	FileSyncerPath        string // to set the path of the file to be logged (default: "")
	FileSyncerMaxSize     int    // to set the max size of the file to be logged (default: 100)
	FileSyncerMaxBackups  int    // to set the max backups of the file to be logged (default: 10)
	FileSyncerMaxAge      int    // to set the max age of the file to be logged (default: 30)
	FileSyncerCompress    bool   // to set the compress of the file to be logged (default: false)
}

// NewDefaultLoggerConfig creates a new default logger config
func NewDefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		ServiceName:           "",
		LogMode:               "INFO",
		JsonEncoderDisabled:   false,
		ConsoleSyncerDisabled: false,
		FileSyncerDisabled:    false,
		SocketLoggingEnabled:  false,
		SocketTimeout:         10,
		FileSyncerPath:        "",
		FileSyncerMaxSize:     100,
		FileSyncerMaxBackups:  10,
		FileSyncerMaxAge:      30,
		FileSyncerCompress:    false,
	}
}

var (
	Logger ILogger
	once   sync.Once
)

// LoggerType is the type of logger
type LoggerType string

const (
	ZapLogger LoggerType = "zap"
)

// init initializes the logger with default config
func Init() {
	InitWithConfig(ZapLogger, NewDefaultLoggerConfig())
}

// init initializes the logger with config
func InitWithConfig(loggerType LoggerType, config *LoggerConfig) {
	if config == nil {
		config = NewDefaultLoggerConfig()
	}
	switch loggerType {
	case ZapLogger:
		once.Do(func() {
			initializeLoggerWithZapLogger(config)
		})
	default:
		Logger.Fatal("Invalid logger type", "loggerType", loggerType)
	}
}
