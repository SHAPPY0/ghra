package handlers

import (
	"log"
	"time"
	"net/http"
	"database/sql"
	"github.com/shappy0/ghra/utils"
	"github.com/shappy0/ghra/models"
)

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

func ProjectHandler(w http.ResponseWriter, r *http.Request) { 
	var db *sql.DB
	if db = (r.Context().Value("db")).(*sql.DB); db == nil {
		RenderErrorTemplate(w, http.StatusInternalServerError, "Invalid DB Connection")
		return
	}
	if r.Method == http.MethodGet {
		getProject(w, r, db)
	} else if r.Method == http.MethodDelete {
		deleteProject(w, r, db)
	}
}

func getProjectList(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	projects := make([]interface{}, 0)
	rows, err := db.Query(`SELECT * FROM projects_tbl`)
	if err != nil {
		RenderErrorTemplate(w, 500, err.Error())
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
	var reqBody models.ProjectReq
	if err := utils.GetBody(r.Body, &reqBody); err != nil {
		RenderErrorTemplate(w, http.StatusBadRequest, "Invalid request data")
		return
	}
	query := "INSERT INTO projects_tbl(name, description, active, createdAt, updatedAt) VALUES(?, ?, ?, ?, ?)"
	result, err := db.Exec(query, 
		reqBody.Name,
		reqBody.Description,
		true,
		time.Now(),
		nil,
	)
	if err != nil {
		RenderErrorTemplate(w, http.StatusInternalServerError, err.Error())
		return
	}
	count, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if count == 1 {
		Response(w, http.StatusOK, "Project created successfully", nil)
	} else {
		RenderErrorTemplate(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	
}

func getProject(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var project models.Project
	projectId := r.PathValue("id") 
	if projectId == "" {
		RenderErrorTemplate(w, http.StatusBadRequest, "Invalid ProjectId")
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
		RenderErrorTemplate(w, http.StatusInternalServerError, err.Error())
		return
	}
	data := map[string]interface{}{
		"project": project,
		"repositories": repos,
	}
	RenderTemplate(w, "project", data)
}

func deleteProject(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	projectId := r.PathValue("id") 
	if projectId == "" {
		RenderErrorTemplate(w, http.StatusBadRequest, "Invalid ProjectId")
		return
	}
	db.Exec(`DELETE FROM projects_tbl WHERE id = ?`, projectId)
	Response(w, 200, "Project deleted successfully", nil)
}