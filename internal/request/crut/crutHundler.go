package crut

import "net/http"

type CrutHudnler interface {
	CreateGood(w http.ResponseWriter, r *http.Request)
}
