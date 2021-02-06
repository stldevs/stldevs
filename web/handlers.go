package web

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/stldevs/db"
)

func search(c *gin.Context) {
	q := c.Request.URL.Query().Get("q")
	kind := c.Request.URL.Query().Get("type")

	if q == "" {
		c.Status(400)
		return
	}

	c.JSON(200, map[string]interface{}{
		"results": db.Search(q, kind),
	})
}
