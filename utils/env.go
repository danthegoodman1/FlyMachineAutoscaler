package utils

import "os"

var (
	Env_FlyAPIToken = os.Getenv("FLY_API_TOKEN")
	Env_FlyOrg      = GetEnvOrDefault("FLY_ORG", "personal")
	Env_ConfigFile  = GetEnvOrDefault("CONFIG", "./config.yml")
)
