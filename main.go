package main

import (
	"github.com/danthegoodman1/FlyMachineAutoscaler/config"
	"github.com/danthegoodman1/FlyMachineAutoscaler/gologger"
	"github.com/danthegoodman1/FlyMachineAutoscaler/utils"
	"github.com/joho/godotenv"
	_ "gopkg.in/yaml.v3"
	"os"
	"os/signal"
	"syscall"
)

var (
	logger = gologger.NewLogger()
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load env file")
	}
	logger.Info().Msg("Starting Fly Machine Autoscaler")

	f, err := os.ReadFile(utils.Env_ConfigFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error reading config file")
	}

	err = config.LoadConfig(f)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error parsing config")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Warn().Msg("received shutdown signal!")
}
