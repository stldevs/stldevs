package web

import (
	"log"
	"net/http"
	"text/template"

	"github.com/google/go-github/github"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
	oa2gh "golang.org/x/oauth2/github"
)

var conf *oauth2.Config

func init() {
	conf = &oauth2.Config{
		ClientID:     "cfa23414a111bbac97c8",
		ClientSecret: "10cb393e043fb569b8428779ebf70285a331915d", // TODO: hide after testing
		Scopes:       []string{"user:email", "public_repo"},
		Endpoint:     oa2gh.Endpoint,
	}
}

func oauth2Handler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//	client_secret := p.ByName("client_secret") // TODO: Verify matches
	//	state := p.ByName("state") // TODO: verify this was the state sent earlier
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Println("code is blank")
		return
	}

	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Println(err)
		return
	}

	client := github.NewClient(conf.Client(oauth2.NoContext, token))
	emails, _, err := client.Users.ListEmails(nil)
	if err != nil {
		log.Println(err)
	}
	var primary string
	for _, email := range emails {
		if email.Primary != nil && *email.Primary == true {
			primary = *email.Email
		}
	}

	template, err := template.New("index.html").ParseFiles("templates/index.html")
	if err != nil {
		log.Println(err)
		return
	}

	data := map[string]string{
		"github": conf.AuthCodeURL("state", oauth2.AccessTypeOffline),
		"user":   primary,
	}
	if err = template.Execute(w, data); err != nil {
		log.Println(err)
		return
	}
}
