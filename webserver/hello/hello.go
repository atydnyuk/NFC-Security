package hello

import (
	"fmt"
	"net/http"
	"html/template"
)

var logTemplate = template.Must(template.New("log.html").ParseFiles("templates/log.html"))

func init() {
	http.HandleFunc("/",root)
}

func root(w http.ResponseWriter, r *http.Request) {
	if err := indexTemplate.Execute(w,nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"<html><head><script type=\"text/javascript\"src=\"/javascript/jquery.js\"></script></head>Hello World!<script>alert(1);</script></html>")
	
}