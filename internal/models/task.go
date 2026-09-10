package models

import "time"

// Task описывает задачу одновременно в Go-коде, JSON-ответе и строке базы данных.
type Task struct {
	// ID - уникальный идентификатор задачи из PostgreSQL.
	ID int `json:"id" db:"id"`
	// Title - короткий обязательный заголовок задачи.
	Title string `json:"title" db:"title"`
	// Description - подробное описание задачи.
	Description string `json:"description" db:"description"`
	// Completed показывает, выполнена ли задача.
	Completed bool `json:"completed" db:"completed"`
	// CreatedAt хранит момент создания записи.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	// UpdatedAt хранит момент последнего изменения записи.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateTaskInput содержит только данные, которые клиент присылает при создании.
type CreateTaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// UpdateTaskInput использует указатели, чтобы отличить отсутствующее поле от значения false или пустой строки.
type UpdateTaskInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}
