package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	keysToMarshal = map[string]bool{"ctx": true, "meta": true}
)

type zapLogger struct {
	sugar *zap.SugaredLogger
}

func initializeLoggerWithZapLogger(config *LoggerConfig) {
	loggerConfig := getZapLoggerConfig(config)

	var encoder zapcore.Encoder
	var isJSONEncDisabled bool

	// Check if config is provided and use it, otherwise fallback to OS environment variable
	if config != nil {
		isJSONEncDisabled = config.JsonEncoderDisabled
	} else {
		isJSONEncDisabledStr := os.Getenv("LOGGER_JSON_ENCODER_DISABLED")
		isJSONEncDisabled, _ = strconv.ParseBool(isJSONEncDisabledStr)
	}

	if !isJSONEncDisabled {
		// Create a JSON encoder for file logging
		encoder = zapcore.NewJSONEncoder(loggerConfig.EncoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(loggerConfig.EncoderConfig)
	}

	writerSyncers := make([]zapcore.WriteSyncer, 0)

	var isConsoleSyncerDisabled bool

	// Check if config is provided and use it, otherwise fallback to OS environment variable
	if config != nil {
		isConsoleSyncerDisabled = config.ConsoleSyncerDisabled
	} else {
		isConsoleSyncerDisabledStr := os.Getenv("LOGGER_CONSOLE_SYNCER_DISABLED")
		isConsoleSyncerDisabled, _ = strconv.ParseBool(isConsoleSyncerDisabledStr)
	}
	if !isConsoleSyncerDisabled {
		// Create a zapcore.WriteSyncer for console logging
		writerSyncers = append(writerSyncers, zapcore.AddSync(os.Stdout))
	}

	var isFileSyncerDisabled bool

	// Check if config is provided and use it, otherwise fallback to OS environment variable
	if config != nil {
		isFileSyncerDisabled = config.FileSyncerDisabled
	} else {
		isFileSyncerDisabledStr := os.Getenv("LOGGER_FILE_SYNCER_DISABLED")
		isFileSyncerDisabled, _ = strconv.ParseBool(isFileSyncerDisabledStr)
	}
	if !isFileSyncerDisabled {
		// Create a lumberjack logger for log file rolling
		var logFileName string
		var logFilePath string
		if config != nil {
			logFilePath = config.FileSyncerPath
		} else {
			logFileName = fmt.Sprintf("logs/logs_%s.log", time.Now().Format("2006-01-02"))
		}

		var logFileSize int
		if config != nil {
			logFileSize = config.FileSyncerMaxSize
		} else {
			logFileSize = 10
		}

		var logFileBackups int
		if config != nil {
			logFileBackups = config.FileSyncerMaxBackups
		} else {
			logFileBackups = 10
		}

		var logFileMaxAge int
		if config != nil {
			logFileMaxAge = config.FileSyncerMaxAge
		} else {
			logFileMaxAge = 1
		}

		var logFileCompress bool
		if config != nil {
			logFileCompress = config.FileSyncerCompress
		} else {
			logFileCompress = false
		}

		logFinalFilePath := filepath.Join(logFilePath, logFileName)
		lumberjackLogger := &lumberjack.Logger{
			Filename:   logFinalFilePath,
			MaxSize:    logFileSize,     // Max size in megabytes before log rotation occurs
			MaxBackups: logFileBackups,  // Max number of old log files to retain
			MaxAge:     logFileMaxAge,   // Max number of days to retain old log files
			Compress:   logFileCompress, // Whether to compress the old log files
			LocalTime:  true,            // Use the local time zone for log rotation
		}
		writerSyncers = append(writerSyncers, zapcore.AddSync(lumberjackLogger))
	}

	// Create a zapcore.WriteSyncer for both console and file logging
	writeSyncer := zapcore.NewMultiWriteSyncer(writerSyncers...)

	// Create a zapcore.Core with the encoders and write syncer
	core := zapcore.NewCore(encoder, writeSyncer, loggerConfig.Level)
	// Create a new logger with the core
	zapLog := zap.New(core, zap.AddCallerSkip(1))

	defer func(zapLogger *zap.Logger) {
		err := zapLogger.Sync()
		if err != nil {
			fmt.Println("Could not sync zap logger")
		}
	}(zapLog)

	Logger = &zapLogger{sugar: zapLog.Sugar()}

	var isSocketLoggingEnabled bool

	// Check if config is provided and use it, otherwise fallback to OS environment variable
	if config != nil {
		isSocketLoggingEnabled = config.SocketLoggingEnabled
	} else {
		isSocketLoggingEnabledStr := os.Getenv("LOGGER_SOCKET_LOGGING_ENABLED")
		isSocketLoggingEnabled, _ = strconv.ParseBool(isSocketLoggingEnabledStr)
	}

	if isSocketLoggingEnabled {
		reInitializeLogger(config)
	}
}

func (z *zapLogger) Write(p []byte) (n int, err error) {
	z.Debug(string(p))
	return len(p), nil
}

func (z *zapLogger) Debug(message string, fields ...interface{}) {
	preprocessLog(fields)
	z.sugar.Debugw(message, fields...)
}

func (z *zapLogger) Infof(message string, fields ...interface{}) {
	preprocessLog(fields)
	z.sugar.Infof(message, fields...)
}

func (z *zapLogger) Info(message string, fields ...interface{}) {
	preprocessLog(fields)
	z.sugar.Infow(message, fields...)
}

func (z *zapLogger) Warn(message string, fields ...interface{}) {
	preprocessLog(fields)
	z.sugar.Warnw(message, fields...)
}

func (z *zapLogger) Error(message string, fields ...interface{}) {
	preprocessLog(fields)
	z.sugar.Errorw(message, fields...)
}

func (z *zapLogger) Fatal(message string, fields ...interface{}) {
	preprocessLog(fields)
	z.sugar.Fatalw(message, fields...)
}

func preprocessLog(fields []interface{}) {
	noOfFields := len(fields)
	for i := 1; i < noOfFields; i += 2 {
		field, ok := fields[i-1].(string)
		if !ok {
			continue
		}
		if keysToMarshal[field] {
			marshalledJSON, _ := json.Marshal(fields[i])
			fields[i] = string(marshalledJSON)
		}
	}
}

// GetHostname returns the hostname.
func GetHostname() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		fmt.Println("Failed to extract hostname from OS", zap.Error(err))
		return "unknown", err
	}
	return host, nil
}

