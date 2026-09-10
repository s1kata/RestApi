package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Connect открывает подключение к PostgreSQL и настраивает пул соединений.
func Connect(databaseURL string) (*sqlx.DB, error) {
	// sqlx.Connect одновременно создаёт DB и проверяет, что база доступна.
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		// Ошибку подключения передаём вызывающему коду, чтобы он мог остановить запуск.
		return nil, err
	}

	// Ограничиваем количество одновременно открытых соединений с PostgreSQL.
	db.SetMaxOpenConns(20)
	// Оставляем несколько свободных соединений для повторного использования.
	db.SetMaxIdleConns(5)
	// Возвращаем настроенный пул и nil вместо ошибки.
	return db, nil
}
