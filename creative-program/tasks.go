package tasks

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"

	// register sql driver
	_ "github.com/jinzhu/gorm/dialects/mssql"
)

var server = "localhost"
var port = 1433
var database = "SampleDB" // TODO: make a new TasksDB

// User is a named owner of lists
type User struct {
	gorm.Model
	FirstName string
	LastName  string
}

// Task is a to-do item
type Task struct {
	gorm.Model
	Title      string
	Details    string
	DueDate    string
	IsComplete bool
	//TaskListID uint
	UserID uint // replace with TaskListID
}

// TaskList is named set of tasks
type TaskList struct {
	//gorm.Model
	Title string
	List  []Task // matched by id in db
	//userID uint
}

// UserFile is User and their TaskLists
type UserFile struct {
	Owner User
	Lists []TaskList
}

// ReadAllTasks prints all the tasks
func ReadAllTasks(db *gorm.DB) {
	var users []User
	var tasks []Task
	db.Find(&users)

	for _, user := range users {
		db.Model(&user).Related(&tasks)
		fmt.Printf("%s %s's tasks:\n", user.FirstName, user.LastName)
		for _, task := range tasks {
			fmt.Printf("Title: %s\nDueDate: %s\nIsComplete:%t\n\n",
				task.Title, task.DueDate, task.IsComplete)
		}
	}
}

// UpdateSomeonesTask updates a task based on a user
func UpdateSomeonesTask(db *gorm.DB, userID int) {
	var task Task
	db.Where("user_id = ?", userID).First(&task).Update("Title", "Buy donuts for Luis")
	fmt.Printf("Title: %s\nDueDate: %s\nIsComplete:%t\n\n",
		task.Title, task.DueDate, task.IsComplete)
}

// DeleteSomeonesTasks deletes all the tasks for a user
func DeleteSomeonesTasks(db *gorm.DB, userID int) {
	db.Where("user_id = ?", userID).Delete(&Task{})
	fmt.Printf("Deleted all tasks for user %d", userID)
}

func not_main() {
	connectionString := fmt.Sprintf("server=%s;port=%d;database=%s",
		server, port, database)
	db, err := gorm.Open("mssql", connectionString)

	if err != nil {
		log.Fatal("Failed to create connection pool. Error: " + err.Error())
	}
	gorm.DefaultCallback.Create().Remove("mssql:set_identity_insert")
	defer db.Close()

	fmt.Println("Migrating models...")
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Task{})

	// Create awesome Users
	fmt.Println("Creating awesome users...")
	db.Create(&User{FirstName: "Andrea", LastName: "Lam"})   //userID: 1
	db.Create(&User{FirstName: "Meet", LastName: "Bhagdev"}) //userID: 2
	db.Create(&User{FirstName: "Luis", LastName: "Bosquez"}) //userID: 3

	// Create appropriate Tasks for each user
	fmt.Println("Creating new appropriate tasks...")
	db.Create(&Task{
		Title: "Do laundry", DueDate: "2017-03-30", IsComplete: false, userID: 1})
	db.Create(&Task{
		Title: "Mow the lawn", DueDate: "2017-03-30", IsComplete: false, userID: 2})
	db.Create(&Task{
		Title: "Do more laundry", DueDate: "2017-03-30", IsComplete: false, userID: 3})
	db.Create(&Task{
		Title: "Watch TV", DueDate: "2017-03-30", IsComplete: false, userID: 3})

	// Read
	fmt.Println("\nReading all the tasks...")
	ReadAllTasks(db)

	// Update - update Task title to something more appropriate
	fmt.Println("Updating Andrea's task...")
	UpdateSomeonesTask(db, 1)

	// Delete - delete Luis's task
	DeleteSomeonesTasks(db, 3)
}
