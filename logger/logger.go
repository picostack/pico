package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Env string

type cfg struct {
	Environment Env           `envconfig:"ENVIRONMENT" default:"production"`
	LogLevel    zapcore.Level `envconfig:"LOG_LEVEL"   default:"info"`
}

const (
	EnvironmentProd Env = "production"
	EnvironmentDev  Env = "development"
)

func (e *Env) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case string(EnvironmentProd):
		*e = EnvironmentProd
	case string(EnvironmentDev):
		*e = EnvironmentDev
	default:
		return errors.Errorf("unknown environment type: '%s'", string(text))
	}
	return nil
}

func init() {
	godotenv.Load("../.env", ".env") //nolint:errcheck

	var c cfg
	envconfig.MustProcess("", &c)

	var config zap.Config
	if c.Environment == EnvironmentDev || isInTests() {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level.SetLevel(c.LogLevel)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		fmt.Println("Error during logging config:", err)
		os.Exit(1)
	}
	zap.ReplaceGlobals(logger)

	zap.L().Info("logger configured",
		zap.String("level", c.LogLevel.String()),
		zap.String("env", string(c.Environment)))

}

func isInTests() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.v=") {
			return true
		}
	}
	return false
}
