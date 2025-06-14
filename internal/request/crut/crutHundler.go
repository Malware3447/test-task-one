package crut

import (
	"context"
	"net/http"
)

type CrutHudnler interface {
	CreateGood(w http.ResponseWriter, r *http.Request)
	GoodUpdate(w http.ResponseWriter, r *http.Request)
	GoodRemove(w http.ResponseWriter, r *http.Request)
	GoodList(w http.ResponseWriter, r *http.Request)
	ReprioritizeGood(w http.ResponseWriter, r *http.Request)
	ProcessNATSMessages(ctx context.Context)
}
