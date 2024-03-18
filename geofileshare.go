package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, I see you are vistiting the page on %v\n", r.URL.Path)
    })

    http.ListenAndServe(":85", nil)
}
