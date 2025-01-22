package handlers

import (
	"log"
	"fmt"
	"time"
	"context"
	"strings"
	"net/http"
	"database/sql"
	"github.com/shappy0/ghra/utils"
	"github.com/shappy0/ghra/models"
	"github.com/shappy0/ghra/github"
	"github.com/vifraa/gopom"
	// "github.com/google/go-github/v68/github"
)

type RepoReq struct {
	ProjectId	int
	Name 		string
	Url			string
	Branch		string
	User		string
	Token		string
	Tags		string
	BuildTool	string
	DepFileName	string
}

type VCDepsReq struct {
	RepoIds []int
	ProjectId int
}

type RepoDeps struct {
	Name		string
	ProjectId	int
	RepoId 		int
	Branch 		string
	Content		string
	SHA			string
	Properties	*gopom.Properties
	Dependencies *[]gopom.Dependency
	LinedContent map[int]string
	DepHashed 	map[string]string
}

func RepositoriesHandler(w http.ResponseWriter, r *http.Request) {
	var db *sql.DB
	if db = (r.Context().Value("db")).(*sql.DB); db == nil {
		http.Error(w, "Invalid DB connection", http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodGet {
		id := r.PathValue("id")
		projectId := r.URL.Query().Get("projectId")
		repoDeps, err := getDependencies(db, github.StrToInt(id), github.StrToInt(projectId))
		if err != nil {
			log.Println("Error: " + err.Error())
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
			return
		}
		RenderTemplate(w, "deps", *repoDeps)
	} else if r.Method == http.MethodPost {
		createRepository(w, r, db)
	}
}

func getRepositoryList(w http.ResponseWriter, r *http.Request, db *sql.DB, projectId int) ([]models.Repository, error) {
	repositories := make([]models.Repository, 0)
	rows, err := db.Query(`SELECT id, projectId, name, url, branch, buildTool, depFileName, user, token, tags, active, createdAt, updatedAt FROM repositories_tbl where projectId = ?`, projectId)
	if err != nil {
		log.Printf("err" + err.Error())
		return repositories, err
	}
	for rows.Next() {
		var repo models.Repository
		rows.Scan(&repo.Id,
			&repo.ProjectId,
			&repo.Name, 
			&repo.Url, 
			&repo.Branch,
			&repo.BuildTool,
			&repo.DepFileName,
			&repo.User, 
			&repo.Token, 
			&repo.Tags, 
			&repo.Active, 
			&repo.CreatedAt, 
			&repo.UpdatedAt)
		repositories = append(repositories, repo)
	}
	
	return repositories, nil
}

func createRepository(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var reqBody RepoReq
	if err := utils.GetBody(r.Body, &reqBody); err != nil {
		log.Println(err.Error())
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}
	result, err := db.Exec("INSERT INTO repositories_tbl(projectId, name, url, branch, buildTool, depFileName, tags, user, token, active, createdAt, updatedAt) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", 
		reqBody.ProjectId,
		reqBody.Name,
		reqBody.Url,
		reqBody.Branch,
		reqBody.BuildTool,
		reqBody.DepFileName,
		reqBody.Tags,
		reqBody.User,
		reqBody.Token,
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
		Response(w, http.StatusOK, "Repository addedd successfully", nil)
	} else {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	
}

func getDependencies(db *sql.DB, repoId, projectId int) (*RepoDeps, error) {
	ctx := context.Background()
	var repository models.Repository
	var repoDeps RepoDeps
	if repoId == 0 || projectId == 0 {
		return nil, fmt.Errorf("RepoId and ProjectId required")
	}
	
	db.QueryRow(`SELECT id, projectId, name, url, branch, user, token, tags,buildTool, depFileName, active, createdAt, updatedAt FROM repositories_tbl WHERE id = ? AND projectId = ?`, repoId, projectId).Scan(
		&repository.Id,
		&repository.ProjectId,
		&repository.Name,
		&repository.Url,
		&repository.Branch,
		&repository.User,
		&repository.Token,
		&repository.Tags,
		&repository.BuildTool,
		&repository.DepFileName,
		&repository.Active,
		&repository.CreatedAt,
		&repository.UpdatedAt,
	)
	if repository.Url == "" || repository.User == "" || repository.Token == "" {
		return nil, fmt.Errorf("Bad Request")
	}
	//Read dep file content
	fileObj, err := readDepFileContent(ctx, repository)
	if err != nil {
		return nil, err 
	}
	//Parse POM xml file
	parsedContent, err := github.Parse(*fileObj.Content)
	if err != nil {
		return nil, err
	}
	repoDeps = RepoDeps{
		Name: 			repository.Name,
		RepoId: 		repository.Id,
		ProjectId: 		repository.ProjectId,
		Content: 		*fileObj.Content,
		Branch: 		repository.Branch,
		SHA: 			fileObj.SHA,
		LinedContent: 	make(map[int]string, 0),
		Dependencies: 	parsedContent.Dependencies,
		Properties: 	parsedContent.Properties,
	}
	contentSeg := strings.Split(repoDeps.Content, "\n")
	for i := 0; i < len(contentSeg); i++ {
		repoDeps.LinedContent[i] = contentSeg[i]
	}
	return &repoDeps, nil 
}

func readDepFileContent(ctx context.Context, repo models.Repository) (*github.GhraFile, error) {
	fileObj, err := github.ReadDepFile(ctx, repo)
	if err != nil {
		return nil, err
	}
	
	//Get dep file content
	content, err := fileObj.GetContent()
	if err != nil {
		return nil, err
	}
	GhraFile := github.GhraFile{
		RepoContent:	fileObj,
		Content:		&content,
		SHA:			fileObj.GetSHA(),
	}
	return &GhraFile, nil 
}

func DependenciesHandler(w http.ResponseWriter, r *http.Request) {
	var db *sql.DB
	if db = (r.Context().Value("db")).(*sql.DB); db == nil {
		ErrorResponse(w, http.StatusInternalServerError, "Invalid DB connection", nil)
		return
	}
	if r.URL.Path == "/vc/deps" {
		if r.Method == http.MethodPost {
			getVCDependencies(w, r, db)
		}
	} else if r.URL.Path == "/deps" {
		if r.Method == http.MethodPut{
			pushChanges(w, r, db)
		}
	}
}

func pushChanges(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := context.Background()
	var reqBody models.CommitReq
	var repository models.Repository
	if err := utils.GetBody(r.Body, &reqBody); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Bad Request", nil)
		return
	}
	db.QueryRow(`SELECT id, projectId, name, url, branch, user, token, tags, buildTool, depFileName, active, createdAt, updatedAt 
					FROM repositories_tbl WHERE id = ? AND projectId = ?`, 
					reqBody.RepoId, reqBody.ProjectId).Scan(
		&repository.Id,
		&repository.ProjectId,
		&repository.Name,
		&repository.Url,
		&repository.Branch,
		&repository.User,
		&repository.Token,
		&repository.Tags,
		&repository.BuildTool,
		&repository.DepFileName,
		&repository.Active,
		&repository.CreatedAt,
		&repository.UpdatedAt,
	)
	if repository.Url == "" || repository.User == "" || repository.Token == "" {
		ErrorResponse(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	
	fileObj, err := readDepFileContent(ctx, repository)
	if err != nil {
		log.Println("Errrrrrrror:" + err.Error())
		ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var updatedDeps string
	if repository.BuildTool == "maven" {
		updatedDeps, err = github.ModifyDeps(*fileObj.Content, reqBody.NewContent)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
			return
		}
		if updatedDeps == "" {
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}

	done, err := github.PushChanges(ctx, repository, reqBody, updatedDeps)
	if err != nil {
		log.Println("Error:" + err.Error())
		ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return	
	}
	if done {
		Response(w, http.StatusOK, "Changes pushed", nil)
	}
}

func getVCDependencies(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var reqBody VCDepsReq
	var repoDeps []RepoDeps
	var commonProps map[string]string
	var commonDeps []gopom.Dependency

	if err := utils.GetBody(r.Body, &reqBody); err != nil {
		log.Println(err.Error())
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if len(reqBody.RepoIds) > 0 {
		for _, id := range reqBody.RepoIds {
			repoDep, err := getDependencies(db, id, reqBody.ProjectId)
			if err != nil {
				log.Println("Error: " + err.Error())
				ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
				return
			}
			repoDeps = append(repoDeps, *repoDep)
		}

		commonProps = make(map[string]string, 0)
		commonDeps = make([]gopom.Dependency, 0)

		//filter common properties
		for k, v := range repoDeps[0].Properties.Entries {
			isCommon := true
			for _, repoDep := range repoDeps[1:] {
				props := repoDep.Properties.Entries
				if vv, exists := props[k]; !exists || vv != v {
					isCommon = false
					break
				}
			}
			if isCommon {
				commonProps[k] = v
			}
		}
		//filter common dependencies
		for _, repoDep := range *repoDeps[0].Dependencies {
			isCommon := false
			for _, repoDep1 := range repoDeps[1:] {
				for _, dep := range *repoDep1.Dependencies {
					if *repoDep.GroupID == *dep.GroupID && *repoDep.ArtifactID == *dep.ArtifactID {
						isCommon = true
						break
					}
				}
			} 
			if isCommon {
				commonDeps = append(commonDeps, repoDep)
			}
		}
		data := map[string]interface{} {
			"RepoIds": reqBody.RepoIds,
			"Properties":	commonProps,
			"Dependencies": commonDeps,
		}
		Response(w, http.StatusOK, "", data)
	} else {
		log.Println("Error: No repositories selected to update")
		ErrorResponse(w, http.StatusInternalServerError, "No repositories selected to update", nil)
		return
	}
}