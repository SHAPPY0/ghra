package models

import (
	"github.com/vifraa/gopom"
)

type Repository struct {
	Id			int
	ProjectId	int
	Name 		string
	Url			string
	Branch		string
	User		string
	Token		string
	Tags		string
	BuildTool	string
	DepFilePath	string
	Active		bool
	CreatedAt	string
	UpdatedAt	string
}

type RepoDeps struct {
	Name		string
	ProjectId	int
	RepoId 		int
	Branch 		string
	Content		string
	DepFilePath string
	SHA			string
	Properties	*gopom.Properties
	Dependencies *[]gopom.Dependency
	LinedContent map[int]string
	DepHashed 	map[string]string
}

type RepoReq struct {
	ProjectId	int
	Name 		string
	Url			string
	Branch		string
	User		string
	Token		string
	Tags		string
	BuildTool	string
	DepFilePath	string
}

type CommitReq struct {
	NewContent	map[string]interface{}
	Message		string
	Branch		string
	SHA			string
	RepoId		int
	ProjectId	int
}

type VCDepsReq struct {
	RepoIds []int
	ProjectId int
}