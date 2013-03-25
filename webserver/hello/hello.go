package hello

import (
	"fmt"
	"net/http"
	"html/template"
	"appengine"
	"appengine/log"
)

var logTemplate = template.Must(template.New("log.html").ParseFiles("templates/log.html"))

func init() {
	http.HandleFunc("/",root)
	http.HandleFunc("/log",logprinter)
}

func root(w http.ResponseWriter, r *http.Request) {
	if err := logTemplate.Execute(w,nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func logprinter(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	query := &log.Query{
    AppLogs:  true,
    Versions: []string{"1"},
	}
	
	for results := query.Run(c); ; {
		record, err := results.Next()
		if err == log.Done {
			fmt.Fprintf(w,"Done processing results")
			break
		}
		if err != nil {
			c.Errorf("Failed to retrieve next log: %v", err)
			break
		}
		fmt.Fprintf(w,"Saw record %v", record)
	}
}