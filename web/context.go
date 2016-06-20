package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Context struct {
	Response http.ResponseWriter
	Request  *http.Request
	Params   httprouter.Params
}

func (c *Context) Render(data interface{}) error {
	c.Response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(c.Response).Encode(data); err != nil {
		log.Println("Error while rendering:", err)
		return err
	}
	return nil
}
