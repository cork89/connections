package main

import "net/http"

func headsupHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl["headsup"].ExecuteTemplate(w, "base.html", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
