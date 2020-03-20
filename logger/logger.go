package logger

import (
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	// constructs a logger and replaces the default global logger
	var config zap.Config
	if d, e := strconv.ParseBool(os.Getenv("DEVELOPMENT")); d && e == nil || isInTests() {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	if d, e := strconv.ParseBool(os.Getenv("DEBUG")); d && e == nil || isInTests() {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func isInTests() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.v=") {
			return true
		}
	}
	return false
}
