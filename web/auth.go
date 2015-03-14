package web

import (
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
)

func oauth2Handler(ctx Context) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		code := r.URL.Query().Get("code")
		if code == "" {
			log.Println("code is blank")
			return
		}

		state := r.URL.Query().Get("state")
		sessState, _ := get_session(r, "state")
		if sessState == nil || state != sessState.(string) {
			ctx.ParseAndExecute(w, "error", map[string]interface{}{"error": "state is incorrect"})
			return
		}

		token, err := conf.Exchange(oauth2.NoContext, code)
		if err != nil {
			panic(err)
		}

		client := github.NewClient(conf.Client(oauth2.NoContext, token))

		user, _, err := client.Users.Get("")
		if err != nil {
			panic(err)
		}

		if err = set_session(w, r, "user", *user); err != nil {
			panic(err)
		}

		http.Redirect(w, r, "/", 302)
	}
}

func logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Println(err)
		return
	}
	delete(session.Values, "user")
	if err := session.Save(r, w); err != nil {
		log.Println(err)
	}
	http.Redirect(w, r, "/", 302)
}
