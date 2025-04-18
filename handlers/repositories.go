package handlers

import (
	"log"
	"fmt"
	"time"
	"context"
	"net/http"
	"database/sql"
	"github.com/shappy0/ghra/utils"
	"github.com/shappy0/ghra/models"
	"github.com/shappy0/ghra/github"
	"github.com/vifraa/gopom"
)

func RepositoriesHandler(w http.ResponseWriter, r *http.Request) {
	var db *sql.DB
	if db = (r.Context().Value("db")).(*sql.DB); db == nil {
		http.Error(w, "Invalid DB connection", http.StatusInternalServerError)
		return
	}
	if utils.CheckRoute(r.URL.Path, `/project/\d+/repositories`) {
		log.Println("Matched")
	} else if utils.CheckRoute(r.URL.Path, `/repositories/cascade/(.+?)/deps`) {
		if r.Method == http.MethodGet {
			cascadeType := r.PathValue("type")
			if cascadeType == "update" {
				renderDepsCascadeUpdate(w, r, db)
			} else if cascadeType == "add" {
				renderDepsCascadeAdd(w, r, db)
			}
		}
	} else if utils.CheckRoute(r.URL.Path, `^/repository/\d+/\d+$`) {
		if r.Method == http.MethodGet {
			getRepository(w, r, db)	 
		} else if r.Method == http.MethodDelete {
			deleteRepository(w, r, db)
		}
	} else if utils.CheckRoute(r.URL.Path, `/repositories/cascade/bfr`) {
		if r.Method == http.MethodGet {
			renderCascadeBfr(w, r, db)
		} else if r.Method == http.MethodPost {
			createBranchFromRelease(w, r, db)
		}
	} else if utils.CheckRoute(r.URL.Path, `/repositories/release-tags`) {
		if r.Method == http.MethodGet {
			getReleaseTags(w, r, db)
		}
	} else {
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
}

func getRepositoryList(w http.ResponseWriter, r *http.Request, db *sql.DB, projectId int) ([]models.Repository, error) {
	repositories := make([]models.Repository, 0)
	query := `SELECT id, projectId, name, url, branch, buildTool, depFilePath, user, token, tags, active, createdAt, updatedAt 
				FROM repositories_tbl 
				WHERE projectId = ?
				ORDER BY createdAt desc`
	rows, err := db.Query(query, projectId)
	if err != nil {
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
			&repo.DepFilePath,
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
	var reqBody models.RepoReq
	if err := utils.GetBody(r.Body, &reqBody); err != nil {
		log.Println(err.Error())
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}
	result, err := db.Exec("INSERT INTO repositories_tbl(projectId, name, url, branch, buildTool, depFilePath, tags, user, token, active, createdAt, updatedAt) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", 
		reqBody.ProjectId,
		reqBody.Name,
		reqBody.Url,
		reqBody.Branch,
		reqBody.BuildTool,
		reqBody.DepFilePath,
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

func getDependencies(db *sql.DB, repoId, projectId int) (*models.RepoDeps, error) {
	ctx := context.Background()
	var repository models.Repository
	var repoDeps models.RepoDeps
	if repoId == 0 || projectId == 0 {
		return nil, fmt.Errorf("RepoId and ProjectId required")
	}
	query := `SELECT id, projectId, name, url, branch, user, token, tags, buildTool, depFilePath, active, createdAt, updatedAt 
				FROM repositories_tbl WHERE id = ? AND projectId = ?`
	db.QueryRow(query, repoId, projectId).Scan(
		&repository.Id,
		&repository.ProjectId,
		&repository.Name,
		&repository.Url,
		&repository.Branch,
		&repository.User,
		&repository.Token,
		&repository.Tags,
		&repository.BuildTool,
		&repository.DepFilePath,
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
	repoDeps = models.RepoDeps{
		Name: 			repository.Name,
		RepoId: 		repository.Id,
		ProjectId: 		repository.ProjectId,
		Content: 		*fileObj.Content,
		Branch: 		repository.Branch,
		DepFilePath:	repository.DepFilePath,
		SHA: 			fileObj.SHA,
		LinedContent: 	make(map[int]string, 0),
		Dependencies: 	parsedContent.Dependencies,
		Properties: 	parsedContent.Properties,
		Parent:			parsedContent.Parent,
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
		Branch:			repo.Branch,
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
		} else if r.Method == http.MethodPut {
			pushVCDependencies(w, r, db)
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
	query := `SELECT id, projectId, name, url, branch, user, token, tags, buildTool, depFilePath, active, createdAt, updatedAt 
					FROM repositories_tbl WHERE id = ? AND projectId = ?`
	db.QueryRow(query, reqBody.RepoId, reqBody.ProjectId).Scan(
		&repository.Id,
		&repository.ProjectId,
		&repository.Name,
		&repository.Url,
		&repository.Branch,
		&repository.User,
		&repository.Token,
		&repository.Tags,
		&repository.BuildTool,
		&repository.DepFilePath,
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
		log.Println("Error: " + err.Error())
		ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var updatedDeps string
	if repository.BuildTool == "maven" {
		updatedDeps, err = github.ModifyDeps(*fileObj.Content, reqBody.Content)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
			return
		}
		if updatedDeps == "" {
			ErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
			return
		}
	}

	done, err := github.PushChanges(ctx, repository, fileObj, reqBody.Message, updatedDeps)
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
	var reqBody models.VCDepsReq
	var repoDeps []models.RepoDeps
	var commonParent *gopom.Parent
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

		//filter common parent
		isParentCommon := true
		baseParent := repoDeps[0].Parent
		if baseParent != nil {
			for _, repoDep := range repoDeps[1:] {
				parent := repoDep.Parent
				if parent != nil {
					if *baseParent.GroupID != *parent.GroupID || *baseParent.ArtifactID != *parent.ArtifactID || *baseParent.Version != *parent.Version {
						isParentCommon = false
					}
				} else {
					isParentCommon = false
				}
			}
		} else {
			isParentCommon = false
		}
		if isParentCommon {
			commonParent = baseParent
		}

		//filter common properties
		for k, v := range repoDeps[0].Properties.Entries {
			isCommon := true
			for _, repoDep := range repoDeps[1:] {
				props := repoDep.Properties.Entries
				if _, exists := props[k]; !exists {
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
			isCommon := true
			for _, repoDep1 := range repoDeps[1:] {
				matched := false
				for _, dep := range *repoDep1.Dependencies {
					if *repoDep.GroupID == *dep.GroupID && *repoDep.ArtifactID == *dep.ArtifactID {
						matched = true
						break
					}
				}
				if !matched {
					isCommon = false
				}
			} 
			if isCommon {
				commonDeps = append(commonDeps, repoDep)
			}
		}

		data := map[string]interface{} {
			"ProjectId": reqBody.ProjectId,
			"RepoIds": reqBody.RepoIds,
			"Parent": commonParent,
			"Properties": commonProps,
			"Dependencies": commonDeps,
		}
		Response(w, http.StatusOK, "", data)
	} else {
		log.Println("Error: No repositories selected to update")
		ErrorResponse(w, http.StatusInternalServerError, "No repositories selected to update", nil)
		return
	}
}

func pushVCDependencies(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := context.Background()
	var reqBody []models.CommitReq
	var errored map[string]string
	
	if err := utils.GetBody(r.Body, &reqBody); err != nil {
		log.Println(err.Error())
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	for i := 0; i < len(reqBody); i++ {
		var repository models.Repository
		db.QueryRow(`SELECT id, projectId, name, url, branch, user, token, tags, buildTool, depFilePath, active, createdAt, updatedAt 
					FROM repositories_tbl WHERE id = ? AND projectId = ?`, 
					reqBody[i].RepoId, reqBody[i].ProjectId).Scan(
			&repository.Id,
			&repository.ProjectId,
			&repository.Name,
			&repository.Url,
			&repository.Branch,
			&repository.User,
			&repository.Token,
			&repository.Tags,
			&repository.BuildTool,
			&repository.DepFilePath,
			&repository.Active,
			&repository.CreatedAt,
			&repository.UpdatedAt,
		)
		if repository.Url == "" || repository.User == "" || repository.Token == "" {
			errored[repository.Name] = "Invalid repo credentials"
 			// ErrorResponse(w, http.StatusInternalServerError, "Something went wrong", nil)
			continue
		}
		fileObj, err := readDepFileContent(ctx, repository)
		if err != nil {
			errored[repository.Name] = err.Error()
			continue
		}
		var updatedDeps string
		if repository.BuildTool == "maven" {
			updatedDeps, err = github.ModifyDeps(*fileObj.Content, reqBody[i].Content)
			if err != nil {
				errored[repository.Name] = err.Error()
				continue
			}
			if updatedDeps == "" {
				errored[repository.Name] = "Deps content is corrupted"
				continue
			}
		}
		_, err = github.PushChanges(ctx, repository, fileObj, reqBody[i].Message, updatedDeps)
		if err != nil {
			errored[repository.Name] = err.Error()
			continue
		}
	}
	if len(errored) > 0 {
		ErrorResponse(w, http.StatusInternalServerError, "Error: All/Partial update failed", errored)
		return
	} else {
		Response(w, http.StatusOK, "Repo versions updated!!!", len(reqBody))
	}
}

func getRepository(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := context.Background()
	result := map[string]interface{}{}
	projectId := r.PathValue("projectId")
	repoId := r.PathValue("repoId")

	if projectId == "" || repoId == "" {
		ErrorResponse(w, http.StatusBadRequest, "RepoId/ProjectId required", nil)
		return
	}
	repository := getRepositoryInfo(db, projectId, repoId)
	branches, err := github.GetBranches(ctx, repository)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	result["repository"] = repository
	result["branches"] = branches
	Response(w, http.StatusOK, "OK", result)
}

func getRepositoryInfo(db *sql.DB, projectId, repoId string) models.Repository {
	var repository models.Repository
	query := `SELECT id, projectId, name, url, branch, user, token, tags, buildTool, depFilePath, active, createdAt, updatedAt FROM repositories_tbl WHERE id = ? AND projectId = ?`;
	db.QueryRow(query, repoId, projectId).Scan(&repository.Id,
		&repository.ProjectId,
		&repository.Name,
		&repository.Url,
		&repository.Branch,
		&repository.User,
		&repository.Token,
		&repository.Tags,
		&repository.BuildTool,
		&repository.DepFilePath,
		&repository.Active,
		&repository.CreatedAt,
		&repository.UpdatedAt,
	)
	return repository
}

func deleteRepository(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	projectId := r.PathValue("projectId")
	repoId := r.PathValue("repoId")
	if projectId == "" || repoId == "" {
		ErrorResponse(w, http.StatusBadRequest, "RepoId/ProjectId required", nil)
		return
	}
	db.Exec(`DELETE FROM repositories_tbl WHERE id = ? AND projectId = ?`, repoId, projectId)
	Response(w, 200, "Repository deleted successfully", nil)
}

func renderDepsCascadeUpdate(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	projectId := r.URL.Query().Get("projectId")
	if projectId == "" {
		ErrorResponse(w, http.StatusInternalServerError, "Error: Invalid projectId", nil)
		return
	}
	projId := utils.StrToInt(projectId)
	var projectName string
	query := `SELECT name FROM projects_tbl WHERE id = ?`
	db.QueryRow(query, projectId).Scan(&projectName)
	repoList, err := getRepositoryList(w, r, db, projId)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "Error: Err getting repo lists", nil)
		return
	}
	data := map[string]interface{}{
		"projectId": projectId,
		"projectName": projectName,
		"repositories": repoList,
	}
	RenderTemplate(w, "cascadingUpdate", data)
}

func renderDepsCascadeAdd(w http.ResponseWriter, r *http.Request, db *sql.DB) {

}