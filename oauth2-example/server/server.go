package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-oauth2/oauth2/manage"
	"github.com/go-oauth2/oauth2/store"
	"github.com/gorilla/sessions"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
)

var sessionStore = sessions.NewCookieStore([]byte("1989"))

func main() {
	manager := manage.NewDefaultManager()
	//memory token store
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	//manage client store
	clientStore := store.NewClientStore()
	clientStore.Set("888888", &models.Client{
		ID:     "888888",
		Secret: "666666",
		Domain: "http://localhost:8888",
	})
	manager.MapClientStorage(clientStore)

	srv := server.NewServer(server.NewConfig(), manager)

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/auth", authHandler)

	//MUST set the handler
	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (string, error) {
		s, _ := sessionStore.Get(r, "LoginUser")
		if val, ok := s.Values["Id"]; ok {
			return val.(string), nil
		}

		w.Header().Set("Location", "/auth")
		w.WriteHeader(http.StatusFound)
		return "", errors.ErrAccessDenied
	})

	//HandleAuthorizeRequest
	//Step(B): generate authorize code, and redirect to client's callback url
	http.HandleFunc("/authorize", func(resp http.ResponseWriter, req *http.Request) {
		log.Println("Step(B): authorize handle")
		s, _ := sessionStore.Get(req, "LoginUser")
		if _, ok := s.Values["LoginId"]; !ok {
			if req.Form == nil {
				req.ParseForm()
			}

			log.Println("QueryString :", req.Form)
			encodedForm, _ := json.Marshal(req.Form)
			s.Values["ReturnUri"] = encodedForm
			err := s.Save(req, resp)
			if err != nil {
				log.Printf("sessionStore save error: %v", err)
			}
			resp.Header().Set("Location", "/login")
			resp.WriteHeader(http.StatusFound)
			return
		}

		if form, ok := s.Values["ReturnUri"]; ok {
			json.Unmarshal(form.([]byte), &req.Form)
			log.Println("Get Form String:", req.Form)
		}

		err := srv.HandleAuthorizeRequest(resp, req)
		if err != nil {
			log.Println("authHandler error")
			http.Error(resp, err.Error(), http.StatusBadRequest)
		}
	})

	//HandleTokenRequest
	//Step(D): auth server generate token by auth code, and redirect to client's callback url
	http.HandleFunc("/token", func(resp http.ResponseWriter, req *http.Request) {
		log.Println("Step(D): auth server generate access token")
		err := srv.HandleTokenRequest(resp, req)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Println("Server is running at 9999 port.")
	log.Fatal(http.ListenAndServe(":9999", nil))
}

//display user login form, record userid in session store
func loginHandler(resp http.ResponseWriter, req *http.Request) {
	s, _ := sessionStore.Get(req, "LoginUser")

	if req.Method == "POST" {
		s.Values["LoginId"] = "1234"
		s.Save(req, resp)

		resp.Header().Set("Location", "/auth")
		resp.WriteHeader(http.StatusFound)
		return
	}
	outputHTML(resp, req, "static/login.html")
}

//authHandler
func authHandler(resp http.ResponseWriter, req *http.Request) {
	s, _ := sessionStore.Get(req, "LoginUser")

	if _, ok := s.Values["LoginId"]; !ok {
		resp.Header().Set("Location", "/login")
		resp.WriteHeader(http.StatusFound)
		return
	}

	if req.Method == "POST" {
		log.Println("LoginID: ", s.Values["LoginId"].(string))
		s.Values["Id"] = s.Values["LoginId"].(string)
		s.Save(req, resp)

		resp.Header().Set("Location", "/authorize")
		resp.WriteHeader(http.StatusFound)

		return
	}
	outputHTML(resp, req, "static/auth.html")
}

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}
