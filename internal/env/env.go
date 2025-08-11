package env

import (
	"os"
	"strconv"
)

func GetString(key string, defaultValue string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return defaultValue
	}

	return val
}

func GetInt(key string, defaultValue int) int {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return defaultValue
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}

	return intVal
}
