package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

var (
	config = oauth2.Config{
		ClientID:     "888888",
		ClientSecret: "666666",
		Scopes:       []string{"all"},
		RedirectURL:  "http://localhost:8888/oauth2",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:9999/authorize",
			TokenURL: "http://localhost:9999/token",
		},
	}
)

//Step(A):redirect to authorize url
func homePage(w http.ResponseWriter, r *http.Request) {
	log.Println("Step(A): redirect handle")
	u := config.AuthCodeURL("xyz")
	log.Println(u)
	http.Redirect(w, r, u, http.StatusFound)
}

//Step(C): get access token from auth server
func authorize(w http.ResponseWriter, r *http.Request) {
	log.Println("Step(C): get access token from auth server")
	r.ParseForm()
	state := r.Form.Get("state")
	if state != "xyz" {
		http.Error(w, "State invalid", http.StatusBadRequest)
		return
	}

	code := r.Form.Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	e.Encode(*token)
}

func main() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/oauth2", authorize)
	log.Println("Client Server is running at 8888 port.")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
