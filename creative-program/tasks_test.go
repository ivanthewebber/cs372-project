package tasks

import (
	"fmt"
	"html/template"
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mssql"
)

func TestTemplate(t *testing.T) {
	var templ = template.Must(template.ParseFiles("tasks.html"))
	f, _ := os.OpenFile("out.html", os.O_RDWR, os.ModePerm)
	defer f.Close()

	type List struct {
		Title string
		Tasks []Task
	}

	type UserFile struct {
		Owner string
		Lists []List
	}

	now := time.Now()
	err := templ.Execute(f,
		UserFile{
			Owner: "Ivan Webber",
			Lists: []List{
				List{
					Title: "Make this work",
					Tasks: []Task{
						Task{
							gorm.Model{0, now, now, &now},
							"The Title",
							"The details...",
							"A DueDate",
							true,
							0,
						},
						Task{
							gorm.Model{0, now, now, &now},
							"Not The Title",
							"Not The details...",
							"Not Today",
							false,
							0,
						},
					},
				},
			},
		})

	if err != nil {
		t.Error("Test Template: ", err.Error())
	}
}

// ReadAllTasks prints all the tasks
func ReadAllTasks(db *gorm.DB) {
	var users []User
	db.Find(&users) // find all users type objects
	for _, user := range users {
		var lists []TaskList
		db.Model(&user).Related(&lists) // all lists for user

		fmt.Printf("%s %s's lists:\n", user.FirstName, user.LastName)

		for _, tl := range lists {
			var tasks []Task
			db.Model(&tl).Related(&tasks)

			fmt.Printf("\t%s:\n", tl.Title)
			for _, task := range tasks {
				fmt.Printf("\t\tTitle: %s\n\t\tDueDate: %s\n\t\tDetails: %s\n\t\tCompleted:%t\n\n",
					task.Title, task.DueDate, task.Details, task.Completed)
			}
		}
	}
}

// UpdateSomeonesTask updates a task based on a user
func UpdateATask(db *gorm.DB, TaskListID int) {
	var task Task
	db.Where("task_list_id = ?", TaskListID).First(&task).Update("Title", "Buy donuts for Luis")
	fmt.Printf("Title: %s\nDueDate: %s\nCompleted:%t\n\n",
		task.Title, task.DueDate, task.Completed)
}

// DeleteSomeonesTasks deletes all the tasks for a user
func DeleteListContents(db *gorm.DB, TaskListID int) {
	db.Where("task_list_id = ?", TaskListID).Delete(&Task{})
	fmt.Printf("Deleted all tasks in list %d", TaskListID)
}

func ExampleConnectEtcDB() {
	db := MustConnect()
	defer db.Close()

	fmt.Println("Reseting DB...")
	db.DropTableIfExists(&User{}, &TaskList{}, &Task{})

	fmt.Println("Migrating models...")
	db.AutoMigrate(&User{})
	db.AutoMigrate(&TaskList{})
	db.AutoMigrate(&Task{})

	// Create awesome Users
	fmt.Println("Creating awesome users...")
	db.Create(&User{FirstName: "Andrea", LastName: "Lam"})   //UserID: 1
	db.Create(&User{FirstName: "Meet", LastName: "Bhagdev"}) //UserID: 2
	db.Create(&User{FirstName: "Luis", LastName: "Bosquez"}) //UserID: 3

	// Create list for each user
	fmt.Println("Creating lists...")
	db.Create(&TaskList{
		Title: "Andrea's list", UserID: 1})
	db.Create(&TaskList{
		Title: "Meet's List", UserID: 2})
	db.Create(&TaskList{
		Title: "Luis's List", UserID: 3})
	db.Create(&TaskList{
		Title: "Luis's Other List", UserID: 3})

	// Create appropriate Tasks for each user
	fmt.Println("Creating new appropriate tasks...")
	db.Create(&Task{
		Title: "Do laundry", DueDate: "2017-03-30", Completed: false, TaskListID: 1})
	db.Create(&Task{
		Title: "Mow the lawn", DueDate: "2017-03-30", Completed: false, TaskListID: 2})
	db.Create(&Task{
		Title: "Do more laundry", DueDate: "2017-03-30", Completed: false, TaskListID: 3})
	db.Create(&Task{
		Title: "Watch TV", DueDate: "2017-03-30", Completed: false, TaskListID: 3})

	// Read
	fmt.Println("\nReading all the tasks...")
	ReadAllTasks(db)

	// Update - update Task title to something more appropriate
	fmt.Println("Updating Andrea's task...")
	UpdateATask(db, 1)

	// Delete - delete Luis's task
	DeleteListContents(db, 3)
}
