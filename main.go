package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

// Названия роутов
const (
	prefixID = "/vacancy/"
	prefix   = "/vacancy"
)

// Путь к файлу настроек
var configurationPath = flag.String("config", "config.json", "Путь к файлу конфигурации")

// vacancy является структурой сущности отдельной вакансии
type vacancy struct {
	ID         int
	Name       string
	Salary     int
	Experience string
	City       string
}

func main() {
	flag.Parse()
	config := loadConfig(*configurationPath)

	if config.LoggerPath != "" {
		// Логер только добавляет данные
		logFile, err := os.OpenFile(config.LoggerPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Printf("Ошибка открытия файла лога: %v", err)
		} else {
			defer logFile.Close()
			log.SetOutput(logFile)
		}
	}

	db, err := newDB(config.DbConnectionString)
	if err != nil {
		log.Printf("Не удалось подключиться к базе данных: %v", err)
	}

	// Установка цепочек обработчиков
	http.HandleFunc(prefixID, authHandler(vacancySlashHandler(db), db))
	http.HandleFunc(prefix, authHandler(vacancyHandler(db), db))

	listenString := config.Server.Address + ":" + config.Server.Port
	log.Print("Запуск сервера: ", listenString)

	if config.Server.TLS {
		err = http.ListenAndServeTLS(listenString, config.Server.CertificatePath, config.Server.KeyPath, nil)
	} else {
		err = http.ListenAndServe(listenString, nil)
	}
	if err != nil {
		log.Printf("Ошибка веб-сервера: %v", err)
	}
}
