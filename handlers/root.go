package handlers

import (
	"log"
	"net/http"
	"encoding/json"
	"database/sql"
	"html/template"
	"github.com/shappy0/ghra/utils"
)

var tmpl *template.Template

type SignupReq struct {
	Username	string
	Email		string
	Password	string
	Role 		int
}

func InitTemplate(mux *http.ServeMux) {
	var err error
	tmm := template.FuncMap{
		"FormatDate": utils.FormateDate,
		"TimeDuration": utils.TimeDuration,
	}
	tmpl, err = template.New("").Funcs(tmm).ParseGlob("templates/*.html")

	// tmpl.Funcs(template.FuncMap{
	// 	"FormateDate": utils.FormateDate,
	// })
	if  err != nil {
		log.Fatal( err)
	}
	//Allow static files
	fs := http.FileServer(http.Dir("./statics"))
	mux.Handle("/statics/*", http.StripPrefix("/statics/", fs))
}

func Response(w http.ResponseWriter, code int, msg string, data any) {
	result := map[string]interface{}{
		"status": code,
		"message": msg,
		"data": data,
	}
	w.Header().Set("Content-Type", "applicationn/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(result)
}

func ErrorResponse(w http.ResponseWriter, code int, msg string, data any) {
	result := map[string]interface{}{
		"status": code,
		"error": msg,
		"data": data,
	}
	w.Header().Set("Content-Type", "applicationn/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(result)
}

func RenderTemplate(w http.ResponseWriter, pageName string, data interface{}) {
	if err := tmpl.ExecuteTemplate(w, pageName, data); err != nil {
		log.Println(err.Error())
	}
}

func RenderErrorTemplate(w http.ResponseWriter, code int, msg string) {
	data := map[string]interface{}{
		"error": msg,
		"code": code,
		"status": "Error",
	}
	if err := tmpl.ExecuteTemplate(w, "error", data); err != nil {
		log.Println(err.Error())
	}
}


func RootHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Content   string
	}{
		Content:   "index",
	}
	RenderTemplate(w, "index", data)
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Content   string
	}{
		Content:   "about",
	}
	RenderTemplate(w, "about", data)
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var reqBody SignupReq
		var db *sql.DB
		if err := utils.GetBody(r.Body, &reqBody); err != nil {
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}
		if db = (r.Context().Value("db")).(*sql.DB); db == nil {
			http.Error(w, "Invalid DB connection", http.StatusInternalServerError)
			return
		}
		result, err := db.Exec("INSERT INTO users_tbl(username, email, password, role, active) VALUES(?, ?, ?, ?, ?)", 
			reqBody.Username,
			reqBody.Email,
			reqBody.Password,
			reqBody.Role,
			true,
		)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		_, err = result.RowsAffected()
		if err != nil {
			log.Fatal(err)
		}
		Response(w, http.StatusOK, "User created successfully", nil)
	} else {
		http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}
}