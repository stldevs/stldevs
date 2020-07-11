package web

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type DevController struct {
	db    *sqlx.DB
	store *SessionStore
}

func (d *DevController) List(c *gin.Context) {
	if listing := PopularDevs(d.db); listing == nil {
		c.JSON(500, "Failed to list")
		return
	} else {
		c.JSON(200, listing)
	}
}

func (d *DevController) Get(c *gin.Context) {
	profile, err := Profile(d.db, c.Params.ByName("login"))
	if err != nil {
		c.JSON(500, "Failed to find user")
		return
	}
	c.JSON(200, profile)
}

type UpdateUser struct {
	Hide bool
}

// Patch allows users and admins show or hide themselves in the site
func (d *DevController) Patch(c *gin.Context) {
	profile, err := Profile(d.db, c.Params.ByName("login"))
	if err != nil {
		c.JSON(500, "Failed to find user")
		return
	}
	var cmd UpdateUser
	if err = c.BindJSON(&cmd); err != nil {
		c.JSON(400, "Failed to bind command object. Are you sending JSON?")
		return
	}
	session := c.MustGet("session").(*SessionEntry)
	if session.User.IsAdmin == false && *session.User.Login != *profile.User.Login {
		c.JSON(403, "Users can only modify themselves")
		return
	}
	result, err := d.db.Exec("update agg_user set hide=$1 where login=$2", cmd.Hide, *profile.User.Login)
	if err != nil {
		log.Println(err)
		c.JSON(500, err.Error())
		return
	}
	if affected, _ := result.RowsAffected(); affected == 1 {
		profile.User.Hide = cmd.Hide
		session.User.Hide = cmd.Hide
	}
	c.JSON(200, profile)
}

// Delete allows admins to easily expunge old data
func (d *DevController) Delete(c *gin.Context) {
	session := c.MustGet("session").(*SessionEntry)
	if session.User.IsAdmin == false {
		c.JSON(403, "Only admins can delete users")
		return
	}

	login := c.Params.ByName("login")
	_, err := d.db.Exec("delete from agg_repo where owner=$1", login)
	if err != nil {
		log.Println(err)
		c.JSON(500, err.Error())
		return
	}

	_, err = d.db.Exec("delete from agg_user where login=$1", login)
	if err != nil {
		log.Println(err)
		c.JSON(500, err.Error())
		return
	}

	c.JSON(200, "deleted")
}
