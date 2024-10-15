package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB
var tmpl *template.Template

type Task struct {
	Id   int
	Task string
	Done bool
}

func init() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:password@(127.0.0.1:3306)/testdbgolang?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	// check db connection
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	initDB()
	defer db.Close()

	gRouter := mux.NewRouter()

	gRouter.HandleFunc("/", HomeHandler)

	//get tasks
	gRouter.HandleFunc("/tasks", fetchTasks).Methods("GET")

	//fetch add task form
	gRouter.HandleFunc("/getnewtaskform", getTaskForm)

	//add task
	gRouter.HandleFunc("/tasks", addTask).Methods("POST")

	//fetch update form
	gRouter.HandleFunc("/gettaskupdateform/{id}", getTaskUpdateForm)

	//update task
	gRouter.HandleFunc("/tasks/{id}", updateTask).Methods("PUT", "POST")

	//delete task
	gRouter.HandleFunc("/tasks/{id}", deleteTask).Methods("DELETE")

	http.ListenAndServe(":3000", gRouter)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "home.html", nil)
	if err != nil {
		http.Error(w, "Error executing template :"+err.Error(), http.StatusInternalServerError)
	}
}

func fetchTasks(w http.ResponseWriter, r *http.Request) {
	todos, _ := getTasks(db)

	tmpl.ExecuteTemplate(w, "todoList", todos)
}

func getTaskForm(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "addTaskForm", nil)
}

func addTask(w http.ResponseWriter, r *http.Request) {
	task := r.FormValue("task")

	query := "INSERT INTO tasks (task) VALUES (?)"

	stmt, err := db.Prepare(query)

	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	_, executeErr := stmt.Exec(task)

	if executeErr != nil {
		log.Fatal(executeErr)
	}

	todos, _ := getTasks(db)

	tmpl.ExecuteTemplate(w, "todoList", todos)
}

func getTaskUpdateForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	taskId, _ := strconv.Atoi(vars["id"])

	task, err := getTaskById(db, taskId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	tmpl.ExecuteTemplate(w, "updateTaskForm", task)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	taskItem := r.FormValue("task")
	isDone := r.FormValue("done")

	var taskStatus bool

	switch strings.ToLower(isDone) {
	case "yes", "on":
		taskStatus = true
	case "no", "off":
		taskStatus = false
	default:
		taskStatus = false
	}

	taskId, _ := strconv.Atoi(vars["id"])

	task := Task{
		taskId,
		taskItem,
		taskStatus,
	}

	query := "UPDATE tasks SET task = ?, done = ? WHERE id = ?"

	result, err := db.Exec(query, task.Task, task.Done, task.Id)

	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		fmt.Println("No rows updated")
	}

	todos, _ := getTasks(db)

	tmpl.ExecuteTemplate(w, "todoList", todos)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	taskId, _ := strconv.Atoi(vars["id"])

	query := "DELETE FROM tasks WHERE id = ?"

	stmt, err := db.Prepare(query)

	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, deleteError := stmt.Exec(taskId)

	if deleteError != nil {
		log.Fatal(deleteError)
	}

	todos, _ := getTasks(db)

	tmpl.ExecuteTemplate(w, "todoList", todos)
}

func getTasks(dbPointer *sql.DB) ([]Task, error) {

	query := "SELECT id, task, done FROM tasks"

	rows, err := dbPointer.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var todo Task
		rowErr := rows.Scan(&todo.Id, &todo.Task, &todo.Done)
		if rowErr != nil {
			return nil, rowErr
		}

		tasks = append(tasks, todo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func getTaskById(dbPointer *sql.DB, id int) (*Task, error) {

	query := "SELECT id, task, done FROM tasks WHERE id = ?"

	var task Task

	row := dbPointer.QueryRow(query, id)

	err := row.Scan(&task.Id, &task.Task, &task.Done)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("No task was found with id %d", id)
		}

		return nil, err
	}

	return &task, nil
}
