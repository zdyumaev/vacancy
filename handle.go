package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// authHandler аутентифицирует пользователя перед обработкой запроса
func authHandler(h http.HandlerFunc, db dbService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		login, password, ok := r.BasicAuth()
		if ok {
			allowedMethods, err := db.userPermissions(login, password)
			if err != nil {
				http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
				log.Printf("Ошибка получения списка методов доступа(логин: %v): %v", login, err)
				return
			}
			// Пользователь без разрешенных методов доступа получает 401
			if len(allowedMethods) == 0 {
				log.Printf("Попытка неавторизованного доступа. Логин: %v", login)
				http.Error(w, "Ошибка доступа", http.StatusUnauthorized)
				return
			}
			for _, v := range allowedMethods {
				if r.Method == v {
					// Вызов основного обработчика
					h.ServeHTTP(w, r)
					return
				}
			}
			// Ошибка в случае отсутствия используемого метода доступа в списке разрешённых
			log.Printf("Попытка доступа неразрешенным методом. Логин: %v", login)
			http.Error(w, "Ошибка доступа", http.StatusForbidden)
			return
		}
		log.Printf("Попытка доступа без аутентификации")
		http.Error(w, "Ошибка доступа", http.StatusUnauthorized)
		return
	}
}

// vacancyHandler реализует функционал сервиса для запросов без id ("/vacancy")
func vacancyHandler(db dbService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getAllMethod(w, db)
		case "PUT":
			putMethod(w, r, db)
		default:
			// На случай, если мы попали сюда случайно из-за разрешения дополнительного метода в базе данных
			// А так же при запросе "DELETE /vacancy"
			http.Error(w, "Метод не поддерживается", http.StatusNotImplemented)
			log.Printf("Неподдерживаемый метод")
		}
	}
}

// vacancyHandler реализует функционал сервиса для запросов с id ("/vacancy/id")
func vacancySlashHandler(db dbService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Выделение id вакансии
		param := strings.TrimPrefix(r.URL.Path, "/vacancy/")
		switch r.Method {
		case "GET":
			if param == "" {
				getAllMethod(w, db)
			} else {
				getOneMethod(w, param, db)
			}
		case "DELETE":
			deleteMethod(w, param, db)
		case "PUT":
			if param != "" {
				http.Error(w, "Ошибка параметра", http.StatusBadRequest)
				return
			}
			putMethod(w, r, db)
		default:
			// На случай, если мы попали сюда случайно из-за разрешения дополнительного метода в базе данных
			http.Error(w, "Метод не поддерживается", http.StatusNotImplemented)
			log.Printf("Неподдерживаемый метод")
		}
	}
}

// getAllMethod выполняет запрос "GET /vacancy" (получить список вакансий)
func getAllMethod(w http.ResponseWriter, db dbService) {
	vacancies, err := db.vacancies()
	if err != nil {
		http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		log.Printf("Ошибка запроса списка вакансий: %v", err)
		return
	}
	js, err := json.Marshal(vacancies)
	if err != nil {
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		log.Printf("Ошибка маршалинга вакансий: %v", err)
		return
	}
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	_, err = w.Write(js)
	if err != nil {
		log.Printf("Ошибка записи результата запроса: %v", err)
	}
	return
}

// getOneMethod выполняет запрос "GET /vacancy/id" (получить отдельную вакансию)
func getOneMethod(w http.ResponseWriter, param string, db dbService) {
	id, err := strconv.Atoi(param)
	if err != nil {
		http.Error(w, "Ошибка параметра", http.StatusBadRequest)
		log.Printf("Ошибка конвертации id вакансии(%v): %v", param, err)
		return
	}
	vac, err := db.vacancy(id)
	if err == sql.ErrNoRows {
		http.Error(w, "Не найдено", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		log.Printf("Ошибка запроса вакансии(%v): %v", id, err)
		return
	}
	js, err := json.Marshal(vac)
	if err != nil {
		http.Error(w, "Ошибка формирования ответа", http.StatusInternalServerError)
		log.Printf("Ошибка маршалинга вакансии(%v): %v", id, err)
		return
	}
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	_, err = w.Write(js)
	if err != nil {
		log.Printf("Ошибка записи результата запроса: %v", err)
	}
}

// deleteMethod выполняет обработку "DELETE /vacancy/id" (удалить вакансию)
func deleteMethod(w http.ResponseWriter, param string, db dbService) {
	id, err := strconv.Atoi(param)
	if err != nil {
		http.Error(w, "Ошибка параметра", http.StatusBadRequest)
		log.Printf("Ошибка конвертации id вакансии(%v): %v", param, err)
		return
	}
	err = db.deleteVacancy(id)
	if err != nil {
		http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		log.Printf("Ошибка удаления вакансии(%v): %v", id, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// putMethod выполняет обработку "PUT /vacancy" (создать вакансию)
func putMethod(w http.ResponseWriter, r *http.Request, db dbService) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		log.Printf("Ошибка чтения тела запроса: %v", err)
		return
	}
	vac := &vacancy{}
	err = json.Unmarshal(body, vac)
	if err != nil {
		http.Error(w, "Ошибка запроса", http.StatusBadRequest)
		log.Printf("Ошибка демаршализации тела запроса: %v", err)
		return
	}
	// Проверка логического ограничения
	if vac.Salary <= 0 {
		http.Error(w, "Ошибка запроса", http.StatusBadRequest)
		log.Print("Поле запроса 'Зарплата' должно быть положительным")
		return
	}
	err = db.createVacancy(vac)
	if err != nil {
		http.Error(w, "Ошибка обработки запроса", http.StatusInternalServerError)
		log.Printf("Ошибка записи запроса: %v", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
