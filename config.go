package main

import (
	"encoding/json"
	"log"
	"os"
)

type configuration struct {
	DbConnectionString string
	LoggerPath         string
	Server             struct {
		Address         string
		Port            string
		TLS             bool
		CertificatePath string
		KeyPath         string
	}
}

// loadConfig читает и парсит настройки сервиса
func loadConfig(configPath string) configuration {
	var config configuration
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Не удалось открыть файл конфигурации: %v", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Не удалось прочесть файл конфигурации: %v", err)
	}
	return config
}
