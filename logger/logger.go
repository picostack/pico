package logger

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	godotenv.Load("../.env", ".env") //nolint:errcheck

	prod, err := strconv.ParseBool(os.Getenv("PRODUCTION"))
	if _, ok := err.(*strconv.NumError); !ok {
		fmt.Println("Error during logging config:", err)
		os.Exit(1)
	}

	var config zap.Config
	if prod && !isInTests() {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(os.Getenv("LOG_LEVEL"))); err != nil {
		fmt.Println("Error during logging config:", err)
		os.Exit(1)
	}

	config.Level.SetLevel(level)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		fmt.Println("Error during logging config:", err)
		os.Exit(1)
	}
	zap.ReplaceGlobals(logger)

	if !prod {
		zap.L().Info("logger configured in development mode", zap.String("level", level.String()))
	}
}

func isInTests() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.v=") {
			return true
		}
	}
	return false
}
