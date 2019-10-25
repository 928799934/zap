package zap

import (
	"encoding/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
)

var (
	logger *zap.Logger
	cleans []*clean
	files  []*writer
)

// Config ...
type Config struct {
	zap.Config
	Outputs        []string `json:"outputs" yaml:"outputs"`
	RetentionHours int      `json:"retentionHours" yaml:"retentionHours"`
}

// Logger zap.Logger
type Logger = zap.Logger

// LoadConfiguration ...
func LoadConfiguration(filename string) *Logger {
	jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
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
	// 生成logger
	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}
	// 替换logger输出
	logger = logger.WithOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return zapcore.NewCore(encoder, open(cfg.Outputs, cfg.RetentionHours), cfg.Level)
	}))
	return logger
}

// Close ...
func Close() {
	logger.Sync()
	for _, f := range files {
		f.close()
	}
	for _, c := range cleans {
		c.Close()
	}
}

// open 打开文件 启动定时清理
func open(paths []string, retentionHours int) zapcore.WriteSyncer {
	writers := make([]zapcore.WriteSyncer, 0, len(paths))
	for _, path := range paths {
		switch path {
		case "stdout":
			writers = append(writers, os.Stdout)
			// Don't close standard out.
			continue
		case "stderr":
			writers = append(writers, os.Stderr)
			// Don't close standard error.
			continue
		}
		f := newWriter(path)
		files = append(files, f)
		writers = append(writers, f)
		c := newClean(path, retentionHours)
		cleans = append(cleans, c)
	}
	return zap.CombineWriteSyncers(writers...)
}
