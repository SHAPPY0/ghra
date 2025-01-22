package github

import (
	"fmt"
	"context"
	"github.com/shappy0/ghra/models"
    "github.com/google/go-github/v68/github"
    "golang.org/x/oauth2"
)

type GhraFile struct {
    RepoContent *github.RepositoryContent
    Content *string
    SHA     string
}

func ReadDepFile(ctx context.Context, repo models.Repository) (*github.RepositoryContent, error) {
	ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: repo.Token},
    )
    client := github.NewClient(oauth2.NewClient(ctx, ts))
	repoName := GetRepoName(repo.Url)
	fileContent, dirContent, _, err := client.Repositories.GetContents(
		ctx, 
		repo.User, 
		repoName, 
		repo.DepFileName, 
		&github.RepositoryContentGetOptions{Ref: repo.Branch})
    if err != nil {
        return nil, err
    }
	if fileContent != nil {
        return fileContent, nil 
    } else if dirContent != nil {
		return nil, fmt.Errorf("Dep filepath is a directory")
    } else {
		return nil,  fmt.Errorf("No content found at the dep file path")
    }
}

func PushChanges(ctx context.Context, repo models.Repository, data models.CommitReq, updatedDeps string) (bool, error) {
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: repo.Token},
    )
    client := github.NewClient(oauth2.NewClient(ctx, ts))
    
    repoName := GetRepoName(repo.Url)

    options := &github.RepositoryContentFileOptions{
		Message: github.String(data.Message),
		Content: []byte(updatedDeps),
		SHA:     &data.SHA,
		Branch:  github.String(data.Branch),
	}
    _, _, err := client.Repositories.UpdateFile(ctx, repo.User, repoName, repo.DepFileName, options)
	if err != nil {
		return false, err
	}
    return true, nil
}