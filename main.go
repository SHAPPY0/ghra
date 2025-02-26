package main

import (
	"log"
	"fmt"
	"flag"
	"context"
	"net/http"
	"os/exec"
	"runtime"
	"database/sql"
	"github.com/shappy0/ghra/handlers"
	"github.com/shappy0/ghra/utils"
)

var db *sql.DB

func setupHandlers(mux *http.ServeMux) {
	// mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/", handlers.ProjectsHandler)
	mux.HandleFunc("/signup", handlers.SignupHandler)
	mux.HandleFunc("/projects", handlers.ProjectsHandler)
	mux.HandleFunc("/project/{id}", handlers.ProjectHandler)
	mux.HandleFunc("/project/{id}/json", handlers.ProjectHandler)
	// mux.HandleFunc("/project/{id}/repositories", handlers.RepositoriesHandler)
	mux.HandleFunc("/repository", handlers.RepositoriesHandler)
	mux.HandleFunc("/repository/{projectId}/{repoId}", handlers.RepositoriesHandler)
	mux.HandleFunc("/repository/{id}/deps", handlers.RepositoriesHandler)
	mux.HandleFunc("/repositories/cascade/{type}/deps", handlers.RepositoriesHandler)
	mux.HandleFunc("/deps", handlers.DependenciesHandler)
	mux.HandleFunc("/vc/deps", handlers.DependenciesHandler)

}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func dbMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "db", db)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8184, "Webserver port")
	flag.Parse()

	mux := http.NewServeMux()
	
	handlers.InitTemplate(mux)

	//Connect to db
	var err error
	db, err = utils.NewDbConnection()
	if err != nil {
		log.Fatal(err)
	}
	setupHandlers(mux)
	
	// mux = enableCORS(mux)
	// mux = dbMiddleware(mux)

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr: addr,
		Handler: enableCORS(dbMiddleware(mux)),
	}
	url := "http://localhost" + addr
	log.Printf("Webserver starting on " + url + " and opening it in browser.")
	go openWeb(url)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error: " + err.Error())
	} 
}

func openWeb(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Run()
}