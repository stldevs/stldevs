package web

import (
	"log"
	"net/http"
)

func get_session(r *http.Request, key string) (interface{}, bool) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Println(err)
		return nil, false
	}
	value := session.Values[key]
	if value == nil {
		return nil, false
	}
	return value, true
}

func set_session(w http.ResponseWriter, r *http.Request, key, value interface{}) error {
	session, err := store.Get(r, "session")
	if err != nil {
		return err
	}
	session.Values[key] = value
	return session.Save(r, w)
}
