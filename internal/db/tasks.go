package database

import (
	"database/sql"
	"fmt"
	"restapi-tasks/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
)

type TaskStore struct {
	// db - общий пул соединений, через который выполняются SQL-запросы.
	db *sqlx.DB
}

// NewTaskStore создаёт объект слоя доступа к данным.
func NewTaskStore(db *sqlx.DB) *TaskStore {
	return &TaskStore{db: db}
}


func (s *TaskStore) GetAll() ([]models.Task, error) {
	
	var tasks []models.Task

	
	query := `
	SELECT id, title, description, completed, created_at, updated_at
	FROM tasks
	ORDER BY created_at DESC`

	
	err := s.db.Select(&tasks, query)
	if err != nil {
	
		return nil, err
	}
	
	return tasks, nil
}


func (s *TaskStore) GetByID(id int) (*models.Task, error) {
	var task models.Task

	query := `
	SELECT id, title, description, completed, created_at, updated_at
	FROM tasks
	WHERE id = $1`
	
	err := s.db.Get(&task, query, id)
	if err == sql.ErrNoRows {
	
		return nil, fmt.Errorf("task with id %d not found", id)
	}
	if err != nil {
		
		return nil, err
	}
	return &task, nil
}

// CreateTask добавляет новую задачу и возвращает созданную запись.
func CreateTask(s *TaskStore, input models.CreateTaskInput) (*models.Task, error) {
	// Переменная будет заполнена данными из RETURNING.
	var task models.Task

	query := `
	INSERT INTO tasks (title, description, completed, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, title, description, completed, created_at, updated_at`

	// Один момент времени используется и для создания, и для обновления.
	now := time.Now()

	// Параметры input передаются отдельно, поэтому не склеиваются с SQL-строкой.
	err := s.db.Get(&task, query, input.Title, input.Description, input.Completed, now, now)
	if err != nil {
		return nil, err
	}
	// Возвращаем запись с ID и датами, назначенными при вставке.
	return &task, nil
}

// UpdateTask изменяет только поля, которые были переданы клиентом.
func UpdateTask(s *TaskStore, id int, input models.UpdateTaskInput) (*models.Task, error) {
	// Сначала читаем текущую запись, чтобы сохранить не переданные поля.
	task, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	// nil означает, что поле отсутствовало в JSON и менять его не нужно.
	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Completed != nil {
		task.Completed = *input.Completed
	}
	// Дата изменения обновляется при каждом успешном изменении.
	task.UpdatedAt = time.Now()

	query := `
	UPDATE tasks
	SET title = $1, description = $2, completed = $3, updated_at = $4
	WHERE id = $5
	RETURNING id, title, description, completed, created_at, updated_at`
	// QueryRow ожидает одну строку, которую затем Scan раскладывает по полям структуры.
	var updatedTask models.Task
	err = s.db.QueryRow(query, task.Title, task.Description, task.Completed, task.UpdatedAt, id).Scan(&updatedTask.ID, &updatedTask.Title, &updatedTask.Description, &updatedTask.Completed, &updatedTask.CreatedAt, &updatedTask.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &updatedTask, nil
}

// DELETEtask удаляет задачу и проверяет, была ли такая строка.
func DELETEtask(s *TaskStore, id int) error {
	// DELETE с параметром удаляет только запись с указанным ID.
	query := `DELETE FROM tasks WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// RowsAffected позволяет отличить удаление от ситуации, когда ID не найден.
	if rows == 0 {
		return fmt.Errorf("id %d not found", id)
	}
	return nil
}
func (s *TaskStore)GetAllFiltered(completed *bool)([]models.Task,error){
	var tasks []models.Task
	
	if completed == nil{
		
		query := `SELECT id, title, description, completed, created_at, updated_at
		FROM tasks
		ORDER BY created_at DESC`
		err := s.db.Select(&tasks, query)
		
		if err != nil{
			return nil,err
		}
		
	}
	if completed != nil{
		query := `SELECT id, title, description, completed, created_at, updated_at
		FROM tasks
		WHERE completed = $1
		ORDER BY created_at DESC`
		err := s.db.Select(&tasks, query,*completed)
		if err != nil{
			return nil, err
		}
	}
	return tasks, nil
	
}
func Statstask(s *TaskStore)(models.TaskStats,error){
		var stats models.TaskStats
		query := `
		SELECT 
		COUNT(*) AS total,
		COUNT(*) FILTER (WHERE  completed = true) AS completed,
		COUNT(*) FILTER (WHERE completed = false) AS pending
		FROM tasks
		`
		err := s.db.Get(&stats, query)
		if err != nil {
			return models.TaskStats{}, err
		}
		return stats, nil
		
	}