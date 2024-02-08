package application

import (
	"os"
	"strconv"
)

type Config struct {
	RedisAdress string
	ServerPort  uint16
}

func LoadConfig() Config {
	cfg := Config{
		RedisAdress: "localhost:6379",
		ServerPort:  3000,
	}

	if redisAddr, exists := os.LookupEnv("REDIS_ADDR"); exists {
		cfg.RedisAdress = redisAddr
	}

	if serverPort, exists := os.LookupEnv("SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	return cfg
}
