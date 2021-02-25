package repo

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/db"
)

var Routes = []crud.Spec{{
	Method:      "GET",
	Path:        "/repos",
	Handler:     List,
	Description: "Lists repositories",
	Tags:        []string{"Repos"},
	Validate: crud.Validate{
		Query: map[string]crud.Field{
			"q": crud.String().Required().Description("Query string"),
		},
	},
}}

func List(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(400, "q is a required query parameter")
		return
	}
	c.JSON(200, db.SearchRepos(q))
}
