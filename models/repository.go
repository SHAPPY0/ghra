package models

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
	DepFileName	string
	Active		bool
	CreatedAt	string
	UpdatedAt	string
}

type CommitReq struct {
	NewContent	map[string]interface{}
	Message		string
	Branch		string
	SHA			string
	RepoId		int
	ProjectId	int
}