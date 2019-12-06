package tasks

import (
	"html/template"
	"log"
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

	now := time.Now()
	err := templ.Execute(f,
		UserFile{
			Owner: User{
				Model:     gorm.Model{0, now, now, &now},
				FirstName: "Ivan",
				LastName:  "Webber",
			},
			Lists: []TaskList{
				TaskList{
					Title: "Make this work",
					List: []Task{
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
		log.Fatal(err.Error())
		t.Error("Test Template: ", err.Error())
	}
}
