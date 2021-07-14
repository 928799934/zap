package zap

import (
	"encoding/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
)

var (
	core   *Logger
	cleans []*clean
	files  []*writer
)

// LoadConfiguration ...
func LoadConfiguration(filename string) {
	jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	LoadConfigurationByContent(jsonData)
}

// GetLogger ...
func GetLogger() *Logger {
	return core
}

// LoadConfigurationByContent ...
func LoadConfigurationByContent(jsonData []byte) {
	cfg := &Config{}
	if err := json.Unmarshal(jsonData, cfg); err != nil {
		panic(err)
	}
	var encoder zapcore.Encoder
	switch cfg.Encoding {
	case "json":
		encoder = zapcore.NewJSONEncoder(cfg.EncoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	default:
		panic("encoding error is " + cfg.Encoding)
	}
	zapcore.NewMultiWriteSyncer()
	// 生成logger
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	// 替换logger输出
	core = logger.WithOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return zapcore.NewCore(encoder, func() zapcore.WriteSyncer {
			writeSyncerList := make([]zapcore.WriteSyncer, 0, len(cfg.Outputs))
			for _, path := range cfg.Outputs {
				switch path {
				case "stdout":
					writeSyncerList = append(writeSyncerList, os.Stdout)
				case "stderr":
					writeSyncerList = append(writeSyncerList, os.Stderr)
				default:
					f := newWriter(path)
					files = append(files, f)
					writeSyncerList = append(writeSyncerList, f)
					// 增加清理队列
					cleans = append(cleans, newClean(path, cfg.RetentionHours))
				}
			}
			return zap.CombineWriteSyncers(writeSyncerList...)
		}(), cfg.Level)
	}))
}

// Close ...
func Close() {
	core.Sync()
	for _, f := range files {
		f.close()
	}
	for _, c := range cleans {
		c.close()
	}
}
