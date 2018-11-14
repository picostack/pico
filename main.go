package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Southclaws/wadsworth/service"
)

var (
	version = "master"
)

func init() {
	// constructs a logger and replaces the default global logger
	var config zap.Config
	if d, e := strconv.ParseBool(os.Getenv("DEVELOPMENT")); d && e == nil {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	if d, e := strconv.ParseBool(os.Getenv("DEBUG")); d && e == nil {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func main() {
	app := cli.NewApp()

	app.Name = "wadsworth"
	app.Usage = "A git-driven task automation butler."
	app.UsageText = `TODO: Short usage info`
	app.Version = version
	app.Description = `TODO: Longform description`
	app.Author = "Southclaws"
	app.Email = "hello@southcla.ws"

	app.Commands = []cli.Command{
		{
			Name:    "run",
			Aliases: []string{"r"},
			Description: `Starts the Wadsworth daemon with the specified target repository. This
repository should contain one or more configuration files for Wadsworth. When
this repository has new commits, Wadsworth will automatically reconfigure.`,
			Usage:     "argument `target` specifies Git repository for configuration.",
			ArgsUsage: "target",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "directory", EnvVar: "DIRECTORY", Value: "./cache/"},
				cli.DurationFlag{Name: "check-interval", EnvVar: "CHECK_INTERVAL", Value: time.Second * 5},
			},
			Action: func(c *cli.Context) (err error) {
				if !c.Args().Present() {
					cli.ShowCommandHelp(c, "run")
					return errors.New("missing argument: target")
				}

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				svc, err := service.Initialise(ctx, service.Config{
					Target:        c.Args().First(),
					Directory:     c.String("directory"),
					CheckInterval: c.Duration("check-interval"),
				})
				if err != nil {
					return errors.Wrap(err, "failed to initialise")
				}

				zap.L().Info("service initialised")

				errs := make(chan error, 1)
				go func() { errs <- svc.Start() }()

				s := make(chan os.Signal, 1)
				signal.Notify(s, os.Interrupt)

				select {
				case sig := <-s:
					err = errors.New(sig.String())
				case err = <-errs:
				}

				return
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		zap.L().Fatal("exit", zap.Error(err))
	}
}
