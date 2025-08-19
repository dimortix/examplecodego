package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB
var MockMode bool = false

func InitDB() {
	if os.Getenv("MOCK_DB") == "true" {
		MockMode = true
		log.Println("Запуск в режиме без базы данных (MOCK_DB=true)")
		return
	}

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "planer")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Предупреждение: Не удалось подключиться к базе данных: %v", err)
		log.Println("Переключение в режим без базы данных")
		MockMode = true
		return
	}

	err = DB.Ping()
	if err != nil {
		log.Printf("Предупреждение: Не удалось проверить соединение с базой данных: %v", err)
		log.Println("Переключение в режим без базы данных")
		MockMode = true
		return
	}

	log.Println("Успешное подключение к базе данных")

	createTables()
}

func createTables() {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Не удалось создать таблицу users: %v", err)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS apartment_plans (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			title VARCHAR(100) NOT NULL,
			area FLOAT NOT NULL,
			rooms INTEGER NOT NULL,
			style VARCHAR(50) NOT NULL,
			features TEXT[],
			floor_plan TEXT,
			render_3d TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Не удалось создать таблицу apartment_plans: %v", err)
	}

	log.Println("Таблицы успешно созданы")
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Соединение с базой данных закрыто")
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
