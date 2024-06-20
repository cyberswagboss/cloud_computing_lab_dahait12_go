package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const todoPrefix = "/todos/"

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func connectToDb() *sql.DB {
	dbUser := flag.String("db_user", getEnv("DB_USER", "go_user"), "Database user")
	dbPassword := flag.String("db_password", getEnv("DB_PASSWORD", "go_pwd"), "Database password")
	dbHost := flag.String("db_host", getEnv("DB_HOST", "mysql"), "Database host")
	dbPort := flag.String("db_port", getEnv("DB_PORT", "3306"), "Database port")
	dbName := flag.String("db_name", getEnv("DB_NAME", "tododb"), "Database name")

	flag.Parse()

	databaseParams := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", *dbUser, *dbPassword, *dbHost, *dbPort, *dbName)
	db, err := sql.Open("mysql", databaseParams)
	if err != nil {
		panic(err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		os.Exit(1)
	}

	fmt.Print("DB Connected!\n")

	return db
}

func getTodos(db *sql.DB) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			http.Error(writer, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var todos []string
		rows, err := db.Query("SELECT task FROM todos")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var task string
			if err := rows.Scan(&task); err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			todos = append(todos, task)
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(todos)
	}
}

func addTodo(db *sql.DB) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			http.Error(writer, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		todo := strings.TrimPrefix(request.URL.Path, todoPrefix)
		decodedTodo, err := url.PathUnescape(todo)
		if err != nil {
			http.Error(writer, "Invalid todo parameter", http.StatusBadRequest)
			return
		}

		if decodedTodo == "" {
			http.Error(writer, "Missing todo parameter", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("INSERT INTO todos (task) values (?)", decodedTodo)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Write([]byte("todo added"))
	}
}

func removeTodo(db *sql.DB) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "DELETE" {
			http.Error(writer, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(request.URL.Path, todoPrefix)
		parts := strings.SplitN(path, "/", 2)
		if len(parts) < 1 || parts[0] == "" {
			http.Error(writer, "Missing todo parameter", http.StatusBadRequest)
			return
		}

		decodedTodo, err := url.PathUnescape(parts[0])
		if err != nil {
			http.Error(writer, "Invalid todo parameter", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("DELETE FROM todos WHERE task = ?", decodedTodo)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Write([]byte("Todo deleted"))
	}
}

func handleTodos(db *sql.DB) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "GET":
			getTodos(db)(writer, request)
		case "POST":
			addTodo(db)(writer, request)
		case "DELETE":
			removeTodo(db)(writer, request)
		default:
			http.Error(writer, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}

func main() {
	db := connectToDb()
	defer db.Close()

	http.HandleFunc("/todos", handleTodos(db))
	http.HandleFunc("/todos/", handleTodos(db))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
