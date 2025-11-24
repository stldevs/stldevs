package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/github"
	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/db"
	"github.com/jakecoffman/stldevs/sessions"
	"golang.org/x/oauth2"
	oa2gh "golang.org/x/oauth2/github"
)

func New(cfg *config.Config) []crud.Spec {
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.GithubClientID,
		ClientSecret: cfg.GithubClientSecret,
		RedirectURL:  "http://localhost:8080/callback",
		Endpoint:     oa2gh.Endpoint,
	}

	var stateConfig gologin.CookieConfig
	if cfg.Environment == "prod" {
		oauth2Config.RedirectURL = "https://stldevs.com/stldevs-api/callback"
		stateConfig = gologin.CookieConfig{
			Name:     "stldevs",
			Path:     "/",
			MaxAge:   60,
			HTTPOnly: true,
			Secure:   true, // secure only
		}
	} else {
		stateConfig = gologin.DebugOnlyCookieConfig
	}

	loginTags := []string{"Login"}

	success := &sessions.Issuer{}
	return []crud.Spec{{
		Method:      "GET",
		Path:        "/login",
		Handler:     github.StateHandler(stateConfig, github.LoginHandler(oauth2Config, nil)),
		Description: "GitHub OAuth Login",
		Tags:        loginTags,
	}, {
		Method:      "GET",
		Path:        "/callback",
		Handler:     github.StateHandler(stateConfig, github.CallbackHandler(oauth2Config, success, nil)),
		Description: "GitHub OAuth Callback",
		Tags:        loginTags,
	}, {
		Method:      "GET",
		Path:        "/logout",
		Handler:     logout,
		Description: "Logout of session",
		Tags:        loginTags,
	}, {
		Method:      "GET",
		Path:        "/me",
		PreHandlers: Authenticated,
		Handler:     me,
		Description: "Get info about the logged in user",
		Tags:        loginTags,
	}, {
		Method:      "PATCH",
		Path:        "/me",
		PreHandlers: Authenticated,
		Handler:     updateMe,
		Description: "Get info about the logged in user",
		Tags:        loginTags,
		Validate: crud.Validate{
			Body: crud.Object(map[string]crud.Field{
				"Hide": crud.Boolean().Required(),
			}),
		},
	}}
}

func Authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessions.Cookie)
		if err != nil || cookie.Value == "" {
			http.Error(w, "Not logged in", 401)
			return
		}
		session, ok := sessions.Store.Get(cookie.Value)
		if !ok {
			http.Error(w, "Not logged in", 401)
			return
		}
		ctx := context.WithValue(r.Context(), sessions.KeySession, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func me(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 200, sessions.GetEntry(r).User)
}

type UpdateUser struct {
	Hide bool
}

// Patch allows users to show or hide themselves in the site.
// This is specifically for the /you page because it sends the same response back.
func updateMe(w http.ResponseWriter, r *http.Request) {
	session := sessions.GetEntry(r)

	var cmd UpdateUser
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Failed to bind command object. Are you sending JSON? "+err.Error(), 400)
		return
	}
	err := db.HideUser(cmd.Hide, session.User.Login)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	session.User.Hide = cmd.Hide
	jsonResponse(w, 200, session.User)
}

func logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessions.Cookie)
	if err != nil {
		jsonResponse(w, 200, "already logged out")
		return
	}
	sessions.Store.Evict(cookie.Value)
	http.SetCookie(w, &http.Cookie{
		Name:     sessions.Cookie,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "stldevs.com",
		HttpOnly: true,
		Secure:   true,
	})
	jsonResponse(w, 200, "logged out")
}

func jsonResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
