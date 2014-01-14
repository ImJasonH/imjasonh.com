package imjasonh

import (
	"net/http"
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://imjasonh.com", http.StatusMovedPermanently)
	})
}
