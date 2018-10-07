package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockDB struct{}

func (m mockDB) userPermissions(login string, password string) ([]string, error) {
	if login == "vi" && password == "pass_vi" {
		return []string{"GET"}, nil
	}
	if login == "ed" && password == "pass_ed" {
		return []string{"GET", "PUT", "DELETE"}, nil
	}
	return []string{}, nil
}

func (m mockDB) createVacancy(v *vacancy) error {
	if v.Salary <= 0 {
		return errors.New("")
	}
	return nil
}

func (m mockDB) vacancies() ([]*vacancy, error) {
	return []*vacancy{&vacancy{}, &vacancy{}}, nil
}

func (m mockDB) vacancy(id int) (*vacancy, error) {
	if id == 42 {
		return &vacancy{
			ID:         42,
			Name:       "",
			Salary:     1,
			Experience: "",
			City:       ""}, nil
	}
	return nil, errors.New("")
}

func (m mockDB) deleteVacancy(id int) error {
	return nil
}

func dummyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

var db mockDB

func TestAuthefication(t *testing.T) {
	// Проверка метода GET для viewer`а
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://127.0.0.1:8080/vacancy", nil)
	r.SetBasicAuth("vi", "pass_vi")
	authHandler(dummyHandler(), db).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Ошибка аутентификации пользователя vi(пароль - pass_vi), метод GET недоступен. Код: %v", w.Code)
	}
	// Проверка метода DELETE для viewer`а
	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "http://127.0.0.1:8080/vacancy/1", nil)
	r.SetBasicAuth("vi", "pass_vi")
	authHandler(dummyHandler(), db).ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Errorf("Ошибка аутентификации пользователя vi(пароль - pass_vi), метод DELETE доступен. Код: %v", w.Code)
	}
	// Проверка метода DELETE для editor`а
	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "http://127.0.0.1:8080/vacancy/1", nil)
	r.SetBasicAuth("ed", "pass_ed")
	authHandler(dummyHandler(), db).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("Ошибка аутентификации пользователя ed(пароль - pass_ed), метод DELETE недоступен. Код: %v", w.Code)
	}
}

func TestVacancyHandler(t *testing.T) {
	// Проверка "GET /vacancy"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://127.0.0.1:8080/vacancy", nil)
	vacancyHandler(db).ServeHTTP(w, r)
	vacancies := make([]*vacancy, 0)
	err := json.Unmarshal(w.Body.Bytes(), &vacancies)
	if err != nil {
		t.Errorf("Ошибка демаршализации тела ответа: %v", err)
	}
	if len(vacancies) != 2 {
		t.Error("Ошибка получения списка вакансий")
	}
	// Проверка корректного "PUT /vacancy"
	w = httptest.NewRecorder()
	vac := vacancy{
		ID:         1,
		Name:       "",
		Salary:     1,
		Experience: "",
		City:       ""}
	js, err := json.Marshal(vac)
	if err != nil {
		t.Errorf("Ошибка маршалинга тела запроса: %v", err)
	}
	r = httptest.NewRequest("PUT", "http://127.0.0.1:8080/vacancy", bytes.NewReader(js))
	vacancyHandler(db).ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Errorf("Ошибка корректного создания вакансии PUT /vacancy. Код: %v", w.Code)
	}
	// Проверка некорректного "PUT /vacancy"
	w = httptest.NewRecorder()
	vac.Salary = -1
	js, err = json.Marshal(vac)
	if err != nil {
		t.Errorf("Ошибка маршалинга тела запроса: %v", err)
	}
	r = httptest.NewRequest("PUT", "http://127.0.0.1:8080/vacancy", bytes.NewReader(js))
	vacancyHandler(db).ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Ошибка некорректного создания вакансии PUT /vacancy. Код: %v", w.Code)
	}
	// Проверка непооддерживаемого "DELETE /vacancy"
	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "http://127.0.0.1:8080/vacancy", nil)
	vacancyHandler(db).ServeHTTP(w, r)
	if w.Code != http.StatusNotImplemented {
		t.Errorf("Ошибка некорректного создания вакансии PUT /vacancy. Код: %v", w.Code)
	}
}

func TestVacancySlashHandler(t *testing.T) {
	// Проверка "GET /vacancy/42"
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://127.0.0.1:8080/vacancy/42", nil)
	vacancySlashHandler(db).ServeHTTP(w, r)
	vac := vacancy{}
	err := json.Unmarshal(w.Body.Bytes(), &vac)
	if err != nil {
		t.Errorf("Ошибка демаршализации тела ответа: %v", err)
	}
	if vac.ID != 42 {
		t.Errorf("Ошибка получения вакансии. Id = %v", vac.ID)
	}
	// Проверка корректного "PUT /vacancy"
	w = httptest.NewRecorder()
	js, err := json.Marshal(vac)
	if err != nil {
		t.Errorf("Ошибка маршалинга тела запроса: %v", err)
	}
	r = httptest.NewRequest("PUT", "http://127.0.0.1:8080/vacancy/", bytes.NewReader(js))
	vacancySlashHandler(db).ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Errorf("Ошибка корректного создания вакансии PUT /vacancy/. Код: %v", w.Code)
	}
	// Проверка некорректного "PUT /vacancy/42" (с id)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "http://127.0.0.1:8080/vacancy/42", bytes.NewReader(js))
	vacancySlashHandler(db).ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Ошибка некорректного создания вакансии PUT /vacancy/42. Код: %v", w.Code)
	}
	// Проверка некорректного "DELETE /vacancy/42" (с id)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "http://127.0.0.1:8080/vacancy/f2", nil)
	vacancySlashHandler(db).ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Ошибка некорректного создания вакансии PUT /vacancy/f2. Код: %v", w.Code)
	}
}
