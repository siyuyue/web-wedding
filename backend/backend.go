package backend

import (
    "fmt"
    "net/http"
)

func init() {
    http.HandleFunc("/rsvp", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Not implemented yet!")
}
