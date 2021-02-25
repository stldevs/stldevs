package lang

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/crud"
	"github.com/jakecoffman/stldevs/db"
)

var Routes = []crud.Spec{{
	Method:      "GET",
	Path:        "/langs",
	Handler:     List,
	Description: "List languages",
	Tags:        []string{"Languages"},
	Validate:    crud.Validate{},
}, {
	Method:      "GET",
	Path:        "/langs/{lang}",
	Handler:     Get,
	Description: "Gets a language and displays repo information",
	Tags:        []string{"Languages"},
	Validate: crud.Validate{
		Query: map[string]crud.Field{
			"limit":  crud.Number().Min(1).Max(25).Description("Maximum number of items to return"),
			"offset": crud.Number().Min(0).Description("Number of entries to skip"),
		},
	},
}}

func List(c *gin.Context) {
	c.JSON(200, db.PopularLanguages())
}

func Get(c *gin.Context) {
	var query struct {
		Limit  int `form:"limit"`
		Offset int `form:"offset"`
	}
	if err := c.BindQuery(&query); err != nil {
		return
	}
	if query.Limit <= 0 {
		query.Limit = 25
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	langs := db.Language(c.Params.ByName("lang"))

	if query.Limit+query.Offset > len(langs) {
		query.Limit = len(langs)
	} else {
		query.Limit += query.Offset
	}
	if query.Offset > len(langs) {
		query.Limit = 0
		query.Offset = 0
	}
	c.JSON(200, map[string]interface{}{
		"languages": langs[query.Offset:query.Limit],
		"count":     len(langs),
		"language":  c.Params.ByName("lang"),
	})
}