func getZapLoggerConfig(config *LoggerConfig) zap.Config {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.DisableCaller = true
	loggerConfig.Sampling = nil
	loggerConfig.OutputPaths = []string{"stdout"}
	loggerConfig.EncoderConfig.EncodeTime = syslogTimeEncoder
	osHostname, _ := GetHostname()
	var serviceName string
	if config != nil {
		serviceName = config.ServiceName
	} else {
		serviceName = os.Getenv("SERVICE")
	}
	initialFields := map[string]interface{}{
		"svc":  serviceName,
		"host": osHostname,
	}

	loggerConfig.Level = getLoggerMode(config)
	loggerConfig.InitialFields = initialFields
	return loggerConfig
}

func syslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000Z07:00"))
}

type SocketSyncer struct {
	config *LoggerConfig
	client net.Conn
}

// NewSocketSyncer create a socket logger push the logs in socket
func NewSocketSyncer(config *LoggerConfig) zapcore.WriteSyncer {
	c, err := net.Dial("tcp", net.JoinHostPort(os.Getenv("LOGGER_SOCKET_ADDRESS"), os.Getenv("LOGGER_SOCKET_PORT")))
	if err != nil {
		fmt.Println("failed to initialize socket logger", err.Error())
		return nil
	}

	hs := &SocketSyncer{client: c, config: config}
	ws := zapcore.Lock(hs)
	return ws
}

func (w SocketSyncer) Sync() error {
	return nil
}

