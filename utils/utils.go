package utils

import (
	"fmt"
	"github.com/danthegoodman1/FlyMachineAutoscaler/gologger"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/segmentio/ksuid"
	"os"
	"strconv"
)

var logger = gologger.NewLogger()

func GetEnvOrDefault(env, defaultVal string) string {
	e := os.Getenv(env)
	if e == "" {
		return defaultVal
	} else {
		return e
	}
}

func GetEnvOrDefaultInt(env string, defaultVal int64) int64 {
	e := os.Getenv(env)
	if e == "" {
		return defaultVal
	} else {
		intVal, err := strconv.ParseInt(e, 10, 16)
		if err != nil {
			logger.Error().Msg(fmt.Sprintf("Failed to parse string to int '%s'", env))
			os.Exit(1)
		}

		return (intVal)
	}
}

func GenRandomID(prefix string) string {
	return prefix + gonanoid.MustGenerate("abcdefghijklmonpqrstuvwxyzABCDEFGHIJKLMONPQRSTUVWXYZ0123456789", 22)
}

func GenKSortedID(prefix string) string {
	return prefix + ksuid.New().String()
}

func Ptr[T any](s T) *T {
	return &s
}

func Deref[T any](ref *T, fallback T) T {
	if ref == nil {
		return fallback
	}
	return *ref
}
