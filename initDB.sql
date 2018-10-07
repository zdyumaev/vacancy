CREATE TABLE vacancy (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	salary INT NOT NULL CHECK (salary > 0),         
	experience TEXT NOT NULL,
	city TEXT NOT NULL);

INSERT INTO vacancy (name, salary, experience, city) VALUES 
    ('Backend-разработчик Go/Golang', 1, '1–3 года', 'Москва'),
    ('Presale-консультант (JSOC)', 2, '1–3 года', 'Москва'),
    ('Scala Архитектор / Team lead', 3, '3–6 лет', 'Москва');

CREATE TABLE  role (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL);

INSERT INTO role (id, name) VALUES
    (1, 'viewer'),
    (2, 'editor');

CREATE TABLE permission (
    role_id INT REFERENCES role (id),
    allowed_method TEXT,
    PRIMARY KEY (role_id, allowed_method));

INSERT INTO permission (role_id, allowed_method) VALUES
    (1, 'GET'),
    (2, 'GET'),
    (2, 'PUT'),
    (2, 'DELETE');

CREATE TABLE account (
	id SERIAL PRIMARY KEY,
	login TEXT UNIQUE  NOT NULL,
	password TEXT NOT NULL,
    role_id INT NOT NULL REFERENCES role (id));

INSERT INTO account (login, password, role_id) VALUES
    ('vi', 'pass_vi', 1),
    ('ed', 'pass_ed', 2);