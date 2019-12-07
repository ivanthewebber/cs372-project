package main

/*
	This file implements a RESTful web app for keeping track of To-Do items using Go.

	In order to learn how to implement each of the components of this application I followed a number of independent tutorials which I highly recomend:


	SQL Server, Windows, Go sql - https://www.microsoft.com/en-us/sql-server/developer-get-started/go/windows
	Writing Web Applications - https://golang.org/doc/articles/wiki/
	A Tour of Go - https://tour.golang.org/
	Using Templates - https://blog.gopheracademy.com/advent-2017/using-go-templates/

*/

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"

	"github.com/jinzhu/gorm"

	// register sql driver
	_ "github.com/jinzhu/gorm/dialects/mssql"
)

/*
	Model: The following section of code declares datatypes and implements access to storage on a database.
*/

// User is a named owner of lists
type User struct {
	gorm.Model
	FirstName string `gorm:"primary_key"`
	LastName  string `gorm:"primary_key"`
}

// Task is a to-do item
type Task struct {
	gorm.Model
	Title      string `gorm:"primary_key"`
	Details    string
	DueDate    string
	Completed  bool
	TaskListID uint
}

// TaskList is named set of tasks
type TaskList struct {
	gorm.Model
	Title  string `gorm:"primary_key"`
	UserID uint
}

func MustConnect() *gorm.DB {
	// constants for accessing the database
	var server = "localhost"
	var port = 1433
	var database = "SampleDB" // TODO: make a new TasksDB

	connectionString := fmt.Sprintf("server=%s;port=%d;database=%s",
		server, port, database)
	db, err := gorm.Open("mssql", connectionString)

	if err != nil {
		log.Fatal("Failed to create connection pool. Error: " + err.Error())
	}
	gorm.DefaultCallback.Create().Remove("mssql:set_identity_insert")
	return db
}

// Example connects to DB and initializes some tables
func Example() *gorm.DB {
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

	return db
}

/*
	View: this next section of code is about presentation to the user and functionality.

	Store templates in tmpl/ and page data in data/.
	Add a handler to make the web root redirect to /view/FrontPage.
	Spruce up the page templates by making them valid HTML and adding some CSS rules.
	Implement inter-page linking by converting instances of [PageName] to
	<a href="/view/PageName">PageName</a>. (hint: you could use regexp.ReplaceAllFunc to do this)
*/

// ./welcome
// ./login
// ./view/firstname lastname
// ./add/firstname lastname
// ./delete/firstname lastname
// ./add/firstname lastname/list
// ./delete/firstname lastname/list
// ./mark complete/firstname lastname/list/task

var validPath = regexp.MustCompile("^/(welcome|login)|((view|add|delete)/(\\w+) (\\w+))|((add|delete)/(\\w+) (\\w+)/([A-Za-z0-9]+))|((mark complete)/(\\w+) (\\w+)/([A-Za-z0-9]+)/([A-Za-z0-9]+))$")

var userPath = regexp.MustCompile("^/(view|add|delete)/(\\w+) (\\w+)")

func getName(url string) (first, last string) {
	m := userPath.FindStringSubmatch(url)

	if m != nil {
		return m[2], m[3]
	}
	return "", ""
}

var listPath = regexp.MustCompile("^/(view|add|delete)/(\\w+)\\s(\\w+)/([^/]+)")

func getNameList(url string) (first, last, list string) {
	m := listPath.FindStringSubmatch(url)
	if m != nil {
		return m[2], m[3], m[4]
	}
	return "", "", ""
}

var taskPath = regexp.MustCompile("^/(view|add|delete|mark complete)/(\\w+)\\s(\\w+)/([^/]+)/([^/]+)")

func getNameListTask(url string) (first, last, list, task string) {
	m := taskPath.FindStringSubmatch(url)
	if m != nil {
		return m[2], m[3], m[4], m[5]
	}
	return "", "", "", ""
}

func delHandler(w http.ResponseWriter, r *http.Request) {
	if taskPath.Match([]byte(r.URL.Path)) {
		delTaskHandler(w, r)
	} else {
		delListHandler(w, r)
	}
}

