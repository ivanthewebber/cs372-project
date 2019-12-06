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
	FirstName string
	LastName  string
}

// Task is a to-do item
type Task struct {
	gorm.Model
	Title      string
	Details    string
	DueDate    string
	Completed  bool
	TaskListID uint
}

// TaskList is named set of tasks
type TaskList struct {
	gorm.Model
	Title  string
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

var validPath = regexp.MustCompile("^/(welcome|login)|((view|add|delete)/(\\w{,32}) (\\w{,32}))|((add|delete)/(\\w{,32}) (\\w{,32})/([A-Za-z0-9]{,128}))|((mark complete)/(\\w{,32}) (\\w{,32})/([A-Za-z0-9]{,128})/([A-Za-z0-9]{,256}))$")

var userPath = regexp.MustCompile("^/(view|add|delete)/(\\w{,32})\\s(\\w{,32})")

func getName(url string) (first, last string) {
	m := userPath.FindStringSubmatch(url)
	if m != nil {
		return m[2], m[3]
	}
	return "", ""
}

var listPath = regexp.MustCompile("^/(view|add|delete)/(\\w{,32})\\s(\\w{,32})/([A-Za-z0-9\\s]{,128})")

func getNameList(url string) (first, last, list string) {
	m := userPath.FindStringSubmatch(url)
	if m != nil {
		return m[2], m[3], m[4]
	}
	return "", "", ""
}

var taskPath = regexp.MustCompile("^/(view|add|delete)/(\\w{,32})\\s(\\w{,32})/([A-Za-z0-9\\s]{,128})/([A-Za-z0-9\\s]{,256})$")

func getNameListTask(url string) (first, last, list, task string) {
	m := userPath.FindStringSubmatch(url)
	if m != nil {
		return m[2], m[3], m[4], m[5]
	}
	return "", "", "", ""
}

// find and pass requested page
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
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
	db2 := db.Find(&User{FirstName: first, LastName: last})
	db2.First(&user)

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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	first := r.FormValue("first name")
	last := r.FormValue("last name")

	var count int // add if absent (this is just for school)
	if db.Find(&User{FirstName: first, LastName: last}).Count(&count); count == 0 {
		db.Create(&User{FirstName: first, LastName: last})
		println("New User:", first, last) // DEBUG
	}

	http.Redirect(w, r, "/view/"+first+" "+last+"/", http.StatusFound)
	viewHandler(w, r)
}

func main() {
	db := MustConnect()
	defer db.Close()

	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/welcome/", welcomeHandler)
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/tasks.css", cssHandler) // FIXME

	log.Fatal(http.ListenAndServe(":8080", nil))
}
