package main

import "net/http"

func setMethod(method string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method is not supported", http.StatusMethodNotAllowed)
			return
		}
		next(w, r)
	}

}
