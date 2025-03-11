package main

import (
	"flag"
	"log"
	"os"
	"soulgame/iternal/app"
)

func main() {

	var cfg app.Config
	flag.StringVar(&cfg.DB.Host, "db-host", getEnvOrDefault("DB_HOST", "localhost"), "PostgreSQL host")
	flag.StringVar(&cfg.DB.Port, "db-port", getEnvOrDefault("DB_PORT", "5432"), "PostgreSQL port")
	flag.StringVar(&cfg.DB.User, "db-user", getEnvOrDefault("DB_USER", "postgres"), "PostgreSQL user")
	flag.StringVar(&cfg.DB.Password, "db-password", getEnvOrDefault("DB_PASSWORD", "postgres"), "PostgreSQL password")
	flag.StringVar(&cfg.DB.Name, "db-name", getEnvOrDefault("DB_NAME", "gotth-boilerplate"), "PostgreSQL database name")
	flag.StringVar(&cfg.Server.Port, "port", getEnvOrDefault("PORT", "8080"), "Server port")
	flag.Parse()

	application := app.NewApplication(cfg)

	if err := application.ConnectToDatabase(); err != nil {
		log.Fatal("Could not connect to PostgreSQL: ", err)
	}

}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
