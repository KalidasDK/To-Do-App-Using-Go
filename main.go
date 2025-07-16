package main

import (
	"fmt"
	"log"
	"time"

	//"html"
	"database/sql"
	"net/http"
	"to-do-app/helper"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "todoapp"
)

type task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var db *sql.DB

func main() {

	initialize_database_and_tables()
	router := gin.Default()
	router.GET("/tasks", getTasks)
	router.POST("/tasks", addTask)
	router.GET("/tasks/completed", func(c *gin.Context) { getTasksByStatus(c, true) })
	router.GET("/tasks/incomplete", func(c *gin.Context) { getTasksByStatus(c, false) })
	router.DELETE("tasks/:id", deleteTask)
	router.PUT("/tasks/:id", updateTask)

	router.Run("localhost:8080")

}

func initialize_database_and_tables() {
	// connect to PostgreSQL server (without connecting to a database)
	serverConnStr := fmt.Sprintf("host = %s port = %d user = %s password = %s sslmode = disable",
		host, port, user, password) // Server connection string

	serverDB, err := sql.Open("postgres", serverConnStr)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL server: ", err)
	}
	defer serverDB.Close()

	// Test the server connection
	err = serverDB.Ping()
	if err != nil {
		log.Fatal("Failed to Ping PostgreSQL server: ", err)
	}

	// Create database if it doesn't exist
	err = helper.CreateDB(serverDB, dbname)
	if err != nil {
		log.Fatal("Failed to create database: ", err)
	}

	// Create a connection to the database
	dbConnStr := fmt.Sprintf("host = %s port = %d user = %s password = %s dbname = %s sslmode = disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatal("Failed to connect to the database: ", err)
	}
	//defer db.Close()

	// Test the database connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}
	fmt.Printf("Successfully connected to PostgreSQL database: %s\n", dbname)

	// Create tasks table
	err = helper.CreateTasksTable(db)
	if err != nil {
		log.Fatal(err)
	}
}

func getTasks(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tasks []task
	for rows.Next() {
		var t task
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, t)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, tasks)
}

func addTask(c *gin.Context) {
	var newTask task
	if err := c.BindJSON(&newTask); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	stmt, err := db.Prepare("INSERT INTO tasks (title, description) VALUES($1, $2)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(newTask.Title, newTask.Description); err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusCreated, newTask)

}

func getTasksByStatus(c *gin.Context, completed bool) {

	c.Header("Content-Type", "application/json")

	rows, err := db.Query("SELECT * FROM tasks where completed = $1", completed)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var completedTasks []task
	for rows.Next() {
		var t task
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			log.Fatal(err)
		}
		completedTasks = append(completedTasks, t)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, completedTasks)
}

func deleteTask(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func updateTask(c *gin.Context) {
	var req task
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	id := c.Param("id")

	result, err := db.Exec("UPDATE tasks SET title = $1, description = $2, completed = $3 WHERE id = $4", req.Title, req.Description, req.Completed, id)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}
