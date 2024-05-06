package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const todoPrefix = "/test/"

func connectToDb() *sql.DB {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	databseParams := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", databseParams)
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

		json.NewEncoder(writer).Encode(todos)
	}

}

func addTodo(db *sql.DB) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			http.Error(writer, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		todo := strings.Trim(request.URL.Path, todoPrefix)

		if todo == "" {
			http.Error(writer, "Missing todo parameter", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("INSERT INTO todos (task) values (?)", todo)
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

		const todoPrefix = "/todos/"
		todo := strings.TrimPrefix(request.URL.Path, todoPrefix)

		if todo == "" || todo == "/" {
			http.Error(writer, "Missing todo parameter", http.StatusBadRequest)
			return
		}

		decodedTodo, err := url.QueryUnescape(todo)
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

func main() {
	db := connectToDb()
	defer db.Close()

	http.HandleFunc("/todos", getTodos(db))
	http.HandleFunc("/todos/", getTodos(db))
	http.HandleFunc("/todos/{todo}", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "POST":
			addTodo(db)(writer, request)
		case "DELETE":
			removeTodo(db)(writer, request)
		default:
			http.Error(writer, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
