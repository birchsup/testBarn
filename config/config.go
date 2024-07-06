package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	viper.SetDefault("environment", "development")

	// Set DATABASE_URL environment variable if not already set
	if os.Getenv("DATABASE_URL") == "" {
		environment := viper.GetString("environment")
		os.Setenv("DATABASE_URL", viper.GetString(environment+".database_url"))
	}
}
