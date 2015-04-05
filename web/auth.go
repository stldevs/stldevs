package web

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
)

func login(ctx Context) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		session := ctx.SessionData(w, r)
		state := randSeq(10)
		session.Values["state"] = state
		session.Save(r, w)
		url := ctx.AuthCodeURL(state, oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, 302)
	}
}

func oauth2Handler(ctx Context) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		code := r.URL.Query().Get("code")
		if code == "" {
			log.Println("code is blank")
			return
		}

		state := r.URL.Query().Get("state")
		session := ctx.SessionData(w, r)
		sessState, ok := session.Values["state"]
		if !ok || state != sessState.(string) {
			log.Println("State mismatch", ok, state, sessState)
			ctx.ParseAndExecute(w, "error", response{"error": "state is incorrect"})
			return
		}

		if user, err := ctx.GithubLogin(code); err != nil {
			log.Println(err)
			ctx.ParseAndExecute(w, "error", response{"error": "error in logging in"})
			return
		} else {
			session.Values["user"] = *user
			http.Redirect(w, r, "/", 302)
		}
	}
}

func logout(ctx Context) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		session := ctx.SessionData(w, r)
		for key, _ := range session.Values {
			delete(session.Values, key)
		}
		http.Redirect(w, r, "/", 302)
	}
}