func (w SocketSyncer) Write(p []byte) (int, error) {
	var err error
	var socketTimeout int
	if w.config != nil {
		socketTimeout = w.config.SocketTimeout
	} else {
		socketTimeout, err = strconv.Atoi(os.Getenv("LOGGER_SOCKET_TIMEOUT"))
		if err != nil {
			fmt.Println("Failed to get timeout from env", err.Error())
			socketTimeout = 10
		}
	}
	err = w.client.SetWriteDeadline(time.Now().Add(time.Duration(socketTimeout) * time.Millisecond))
	if err != nil {
		fmt.Println("Failed to set deadline", err.Error())
	}
	cnt, err := w.client.Write(p)

	if err != nil {
		cnt, _ = fmt.Print(string(p))
		if errors.Is(err, syscall.EPIPE) {
			reInitializeLogger(w.config)
		}
		return cnt, nil
	}
	return cnt, err
}

func reInitializeLogger(config *LoggerConfig) {
	sink, errSink, err := openSink()
	if err != nil {
		fmt.Println("sink open error", err.Error())
		return
	}
	opts := buildOptions(config, errSink)
	socketWriteSyncer := NewSocketSyncer(config)
	if socketWriteSyncer == nil {
		return
	}
	loggerConfig := getZapLoggerConfig(config)
	jsonEncoder := zapcore.NewJSONEncoder(loggerConfig.EncoderConfig)
	var core zapcore.Core
	var isConsoleSyncerDisabled bool

	// Check if config is provided and use it, otherwise fallback to OS environment variable
	if config != nil {
		isConsoleSyncerDisabled = config.ConsoleSyncerDisabled
	} else {
		isConsoleSyncerDisabledStr := os.Getenv("LOGGER_CONSOLE_SYNCER_DISABLED")
		isConsoleSyncerDisabled, _ = strconv.ParseBool(isConsoleSyncerDisabledStr)
	}
	if isConsoleSyncerDisabled {
		core = zapcore.NewCore(jsonEncoder, socketWriteSyncer, loggerConfig.Level)
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(jsonEncoder, socketWriteSyncer, loggerConfig.Level),
			zapcore.NewCore(jsonEncoder, sink, loggerConfig.Level),
		)
	}
	zapLog := zap.New(core, opts...)
	defer func(zapLogger *zap.Logger) {
		err := zapLogger.Sync()
		if err != nil {
			fmt.Println("Count not sync zap logger")
		}
	}(zapLog)
	Logger = &zapLogger{sugar: zapLog.Sugar()}
}

func openSink() (sink, errSink zapcore.WriteSyncer, err error) {
	sink, closeOut, err := zap.Open([]string{"stdout"}...)
	if err != nil {
		return nil, nil, err
	}
	errSink, _, err = zap.Open([]string{"stderr"}...)
	if err != nil {
		closeOut()
		return nil, nil, err
	}
	return sink, errSink, nil
}

func buildOptions(config *LoggerConfig, errSink zapcore.WriteSyncer) []zap.Option {
	stackLevel := zap.ErrorLevel
	opts := []zap.Option{zap.ErrorOutput(errSink)}
	opts = append(opts, zap.AddCallerSkip(1), zap.AddStacktrace(stackLevel))
	osHostname, _ := GetHostname()

	var serviceName string

	if config != nil {
		serviceName = config.ServiceName
	} else {
		serviceName = os.Getenv("SERVICE")
	}

	fs := []zap.Field{zap.Any("host", osHostname), zap.Any("svc", serviceName)}
	opts = append(opts, zap.Fields(fs...))
	return opts
}

func getLoggerMode(config *LoggerConfig) zap.AtomicLevel {
	var loggingMode string
	if config != nil {
		loggingMode = config.LogMode
	} else {
		loggingMode = os.Getenv("LOGGER_MODE")
	}
	if loggingMode == "DEBUG" {
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	} else if loggingMode == "ERROR" {
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	}
	return zap.NewAtomicLevelAt(zap.InfoLevel)
}
