-- Удаляем старую таблицу перед учебной инициализацией базы.
DROP TABLE IF EXISTS tasks;

-- Создаём таблицу, в которой одна строка соответствует одной задаче.
CREATE TABLE tasks (
	-- SERIAL автоматически выдаёт следующий числовой ID.
	id SERIAL PRIMARY KEY,
	-- Заголовок обязателен и ограничен 255 символами.
	title VARCHAR(255) NOT NULL,
	-- Описание сейчас может быть NULL.
	description TEXT,
	-- Новая задача по умолчанию считается невыполненной.
	completed BOOLEAN NOT NULL DEFAULT FALSE,
	-- PostgreSQL сам подставляет время создания.
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	-- При создании updated_at совпадает с created_at.
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Добавляем стартовые данные, чтобы API можно было проверить сразу после запуска.
INSERT INTO tasks (title, description, completed) VALUES
	('Изучить Go', 'Изучить язык программирования Go и его особенности', TRUE),
	('Написать REST API', 'Создать REST API для управления задачами с использованием Go и PostgreSQL', FALSE),
	('Зарелизить приложение', 'Развернуть приложение на сервере', FALSE);
