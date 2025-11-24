package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/dghubble/gologin/v2/github"
	"github.com/jakecoffman/stldevs/db"
	"github.com/jakecoffman/stldevs/db/sqlc"
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
	if err != nil || user.Login == "" {
		// user not found or something?
		user = sqlc.GetUserRow{
			Login: *githubUser.Login,
		}
		if githubUser.AvatarURL != nil {
			user.AvatarUrl = *githubUser.AvatarURL
		}
		if githubUser.Name != nil {
			user.Name = *githubUser.Name
		}
		if githubUser.Email != nil {
			user.Email = *githubUser.Email
		}
	}

	expire := time.Now().AddDate(0, 0, 1)
	cookie := http.Cookie{
		Name:    Cookie,
		Value:   Store.Add(&user),
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
