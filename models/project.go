package models

type Project struct {
	Id			int
	Name		string
	Description	string
	Active		bool
	CreatedAt 	string
	UpdatedAt 	string	
}


type ProjectReq struct {
	Name 	string
	Description string
}