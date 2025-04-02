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
    Branch  string
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
		repo.DepFilePath, 
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

func PushChanges(ctx context.Context, repo models.Repository, fileObj *GhraFile, msg string, updatedDeps string) (bool, error) {
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: repo.Token},
    )
    client := github.NewClient(oauth2.NewClient(ctx, ts))
    
    repoName := GetRepoName(repo.Url)
    depFilePath := GetDepFilePath(repo.DepFilePath)
    options := &github.RepositoryContentFileOptions{
		Message: github.String(msg),
		Content: []byte(updatedDeps),
		SHA:     &fileObj.SHA,
		Branch:  github.String(fileObj.Branch),
	}

    _, _, err := client.Repositories.UpdateFile(ctx, repo.User, repoName, depFilePath, options)
	if err != nil {
		return false, err
	}
    return true, nil
}

func GetReleaseTags(ctx context.Context, repo models.Repository) ([]*github.RepositoryRelease, error) {
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: repo.Token},
    )
    client := github.NewClient(oauth2.NewClient(ctx, ts))
	repoName := GetRepoName(repo.Url)
    releases, _, err := client.Repositories.ListReleases(ctx, repo.User, repoName, nil)
    if err != nil {
        return nil, err
    }
    return releases, nil
}

func GetBranches(ctx context.Context, repo models.Repository) ([]string, error) {
    var branchList []string
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: repo.Token},
    )
    client := github.NewClient(oauth2.NewClient(ctx, ts))
	repoName := GetRepoName(repo.Url)
    opt := &github.BranchListOptions{
        ListOptions: github.ListOptions{PerPage: 100},
    }
    for {
        branches, resp, err := client.Repositories.ListBranches(ctx, repo.User, repoName, opt)
        if err != nil {
            fmt.Errorf("Error fetching branches: %v", err)
        }
        for _, branch := range branches {
            branchList = append(branchList, *branch.Name)
        }
        if resp.NextPage == 0 {
            break
        }
        opt.Page = resp.NextPage
    }
    return branchList, nil
}

func getSHA(ctx context.Context, client *github.Client, owner, repo, tagName string) (string, error) {
    tagRef, _, err := client.Git.GetRef(ctx, owner, repo, "tags/"+tagName)
	if err != nil {
		return "", err
	}
	sha := tagRef.Object.GetSHA()
    return sha, nil
}

func CreateBranch(ctx context.Context, repo models.Repository, tag, branch string) (bool, error) {
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: repo.Token},
    )
    client := github.NewClient(oauth2.NewClient(ctx, ts))
    repoName := GetRepoName(repo.Url)
    owner := repo.User
    sha, err := getSHA(ctx, client, owner, repoName, tag)
    if err != nil {
        return false, err
    }
    newRef := &github.Reference{
        Ref:	github.String("refs/heads/" + branch),
        Object: &github.GitObject{SHA: github.String(sha)},
    }
    ref, resp, err := client.Git.CreateRef(ctx, owner, repoName, newRef)
    if err != nil {
        return false, err
    }
    _ = ref
    _ = resp
    return true, nil
}