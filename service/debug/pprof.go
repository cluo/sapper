package debug

import (
	"net/http"
	"net/http/pprof"
)

type Debug struct {
}

func (d *Debug) GET(w http.ResponseWriter, r *http.Request) {
	pprof.Index(w, r)
}
