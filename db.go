package main

import "github.com/jmoiron/sqlx"

const (
	vacanciesQuery = `SELECT id, name, salary, experience, city 
					  FROM vacancy
					  ORDER BY name`

	vacancyQuery = `SELECT id, name, salary, experience, city 
					FROM vacancy
					WHERE id = $1`

	deleteQuery = `DELETE FROM vacancy
				   WHERE id = $1`

	createQuery = `INSERT INTO vacancy (id, name, salary, experience, city) VALUES
				   ($1, $2, $3, $4, $5)
				   ON CONFLICT (id) DO
	   			   UPDATE
	   			   SET name = $2,
	   	   		   salary = $3,
	 	   		   experience = $4,
				   city = $5`

	allowedQuery = `SELECT allowed_method FROM permission
					INNER JOIN role ON permission.role_id = role.id
					INNER JOIN account ON role.id = account.role_id
					WHERE account.login = $1 AND account.password = $2`
)

// database структура подключения к базе данных
type database struct {
	conn *sqlx.DB
}

// dbService представляет интерфейс взаимодействия с базой данных
type dbService interface {
	userPermissions(string, string) ([]string, error)
	createVacancy(*vacancy) error
	vacancies() ([]*vacancy, error)
	vacancy(int) (*vacancy, error)
	deleteVacancy(int) error
}

// newDB открывает соединение с базой данных и создаёт основную структуру сервиса
func newDB(connectionString string) (database, error) {
	dbConn, err := sqlx.Open("postgres", connectionString)
	return database{conn: dbConn}, err
}

// userPermissions выполняет запрос на получение списка разрешённых методов доступа
func (db database) userPermissions(login string, password string) ([]string, error) {
	allowed := make([]string, 0)
	err := db.conn.Select(&allowed, allowedQuery, login, password)
	return allowed, err
}

// createVacancy выполняет запрос на создание новой вакансии или перезаписывает существующую
func (db database) createVacancy(v *vacancy) error {
	_, err := db.conn.Exec(createQuery, v.ID, v.Name, v.Salary, v.Experience, v.City)
	return err
}

// vacancies выполняет запрос на получение списка вакансий
func (db database) vacancies() ([]*vacancy, error) {
	vacancies := make([]*vacancy, 0)
	err := db.conn.Select(&vacancies, vacanciesQuery)
	return vacancies, err
}

// vacancy выполняет запрос на получение конкретной вакансии по id
func (db database) vacancy(id int) (*vacancy, error) {
	vac := &vacancy{}
	err := db.conn.Get(vac, vacancyQuery, id)
	return vac, err
}

// deleteVacancy выполняет запрос на удаление вакансии по id
func (db database) deleteVacancy(id int) error {
	_, err := db.conn.Exec(deleteQuery, id)
	return err
}
