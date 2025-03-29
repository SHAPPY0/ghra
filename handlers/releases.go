package handlers

import (
	"fmt"
	"log"
	"strings"
	"context"
	"net/http"
	"database/sql"
	"github.com/shappy0/ghra/github"
	"github.com/shappy0/ghra/utils"
	"github.com/shappy0/ghra/models"
)

func renderCascadeBfr(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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
		ErrorResponse(w, http.StatusInternalServerError, "Error: Getting repo lists", nil)
		return
	}
	data := map[string]interface{}{
		"projectId": projectId,
		"projectName": projectName,
		"repositories": repoList,
	}
	RenderTemplate(w, "releaseBranch", data)
}

func getReleaseTags(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := context.Background()
	var repoIds []string
	var releases []*models.RepoRelease

	projectId := r.URL.Query().Get("projectId")
	if projectId == "" {
		ErrorResponse(w, http.StatusBadRequest, "Error: Invalid projectId", nil)
		return
	}
	repoIdsStr := r.URL.Query().Get("repoIds")
	if repoIdsStr != "" {
		repoIds = strings.Split(repoIdsStr, ",")
	}
	for i := 0; i < len(repoIds); i++ {
		repo := getRepositoryInfo(db, projectId, repoIds[i])
		if &repo != nil {
			tags, err := github.GetReleaseTags(ctx, repo)
			if err != nil {
				fmt.Println("Error while getting release " + err.Error())
			} else {
				var repoRelease models.RepoRelease
				repoRelease.RepoId = repoIds[i]
				repoRelease.RepoName = repo.Name
				repoRelease.Url = repo.Url
				for j := 0; j < len(tags); j++ {
					repoRelease.ReleaseTags = append(repoRelease.ReleaseTags, *tags[j].TagName)
				}
				if len(tags) == 0 {
					repoRelease.ReleaseTags = []string{}
				}
				releases = append(releases, &repoRelease)
			}
		}
	}
	Response(w, http.StatusOK, "", releases)
}

func createBranchFromRelease(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	ctx := context.Background()
	var reqBody []models.BFRReq
	if err := utils.GetBody(r.Body, &reqBody); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "Error: Invalid request data", nil)
		return
	}
	count := 0
	for i := 0; i < len(reqBody); i++ {
		repo := getRepositoryInfo(db, reqBody[i].ProjectId, reqBody[i].RepoId)
		if repo.Name != "" {
			ok, err := github.CreateBranch(ctx, repo, reqBody[i].ReleaseTag, reqBody[i].Branch)
			if err != nil {
				log.Fatal("Err creating branch: " + reqBody[i].Branch + " from release: " + reqBody[i].ReleaseTag + " : " + err.Error())
			}
			if ok {
				count += 1
			}
		}
	}
	Response(w, http.StatusOK, fmt.Sprintf("%d branches created from release tag", count), "")
}