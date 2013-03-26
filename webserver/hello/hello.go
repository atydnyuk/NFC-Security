package hello

import (
	"fmt"
	"net/http"
	"html/template"
	"appengine"
	"appengine/datastore"
	"time"
)

type RequestRecord struct {
	RemoteAddr string
	Host string
	Method string
	Header string
	URL string
	Time time.Time
}

var logTemplate = template.Must(template.New("log.html").ParseFiles("templates/log.html"))

func init() {
	http.HandleFunc("/",logprinter)
	http.HandleFunc("/log",root)
}

func root(w http.ResponseWriter, r *http.Request) {
	if err := logTemplate.Execute(w,nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func logprinter(w http.ResponseWriter, r *http.Request) {
	stringURL := fmt.Sprintf("%#v",r.URL)
	stringHeader := fmt.Sprintf("%#v",r.Header)
	c := appengine.NewContext(r)
    req := RequestRecord{
    RemoteAddr:r.RemoteAddr,
	Host:r.Host,
	Method:r.Method,
	Header:stringHeader,
	URL:stringURL,
	Time:time.Now(),
	}
	_, err := datastore.Put(c, datastore.NewIncompleteKey(c,"Record", nil), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
	
    q := datastore.NewQuery("Record").Order("-Time").Limit(10)
    records := make([]RequestRecord, 0, 10)
    if _, err := q.GetAll(c, &records); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	for key := range records {
		fmt.Fprintf(w,"%#v\n",records[key])
	}
}