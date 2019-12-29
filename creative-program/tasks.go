package main

/*
	This file implements a RESTful web app for keeping track of To-Do items
	using Go.

	In order for this code to run it must connect to a server. It is currently
	configured to connect to a demo server on my local machine. Golang.org has
	a sql package that abstracts details about drivers. However, because I was
	using an ORM DB I sometimes used other methods.

	NOTE: This web app is for demonstration only. It has no protection measures
	beyond simple URL validation in place.

	I attempted to follow golang.org's recommendations for documentation and
	style. Namely, the doc comment for each method starts with the method name
	and is concise.

	In order to learn how to implement each of the components of this
	application I followed a number of independent tutorials which I highly
	recomend (below).

	| name                        | url                                                |
	| --------------------------- | -------------------------------------------------- |
	| SQL Server, Windows, Go sql | https://www.microsoft.com/en-us/sql-server/developer-get-started/go/windows |
	| Writing Web Applications    | https://golang.org/doc/articles/wiki/ |
	| A Tour of Go                | https://tour.golang.org/ |
	| Using Templates             | https://blog.gopheracademy.com/advent-2017/using-go-templates/ |

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
	## Model:
	The following section of code declares datatypes and implements access to
	storage on a database.

	Connection to the database (and conversions of data) is aided by the gorm
	package from github.com/jinzhu/gorm. This popular package helps implement
	an Object-Relational Mapping (ORM) style of database. I stuck to the more
	traditional database style in my implementation.

	Each struct is converted by gorm into a corresponding table in the
	database. The name of the table is the lower_snake_case plural name of the
	struct. Each field corresponds to a lower_snake_case column in the table
	(except the gorm.Model which becomes a number of rows with ID, last edit
	date, and etc.).
*/

type (
	// User is a named owner of lists
	User struct {
		gorm.Model
		FirstName string `gorm:"primary_key"`
		LastName  string `gorm:"primary_key"`
	}

	// Task is a to-do item
	Task struct {
		gorm.Model
		Title      string `gorm:"primary_key"`
		Details    string
		DueDate    string
		Completed  bool
		TaskListID uint
	}

	// TaskList is named set of tasks
	TaskList struct {
		gorm.Model
		Title  string `gorm:"primary_key"`
		UserID uint
	}
)

// MustConnect connects to my local SampleDB
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

