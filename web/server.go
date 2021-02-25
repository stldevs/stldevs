package web

import (
	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/config"
	"github.com/jakecoffman/stldevs/web/auth"
	"github.com/jakecoffman/stldevs/web/dev"
	"github.com/jakecoffman/stldevs/web/lang"
	"github.com/jakecoffman/stldevs/web/repo"
	"github.com/jakecoffman/stldevs/web/run"
	"log"
)

func Run(cfg *config.Config) {
	r := crud.NewRouter("StL Devs API", "1.0.0")

	r.Add(auth.New(cfg)...)
	r.Add(repo.Routes...)
	r.Add(run.Routes...)
	r.Add(dev.Routes...)
	r.Add(lang.Routes...)

	log.Println("Serving on http://127.0.0.1:8080")
	if err := r.Serve("0.0.0.0:8080"); err != nil {
		log.Println(err)
	}
}
