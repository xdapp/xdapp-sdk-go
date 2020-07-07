package register

import (
	"sort"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	StdErrLogOutput = "stderr"
	StdOutLogOutput = "stdout"
)

type defaultLogger struct {
	*zap.Logger
}

type logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

func (cfg *Config) setupLogging() error {
	logCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      make([]string, 0),
		ErrorOutputPaths: make([]string, 0),
	}

	outputPaths, errOutputPaths := make(map[string]struct{}), make(map[string]struct{})
	for _, v := range cfg.LogOutputs {
		switch v {
		case StdErrLogOutput:
			outputPaths[StdErrLogOutput] = struct{}{}
			errOutputPaths[StdErrLogOutput] = struct{}{}

		case StdOutLogOutput:
			outputPaths[StdOutLogOutput] = struct{}{}
			errOutputPaths[StdOutLogOutput] = struct{}{}

		default:
			outputPaths[v] = struct{}{}
			errOutputPaths[v] = struct{}{}
		}
	}

	for v := range outputPaths {
		logCfg.OutputPaths = append(logCfg.OutputPaths, v)
	}
	for v := range errOutputPaths {
		logCfg.ErrorOutputPaths = append(logCfg.ErrorOutputPaths, v)
	}
	sort.Strings(logCfg.OutputPaths)
	sort.Strings(logCfg.ErrorOutputPaths)

	if cfg.Debug {
		logCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		logCfg.Encoding = "console"
		grpc.EnableTracing = true
	}

	var err error
	cfg.logger, err = logCfg.Build()
	if err != nil {
		return err
	}

	cfg.loggerMu = new(sync.RWMutex)

	cfg.loggerConfig = &logCfg

	return nil
}

func (cfg *Config) Logger() logger {
	cfg.loggerMu.RLock()
	l := cfg.logger
	cfg.loggerMu.RUnlock()
	return NewDefaultLogger(l)
}

func NewDefaultLogger(lg *zap.Logger) *defaultLogger {
	return &defaultLogger{
		lg,
	}
}

func (lg *defaultLogger) Info(msg string) {
	lg.Logger.Info(msg)
}
func (lg *defaultLogger) Debug(msg string) {
	lg.Logger.Debug(msg)
}
func (lg *defaultLogger) Warn(msg string) {
	lg.Logger.Warn(msg)
}
func (lg *defaultLogger) Error(msg string) {
	lg.Logger.Error(msg)
}

