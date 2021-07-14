package zap

import (
	"encoding/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
)

var (
	core   *zap.Logger
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
	return LoadConfigurationByContent(jsonData)
}

// LoadConfigurationByContent ...
func LoadConfigurationByContent(jsonData []byte) *Logger {
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
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	// 替换logger输出
	core = logger.WithOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return zapcore.NewCore(encoder, open(cfg.Outputs, cfg.RetentionHours), cfg.Level)
	}))
	return core
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

// open 打开文件 启动定时清理
func open(paths []string, retentionHours int) zapcore.WriteSyncer {
	writeSyncerList := make([]zapcore.WriteSyncer, 0, len(paths))
	for _, path := range paths {
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
			cleans = append(cleans, newClean(path, retentionHours))
		}
	}
	return zap.CombineWriteSyncers(writeSyncerList...)
}
