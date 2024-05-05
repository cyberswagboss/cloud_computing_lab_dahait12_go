package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func connectToDb() *sql.DB {
	db, err := sql.Open("mysql", "go_user:go_pwd@tcp(localhost:3306)/tododb")
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

		todo := request.URL.Query().Get("todo")

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

		todo := request.URL.Query().Get("todo")

		if todo == "" {
			http.Error(writer, "Missing todo parameter", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("DELETE FROM todos WHERE task = ?", todo)
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
