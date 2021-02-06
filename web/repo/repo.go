package repo

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/stldevs/db"
)

func List(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(400, "q is a required query parameter")
		return
	}
	c.JSON(200, db.SearchRepos(q))
}
