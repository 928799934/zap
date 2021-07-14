package zap

import "go.uber.org/zap"

// Config ...
type Config struct {
	zap.Config
	Outputs        []string `json:"outputs" yaml:"outputs"`
	RetentionHours int      `json:"retentionHours" yaml:"retentionHours"`
}

// Logger zap.Logger
type Logger = zap.Logger
