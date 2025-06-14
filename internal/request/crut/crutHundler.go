package crut

import (
	"context"
	"net/http"
)

type CrutHudnler interface {
	CreateGood(w http.ResponseWriter, r *http.Request)
	ProcessNATSMessages(ctx context.Context)
}
