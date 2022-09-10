package config

import (
	"flag"
	"log"
	"os"
	"sync"
)

const (
	ServerAddressEnv     = "RUN_ADDRESS"
	ServerAddressDefault = "localhost:8080"

	DataBaseAddressEnv     = "DATABASE_URI"
	DataBaseAddressDefault = ""

	AccrualSystemEnvAddress     = "ACCRUAL_SYSTEM_ADDRESS"
	AccrualSystemAddressDefault = "http://localhost:80"

	LogLevelEnv     = "LOG_LEVEL"
	LogLevelDefault = "Info"
)

type (
	Config struct {
		App        appConfig
		HTTPServer serverConfig
	}

	appConfig struct {
		DataBaseURI          string
		AccrualSystemAddress string
		LogLevel             string
	}

	serverConfig struct {
		ServerAddress string
	}
)

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		log.Println("Parse config with osArgs:", os.Args)

		hostFlag := flag.String("a", ServerAddressDefault, "адрес и порт сервера")
		DBFlag := flag.String("d", DataBaseAddressDefault, "адрес и порт сервера")
		AccrualFlag := flag.String("r", AccrualSystemAddressDefault, "адрес и порт сервера")
		LogFlag := flag.String("l", LogLevelDefault, "адрес и порт сервера")
		flag.Parse()

		serverCfg := serverConfig{
			ServerAddress: getEnvString(ServerAddressEnv, *hostFlag),
		}

		appCfg := appConfig{
			DataBaseURI:          getEnvString(DataBaseAddressEnv, *DBFlag),
			AccrualSystemAddress: getEnvString(AccrualSystemEnvAddress, *AccrualFlag),
			LogLevel:             getEnvString(LogLevelEnv, *LogFlag),
		}

		instance = &Config{
			App:        appCfg,
			HTTPServer: serverCfg,
		}

		log.Printf("Parsed config: %+v", instance)
	})

	return instance
}

func getEnvString(envName, defaultValue string) string {
	value := os.Getenv(envName)
	if value == "" {
		log.Printf("empty env: %s, default: %s", envName, defaultValue)
		return defaultValue
	}
	return value
}
