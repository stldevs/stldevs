package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/dghubble/gologin/v2/github"
	"github.com/jakecoffman/stldevs/db"
	"log"
	"net/http"
	"time"
)

type Issuer struct{}

func (s *Issuer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	githubUser, err := github.UserFromContext(r.Context())
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		_, _ = w.Write([]byte("\"" + err.Error() + "\""))
		return
	}

	log.Println("Login success", *githubUser.Login)

	user, err := db.GetUser(*githubUser.Login)
	if err != nil || user == nil {
		// user not found or something?
		user = &db.StlDevsUser{
			User: githubUser,
		}
	}

	expire := time.Now().AddDate(0, 0, 1)
	cookie := http.Cookie{
		Name:    Cookie,
		Value:   Store.Add(user),
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/you", http.StatusFound)
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateSessionCookie() string {
	b, _ := GenerateRandomBytes(64)
	return base64.URLEncoding.EncodeToString(b)
}
