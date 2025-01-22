package main

import (
	"log"
	"fmt"
	"flag"
	"context"
	"net/http"
	"database/sql"
	"github.com/shappy0/ghra/handlers"
	"github.com/shappy0/ghra/utils"
)

var db *sql.DB

func setupHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/signup", handlers.SignupHandler)
	mux.HandleFunc("/projects", handlers.ProjectsHandler)
	mux.HandleFunc("/project/{id}", handlers.ProjectHandler)
	mux.HandleFunc("/repository", handlers.RepositoriesHandler)
	mux.HandleFunc("/repository/{id}/deps", handlers.RepositoriesHandler)
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
	flag.IntVar(&port, "port", 8080, "Webserver port")
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
	log.Printf("Webserver started on http://0.0.0.0" + addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error: " + err.Error())
	} 
}