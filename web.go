package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"os"
)

// Server is an abstraction that runs via a web interface
type Server struct {
	DB     *Database
	router http.Handler
	logger http.Handler
}

// NewServer instantiates a server with an active database
func NewServer(db *Database) *Server {
	s := Server{DB: db}
	s.router = s.BuildRouter()

	http.Handle("/", s.router)
	s.logger = handlers.LoggingHandler(os.Stdout, s.router)

	return &s
}

// StartHandler is the entry point to the web GUI
//
// If not authenticated, it allows for registration.
// If authenticated and a tournament is running, show that tournament.
// If authenticated and no tournament is running, show a list of tournaments.
func (s *Server) StartHandler(w http.ResponseWriter, r *http.Request) {
	t := getTemplates("static/tournaments.html")
	data := struct {
		Tournaments []*Tournament
	}{
		s.DB.Tournaments,
	}

	render(t, w, r, data)
}

// NewHandler shows the page to create a new tournament
func (s *Server) NewHandler(w http.ResponseWriter, r *http.Request) {
	// If there is a post to this URL, it means we are making a new tournament
	if r.Method == "POST" {
		name := r.PostFormValue("name")
		id := r.PostFormValue("id")
		t, _ := NewTournament(name, id, s.DB)
		log.Printf("Created tournament %s!", t.Name)
		http.Redirect(w, r, "/", 302)
		return
	}

	// Elsewise, show the GUI
	t := getTemplates("static/new.html")
	render(t, w, r, struct{}{})
}

// TournamentHandler shows the tournament view and handles tournaments
func (s *Server) TournamentHandler(w http.ResponseWriter, r *http.Request) {
	t := getTemplates("static/tournament.html", "static/match.html")
	vars := mux.Vars(r)
	data := struct {
		Tournament *Tournament
	}{
		s.DB.tournamentRef[vars["id"]],
	}

	render(t, w, r, data)
}

// MatchHandler shows the Match view and handles Matches
func (s *Server) MatchHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "The battle raged ever on.\n")
}

// ActionHandler handles judge requests for player action
func (s *Server) ActionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Arrows were fired, archers were killed, and shots were had.\n")
}

// BuildRouter sets up the routes
func (s *Server) BuildRouter() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/", s.StartHandler)
	r.HandleFunc("/new", s.NewHandler)
	r.HandleFunc("/{id}/", s.TournamentHandler)
	r.HandleFunc("/{id}/{kind:(tryout|runnerup|semi|final)}/{index:[0-9]+}/", s.MatchHandler)
	r.HandleFunc("/{id}/{kind:(tryout|runnerup|semi|final)}/{index:[0-9]+}/{player:[0-3]}", s.ActionHandler)

	return r
}

// Serve serves forever
func (s *Server) Serve() error {
	log.Print("Listening on :3420")
	return http.ListenAndServe(":3420", s.logger)
}

// getTemplates gets a template with the context set to `extra`, with index.html backing it.
func getTemplates(extra string) *template.Template {
	t, err := template.ParseFiles(extra, "static/index.html")
	if err != nil {
		log.Fatal(err)
	}

	return t
}

func render(t *template.Template, w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := t.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	db, err := NewDatabase("production.db")
	if err != nil {
		log.Fatal(err)
	}

	err = NewServer(db).Serve()
	if err != nil {
		log.Fatal(err)
	}
}
