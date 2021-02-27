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

	must(r.Add(auth.New(cfg)...))
	must(r.Add(repo.Routes...))
	must(r.Add(run.Routes...))
	must(r.Add(dev.Routes...))
	must(r.Add(lang.Routes...))

	log.Println("Serving on http://127.0.0.1:8080")
	if err := r.Serve("0.0.0.0:8080"); err != nil {
		log.Println(err)
	}
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
