package web

import (
	"log"
	"net/http"
)

func get_session(r *http.Request, key string) (string, bool) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Println(err)
		return "", false
	}
	value := session.Values[key]
	if value == nil {
		return "", false
	}
	return value.(string), true
}

func set_session(w http.ResponseWriter, r *http.Request, key, value string) error {
	session, err := store.Get(r, "session")
	if err != nil {
		return err
	}
	session.Values[key] = value
	return session.Save(r, w)
}
