package main

import (
	"github.com/danthegoodman1/FlyMachineAutoscaler/gologger"
	"github.com/joho/godotenv"
	_ "gopkg.in/yaml.v3"
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
}
