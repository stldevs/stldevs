package lang

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/stldevs/db"
)

func List(c *gin.Context) {
	c.JSON(200, db.PopularLanguages())
}

func Get(c *gin.Context) {
	var query struct {
		Limit  int `form:"limit"`
		Offset int `form:"offset"`
	}
	_ = c.BindQuery(&query)
	if query.Limit <= 0 {
		query.Limit = 25
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	langs, userCount := db.Language(c.Params.ByName("lang"))

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
		"count":     userCount,
		"language":  c.Params.ByName("lang"),
	})
}
