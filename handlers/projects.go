package handlers

import (
	"log"
	"time"
	"net/http"
	"database/sql"
	"github.com/shappy0/ghra/utils"
	"github.com/shappy0/ghra/models"
)

type ProjectReq struct {
	Name 	string
	Description string
}

func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	var db *sql.DB
	if db = (r.Context().Value("db")).(*sql.DB); db == nil {
		http.Error(w, "Invalid DB connection", http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodGet {
		getProjectList(w, r, db)
	} else if r.Method == http.MethodPost {
		createProject(w, r, db)
	}
}

func getProjectList(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	projects := make([]interface{}, 0)
	rows, err := db.Query(`SELECT * FROM projects_tbl`)
	if err != nil {
		log.Printf("err" + err.Error())
		return
	}
	for rows.Next() {
		var id int
		var name string
		var description string
		var active bool
		var createdAt string
		var updatedAt string
		rows.Scan(&id, &name, &description, &active, &createdAt, &updatedAt)
		projects = append(projects, map[string]interface{}{
			"id": id,
			"name": name,
			"description": description,
			"active": active,
			"createdAt": createdAt,
			"updatedAt": updatedAt,
		})
	}
	data := map[string]interface{}{
		"projects": projects,
	}
	RenderTemplate(w, "projects", data)
}

func createProject(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var reqBody ProjectReq
	if err := utils.GetBody(r.Body, &reqBody); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}
	result, err := db.Exec("INSERT INTO projects_tbl(name, description, active, createdAt, updatedAt) VALUES(?, ?, ?, ?, ?)", 
		reqBody.Name,
		reqBody.Description,
		true,
		time.Now(),
		nil,
	)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	count, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if count == 1 {
		Response(w, http.StatusOK, "Project created successfully", nil)
	} else {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	
}

func ProjectHandler(w http.ResponseWriter, r *http.Request) { 
	var db *sql.DB
	if db = (r.Context().Value("db")).(*sql.DB); db == nil {
		http.Error(w, "Invalid DB connection", http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodGet {
		getProject(w, r, db)
	}
}

func getProject(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var project models.Project
	projectId := r.PathValue("id") 
	if projectId == "" {
		http.Error(w, "Invalid projectId", http.StatusBadRequest)
		return
	}
	db.QueryRow(`SELECT * FROM projects_tbl WHERE id = ?`, projectId).Scan(&project.Id, 
			&project.Name, 
			&project.Description, 
			&project.Active, 
			&project.CreatedAt, 
			&project.UpdatedAt)
	repos, err := getRepositoryList(w, r, db, project.Id)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"project": project,
		"repositories": repos,
	}
	RenderTemplate(w, "project", data)
}