/*
 Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

// Package log output logged entries to respective logging hooks
package log

import (
	"bytes"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	logger LoggingInterface

	logModule = flag.String("log-module",
		"file",
		"Flag enable one of available logging module (file, console)")
	logLevel = flag.String("log-level",
		"info",
		"Set logging level (debug, info, error, warning, fatal)")
	logFileDir = flag.String("log-file-dir",
		defaultLogDir,
		"The flag to specify logging directory. The flag is only supported if logging module is file")
)

type key string

const (
	defaultLogDir   = "/var/log/huawei-cosi"
	timestampFormat = "2006-01-02 15:04:05.000000"

	cosiRequestID      key = "cosi.requestid"
	cosiChainRequestID     = "cosi-chain-requestid"
	requestID              = "requestID"
)

// LoggingInterface is an interface exposes logging functionality
type LoggingInterface interface {
	Logger

	flushable

	closable

	AddContext(ctx context.Context) Logger
}

// GetCosiRequestID get cosiRequestID
func GetCosiRequestID() key {
	return cosiRequestID
}

// Closable is an interface for closing logging streams.
// The interface should be implemented by hooks.
type closable interface {
	close()
}

// Flushable is an interface to commit current content of logging stream
type flushable interface {
	flush()
}

// Logger exposes logging functionality
type Logger interface {
	Debugf(format string, args ...interface{})

	Debugln(args ...interface{})

	Infof(format string, args ...interface{})

	Infoln(args ...interface{})

	Warningf(format string, args ...interface{})

	Warningln(args ...interface{})

	Errorf(format string, args ...interface{})

	Errorln(args ...interface{})

	Fatalf(format string, args ...interface{})

	Fatalln(args ...interface{})

	AddField(field string, value interface{}) Logger
}

type loggerImpl struct {
	*logrus.Entry
	hooks     []logrus.Hook
	formatter logrus.Formatter
}

var _ LoggingInterface = &loggerImpl{}

func parseLogLevel() (logrus.Level, error) {
	switch *logLevel {
	case "debug":
		return logrus.DebugLevel, nil
	case "info":
		return logrus.InfoLevel, nil
	case "warning":
		return logrus.WarnLevel, nil
	case "error":
		return logrus.ErrorLevel, nil
	case "fatal":
		return logrus.FatalLevel, nil
	default:
		return logrus.FatalLevel, fmt.Errorf("invalid logging level [%v]", logLevel)
	}
}

// Default init function for log module
func init() {
	*logModule = "console"
	err := InitLogging("dummy-name")
	if err != nil {
		logrus.Fatalf("Failed to initialize logging module")
	}
}

// InitLogging configures logging. Logs are written to a log file or stdout/stderr.
// Since logrus doesn't support multiple writers, each log stream is implemented as a hook.
func InitLogging(logName string) error {
	var tmpLogger loggerImpl
	tmpLogger.Entry = new(logrus.Entry)

	// initialize logrus in wrapper
	tmpLogger.Logger = logrus.New()

	// No output except for the hooks
	tmpLogger.Logger.SetOutput(ioutil.Discard)

	// set logging level
	level, err := parseLogLevel()
	if err != nil {
		return err
	}
	tmpLogger.Logger.SetLevel(level)

	// initialize log formatter
	formatter := &PlainTextFormatter{TimestampFormat: timestampFormat, pid: os.Getpid()}

	hooks := make([]logrus.Hook, 0)
	switch *logModule {
	case "file":
		logFilePath := fmt.Sprintf("%s/%s", *logFileDir, logName)
		// Write to the log file
		logFileHook, err := newFileHook(logFilePath, formatter)
		if err != nil {
			return fmt.Errorf("could not initialize logging to file: %v", err)
		}
		hooks = append(hooks, logFileHook)
	case "console":
		// Write to stdout/stderr
		logConsoleHook, err := newConsoleHook(formatter)
		if err != nil {
			return fmt.Errorf("could not initialize logging to console: %v", err)
		}
		hooks = append(hooks, logConsoleHook)
	default:
		return fmt.Errorf("invalid logging module [%v]. Support only 'file' or 'console'", logModule)
	}

	tmpLogger.hooks = hooks
	for _, hook := range tmpLogger.hooks {
		// initialize logrus with hooks
		tmpLogger.Logger.AddHook(hook)
	}

	logger = &tmpLogger
	return nil
}

// PlainTextFormatter is a formatter to ensure formatted logging output
type PlainTextFormatter struct {
	// TimestampFormat to use for display when a full timestamp is printed
	TimestampFormat string

	// process identity number
	pid int
}

var _ logrus.Formatter = &PlainTextFormatter{}

// Format ensure unified and formatted logging output
func (f *PlainTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := entry.Buffer
	if entry.Buffer == nil {
		b = &bytes.Buffer{}
	}

	_, _ = fmt.Fprintf(b, "%s %d", entry.Time.Format(f.TimestampFormat), f.pid)
	if len(entry.Data) != 0 {
		for key, value := range entry.Data {
			_, _ = fmt.Fprintf(b, "[%s:%v] ", key, value)
		}
	}

	_, _ = fmt.Fprintf(b, "%s %s\n", getLogLevel(entry.Level), entry.Message)

	return b.Bytes(), nil
}

func getLogLevel(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "[DEBUG]: "
	case logrus.InfoLevel:
		return "[INFO]: "
	case logrus.WarnLevel:
		return "[WARNING]: "
	case logrus.ErrorLevel:
		return "[ERROR]: "
	case logrus.FatalLevel:
		return "[FATAL]: "
	default:
		return "[UNKNOWN]: "
	}
}

// Debugf ensures output of formatted debug logs
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

// Debugln ensures output of Debug logs
func Debugln(args ...interface{}) {
	logger.Debugln(args...)
}

// Infof ensures output of formatted info logs
func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// Infoln ensures output of info logs
func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

// Warningf ensures output of formatted warning logs
func Warningf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

// Warningln ensures output of warning logs
func Warningln(args ...interface{}) {
	logger.Warningln(args...)
}

// Errorf ensures output of formatted error logs
func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Errorln ensures output of error logs
func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

// Fatalf ensures output of formatted fatal logs
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

// Fatalln ensures output of fatal logs
func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
}

// AddContext ensures appending context info in log
func AddContext(ctx context.Context) Logger {
	return logger.AddContext(ctx)
}

// AddField add value into field
func AddField(field string, value interface{}) Logger {
	return logger.AddField(field, value)
}

func (logger *loggerImpl) flush() {
	for _, hook := range logger.hooks {
		flushable, ok := hook.(flushable)
		if ok {
			flushable.flush()
		}
	}
}

func (logger *loggerImpl) close() {
	for _, hook := range logger.hooks {
		flushable, ok := hook.(closable)
		if ok {
			flushable.close()
		}
	}
}

// AddContext ensures appending context info in log
func (logger *loggerImpl) AddContext(ctx context.Context) Logger {
	if ctx.Value(cosiRequestID) == nil {
		return logger
	}
	return logger.AddField(requestID, ctx.Value(cosiRequestID))
}

// AddField ensures appending field info in log
func (logger *loggerImpl) AddField(field string, value interface{}) Logger {
	entry := logger.WithFields(logrus.Fields{
		field: value,
	})
	return &loggerImpl{
		Entry:     entry,
		hooks:     logger.hooks,
		formatter: logger.formatter,
	}
}

// EnsureGRPCContext ensures adding request id in incoming unary grpc context
func EnsureGRPCContext(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	newCtx, err := HandleRequestId(ctx)
	if err != nil {
		return handler(ctx, req)
	}

	return handler(newCtx, req)
}

type serverStreamWithContext struct {
	grpc.ServerStream
	ctx context.Context
}

// Context implement context func of serverStreamWithContext
func (ss serverStreamWithContext) Context() context.Context {
	return ss.ctx
}

// NewServerStreamWithContext returns a new serverStreamWithContext
func NewServerStreamWithContext(stream grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	return serverStreamWithContext{
		ServerStream: stream,
		ctx:          ctx,
	}
}

// EnsureStreamGRPCContext ensures adding request id in incoming stream grpc context
func EnsureStreamGRPCContext(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) (err error) {
	ctx := stream.Context()
	newCtx, err := HandleRequestId(ctx)
	if err != nil {
		return handler(srv, NewServerStreamWithContext(stream, ctx))
	}

	return handler(srv, NewServerStreamWithContext(stream, newCtx))
}

// SetRequestInfo is used to set the context with requestID value
func SetRequestInfo(ctx context.Context) (context.Context, error) {
	randomID, err := rand.Prime(rand.Reader, 32)
	if err != nil {
		Errorf("Failed in random ID generation for GRPC request ID logging: [%v]", err)
		return ctx, err
	}

	// the requestID value in metadata can be transferred between services via grpc
	ctx = metadata.AppendToOutgoingContext(ctx, cosiChainRequestID, randomID.String())
	return context.WithValue(ctx, cosiRequestID, randomID.String()), nil
}

// HandleRequestId is used to handle the requestId when the context is transferred between services via grpc
func HandleRequestId(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	// if ctx metadata not exist, generate one with metadata and value
	if !ok {
		Debugln("ctx not include metadata info, generate a new ctx with metadata and value")
		return SetRequestInfo(ctx)
	}

	// If ctx metadata exist and metadata includes cosiRequestId info,
	// then return ctx with value and append requestId to metadata again.
	// When the service acts as a new client and connects to other server,
	// the requestID information needs to be added to metadata again.
	// So, the requestId can be transferred between multiple services.
	if reqIDs, ok := md[cosiChainRequestID]; ok && len(reqIDs) == 1 {
		ctx = metadata.AppendToOutgoingContext(ctx, cosiChainRequestID, reqIDs[0])
		return context.WithValue(ctx, cosiRequestID, reqIDs[0]), nil
	}

	// if ctx metadata exist, but metadata not include requestId info, generate one with metadata and value
	Debugln("ctx metadata not include requestId info, generate a new ctx with metadata and value")
	return SetRequestInfo(ctx)
}

// Flush ensures to commit current content of logging stream
func Flush() {
	logger.flush()
}

// Close ensures closing output stream
func Close() {
	logger.close()
}