// Example connects to my local DB and resets tables (populating them with test
// data)
func Example() *gorm.DB {
	db := MustConnect()
	defer db.Close()

	fmt.Println("Reseting DB...")
	db.DropTableIfExists(&User{}, &TaskList{}, &Task{})

	fmt.Println("Migrating models...")
	db.AutoMigrate(&User{})
	db.AutoMigrate(&TaskList{})
	db.AutoMigrate(&Task{})

	// Create test Users
	fmt.Println("Creating users...")
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

	// Create  Tasks for each user
	fmt.Println("Creating new tasks...")
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
	## View
	This next section of code is about presentation to the user (via
		implementation of a web server).

	The html package provides many methods for serving content, but none of
	these methods are encrytped. For encryption one is able to change html
	out for the net package which uses TLS 1.3. Following the extra steps for
	TLS was beyond the scope of this assignment.

	| endpoint                              | purpose                        |
	| ------------------------------------- | ------------------------------ |
	| /welcome	                            | user's first point of contact  |
	| /login                                | finds or lists user in DB      |
	| /view/firstname lastname              | display's user's to-do lists   |
	| /add/firstname lastname               | request to add a list          |
	| /delete/firstname lastname            | request to delete a list       |
	| /add/firstname lastname/list          | request to add task to list    |
	| /delete/firstname lastname/list     | request to delete task from list |
	| /mark/firstname lastname/list/task  | toggle's the .Completed field of a task |
	NOTE: the server will be live at localhost:8080

	## RegEx
	Most open-source implementations of RegEx tend to be very slow (including
	Python's). However, Go's is much faster because it creates a digraph and
	iteratively searches for matches instead of recursing. [I think it's very
	interesting.](https://swtch.com/~rsc/regexp/regexp1.html)

	It's best to compile each Regular Expression only once, so it's an accepted
	idiom to have them as global variables (this also allows sharing). Notice
	that these are inherantly constants.
*/

// Compile each Regular Expression only once for efficiency
// All urls should match this path, but parts don't share consistent positions.
// TODO: improve security by checking for a match with this path.
var validPath = regexp.MustCompile("^/(welcome|login)|((view|add|delete)/(\\w+) (\\w+))|((add|delete)/(\\w+) (\\w+)/([^/]+))|((mark)/(\\w+) (\\w+)/([^/]+)/([^/]+))$")

// useful for parsing name (trailing optional parts could disqualify a match)
var userPath = regexp.MustCompile("^/(view|add|delete)/(\\w+) (\\w+)")

// getName get's the first and last name from a request
// Returns emtpy strings if invalid.
func getName(url string) (first, last string) {
	m := userPath.FindStringSubmatch(url)

	if m != nil {
		return m[2], m[3]
	}
	return "", ""
}

// useful for parsing name and list title
var listPath = regexp.MustCompile("^/(view|add|delete)/(\\w+)\\s(\\w+)/([^/]+)")

// getNameList parses first name, last name, and list title.
// Returns emtpy strings if invalid.
func getNameList(url string) (first, last, list string) {
	m := listPath.FindStringSubmatch(url)
	if m != nil {
		return m[2], m[3], m[4]
	}
	return "", "", ""
}

// useful for parsing entire path (name, list, task)
var taskPath = regexp.MustCompile("^/(view|add|delete|mark)/(\\w+)\\s(\\w+)/([^/]+)/([^/]+)")

// getNameListTask parses path for info.
// Returns emtpy strings if invalid.
func getNameListTask(url string) (first, last, list, task string) {
	m := taskPath.FindStringSubmatch(url)
	if m != nil {
		return m[2], m[3], m[4], m[5]
	}
	return "", "", "", ""
}

// the following handlers rely on this connection to operate
var db = MustConnect()

// delHandler Delegates delete requests by path contents.
func delHandler(w http.ResponseWriter, r *http.Request) {
	if taskPath.Match([]byte(r.URL.Path)) {
		delTaskHandler(w, r)
	} else {
		delListHandler(w, r)
	}
}

// delTaskHandler Deletes a task from a user's list (DB).
// Redirects user to the updated view.
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

// delListHandler Deletes a user's list and all tasks within (DB).
// Redirects user to the updated view.
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

// addHandler Delegates add requests by path contents.
func addHandler(w http.ResponseWriter, r *http.Request) {
	if listPath.Match([]byte(r.URL.Path)) {
		addTaskHandler(w, r)
	} else {
		addListHandler(w, r)
	}
}

// addTaskHandler Adds a task to a user's list (DB).
// Redirects user to the updated view.
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

// addListHandler Creates a new list associated with the user (DB).
// Redirects user to the updated view.
func addListHandler(w http.ResponseWriter, r *http.Request) {
	first, last := getName(r.URL.Path)

	var user User
	db.Find(&user, &User{FirstName: first, LastName: last})

	// make list associated with user
	db.Create(&TaskList{Title: r.FormValue("list title"), UserID: user.ID})

	retToView(first, last, w, r)
}

// markHandler toggles the is/isn't complete status of a user's task.
// Redirects user to the updated view.
func markHandler(w http.ResponseWriter, r *http.Request) {
	first, last, title, taskTitle := getNameListTask(r.URL.Path)
	fmt.Printf("Path: %s\nF: %s\nL: %s\nTitle: %s\nTask: %s\n", r.URL.Path, first, last, title, taskTitle)

	var user User
	db.First(&user, &User{FirstName: first, LastName: last})

	var list TaskList
	db.First(&list, &TaskList{Title: title, UserID: user.ID})

	var task Task
	db.First(&task, &Task{Title: taskTitle, TaskListID: list.ID})

	task.Completed = !task.Completed // toggle boolean

	db.Save(&task)

	retToView(first, last, w, r)
}

// welcomeHandler Serves a static login webpage to the client.
func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "welcome.html")
}

// loginHandler Find's the user's table or makes a new one.
// Redirects user to the updated view.
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

// cssHandler serves the css for this app to the client.
func cssHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "tasks.css")
}

/*
	## templates
	Golang's template package provides powerful tools for templating text and
	html.

	By execution the template a user-specific page is generated.

	Like the Regular Expressions it's best to only parse the template once.
*/

// provides view of a user's task lists
var templates = template.Must(template.ParseFiles("tasks.html"))

// viewHandler executes templates with the user's data.
func viewHandler(w http.ResponseWriter, r *http.Request) {
	type List struct {
		Title string
		Tasks []Task
	}

	// a temp struct for organizing a user's collective information
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
		db.Model(&tl).Related(&tasks) // all tasks for list

		uFile.Lists = append(uFile.Lists, List{tl.Title, tasks})
	}

	err := templates.ExecuteTemplate(w, "tasks.html", uFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// retToView redirects user to the updated view.
// Alters path to ensure template has valid links.
func retToView(first string, last string, w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/view/" + first + " " + last + "/"
	http.Redirect(w, r, r.URL.Path, http.StatusFound)
	viewHandler(w, r)
}

// Lauches server with all handlers.
// Uses port 8080
func main() {
	db := Example()
	defer db.Close()

	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/welcome/", welcomeHandler)
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/tasks.css", cssHandler)
	http.HandleFunc("/add/", addHandler)
	http.HandleFunc("/delete/", delHandler)
	http.HandleFunc("/mark/", markHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
