package main

import "net/http"

//type indexHandler struct{}
// func (indexHandler) ServeHTTP(http.FileServer(http.Dir("index.html")), *http.Request) {}

func main() {
	smu := http.NewServeMux()
	smu.Handle("/", http.FileServer(http.Dir(".")))

	server := http.Server{
		Addr:    ":8080",
		Handler: smu,
	}
	server.ListenAndServe()
}