func delTaskHandler(w http.ResponseWriter, r *http.Request) {
	first, last, title, task := getNameListTask(r.URL.Path)
	fmt.Printf("F: %s\nL: %s\nTitle: %s\nTask: %s\n", first, last, title, task)

	var user User
	db.First(&user, &User{FirstName: first, LastName: last})

	var list TaskList
	db.First(&list, &TaskList{Title: title, UserID: user.ID})

	db.Delete(&Task{}, &Task{Title: task, TaskListID: list.ID})

	retToView(first, last, w, r)
}

func delListHandler(w http.ResponseWriter, r *http.Request) {
	first, last, title := getNameList(r.URL.Path)
	fmt.Printf("F: %s\nL: %s\nTitle: %s\n", first, last, title)

	var user User
	db.First(&user, &User{FirstName: first, LastName: last})

	var list TaskList
	db.First(&list, &TaskList{Title: title, UserID: user.ID})
	db.Delete(&Task{}, &Task{TaskListID: list.ID})
	db.Delete(&TaskList{}, &list)

	retToView(first, last, w, r)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if listPath.Match([]byte(r.URL.Path)) {
		addTaskHandler(w, r)
	} else {
		addListHandler(w, r)
	}
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	first, last, title := getNameList(r.URL.Path)

	var user User
	db.First(&user, &User{FirstName: first, LastName: last})

	var list TaskList
	db.First(&list, &TaskList{Title: title, UserID: user.ID})

	// make list associated with user
	db.Create(&Task{Title: r.FormValue("title"), DueDate: r.FormValue("due date"), Details: r.FormValue("details"), TaskListID: list.ID})

	retToView(first, last, w, r)
}

func addListHandler(w http.ResponseWriter, r *http.Request) {
	first, last := getName(r.URL.Path)

	var user User
	db.Find(&user, &User{FirstName: first, LastName: last})

	// make list associated with user
	db.Create(&TaskList{Title: r.FormValue("list title"), UserID: user.ID})

	retToView(first, last, w, r)
}

func markHandler(w http.ResponseWriter, r *http.Request) {
	first, last, title, taskTitle := getNameListTask(r.URL.Path)
	fmt.Printf("Path: %s\nF: %s\nL: %s\nTitle: %s\nTask: %s\n", r.URL.Path, first, last, title, taskTitle)

	var user User
	db.First(&user, &User{FirstName: first, LastName: last})

	var list TaskList
	db.First(&list, &TaskList{Title: title, UserID: user.ID})

	var task Task
	db.First(&task, &Task{Title: taskTitle, TaskListID: list.ID})

	task.Completed = !task.Completed

	db.Save(&task)

	retToView(first, last, w, r)
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "welcome.html")
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "tasks.css")
}

// load personalized view
var templates = template.Must(template.ParseFiles("tasks.html"))
var db = MustConnect()

func viewHandler(w http.ResponseWriter, r *http.Request) {
	type List struct {
		Title string
		Tasks []Task
	}

	type UserFile struct {
		Owner string
		Lists []List
	}

	first, last := getName(r.URL.Path)

	var uFile = UserFile{first + " " + last, nil}

	var user User
	db.Find(&user, &User{FirstName: first, LastName: last})

	var lists []TaskList
	db.Model(&user).Related(&lists) // all lists for user

	for _, tl := range lists {
		var tasks []Task
		db.Model(&tl).Related(&tasks)

		uFile.Lists = append(uFile.Lists, List{tl.Title, tasks})
	}

	err := templates.ExecuteTemplate(w, "tasks.html", uFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func retToView(first string, last string, w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/view/" + first + " " + last + "/"
	http.Redirect(w, r, r.URL.Path, http.StatusFound)
	viewHandler(w, r)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	first := r.FormValue("first name")
	last := r.FormValue("last name")

	var count int // add if absent (this is just for school)
	if db.Find(&User{FirstName: first, LastName: last}).Count(&count); count == 0 {
		db.Create(&User{FirstName: first, LastName: last})
		println("New User:", first, last) // DEBUG
	}

	retToView(first, last, w, r)
}

func missHandler(w http.ResponseWriter, r *http.Request) {
	println("we saw: ", r.URL.Path)
}

func main() {
	db := Example()
	defer db.Close()

	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/welcome/", welcomeHandler)
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/tasks.css", cssHandler)
	http.HandleFunc("/add/", addHandler)
	http.HandleFunc("/delete/", delHandler)
	http.HandleFunc("/mark complete/", markHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
