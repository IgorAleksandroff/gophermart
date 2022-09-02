package config

import (
	"flag"
	"log"
	"os"
	"sync"
)

const (
	ServerEnvServerAddress     = "RUN_ADDRESS"
	ServerDefaultServerAddress = "localhost:8080"

	DataBaseEnvAddress     = "DATABASE_URI"
	DataBaseDefaultAddress = ""

	AccrualSystemEnvAddress     = "ACCRUAL_SYSTEM_ADDRESS"
	AccrualSystemDefaultAddress = ""
)

type Config struct {
	AppConfig struct {
		DataBaseURI          string
		AccrualSystemAddress string
	}
	ServerConfig struct {
		ServerAddress string
	}
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		log.Println("Parse config with osArgs:", os.Args)

		hostFlag := flag.String("a", ServerDefaultServerAddress, "адрес и порт сервера")
		DBFlag := flag.String("d", DataBaseDefaultAddress, "адрес и порт сервера")
		AccrualFlag := flag.String("r", AccrualSystemDefaultAddress, "адрес и порт сервера")
		flag.Parse()

		instance = &Config{}

		instance.ServerConfig.ServerAddress = GetEnvString(ServerEnvServerAddress, *hostFlag)
		instance.AppConfig.DataBaseURI = GetEnvString(DataBaseEnvAddress, *DBFlag)
		instance.AppConfig.AccrualSystemAddress = GetEnvString(AccrualSystemEnvAddress, *AccrualFlag)

		log.Printf("Parsed config: %+v", instance)
	})

	return instance
}

func GetEnvString(envName, defaultValue string) string {
	value := os.Getenv(envName)
	if value == "" {
		log.Printf("empty env: %s, default: %s", envName, defaultValue)
		return defaultValue
	}
	return value
}
