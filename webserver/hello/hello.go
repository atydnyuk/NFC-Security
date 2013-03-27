package hello

import (
	"fmt"
	"net/http"
	"html/template"
	"appengine"
	"appengine/datastore"
	"time"
	"strings"
)

type RequestRecord struct {
	RemoteAddr string
	Host string
	Method string
	Header string
	Path string
	RawQuery string
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
	stringpath := fmt.Sprintf("%#v",r.URL.Path)
	rawquery := fmt.Sprintf("%#v",r.URL.RawQuery)
	c := appengine.NewContext(r)
    req := RequestRecord{
    RemoteAddr:r.RemoteAddr,
	Host:r.Host,
	Method:r.Method,
	Header:stringHeader,
	Path:stringpath,
	RawQuery:rawquery,
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
	fmt.Fprintf(w,"<html><head>")
	fmt.Fprintf(w,"<link rel=\"stylesheet\"")
	fmt.Fprintf(w,"href='templates/log.css'/>")
	fmt.Fprintf(w,"</head>")
	for key := range records {
		fmt.Fprintf(w,"<div class=\"match\">")
		fmt.Fprintf(w,"<div><span class=\"label\">")
		fmt.Fprintf(w,"%s",records[key].RemoteAddr)
		fmt.Fprintf(w,"</span></div>")
		fmt.Fprintf(w,"<div class=\"contents\">")
		cPath:=clipQuotes(records[key].Path)
		cQuery:=clipQuotes(records[key].RawQuery)
		if (len(cQuery)!=0) {
			fmt.Fprintf(w,"%s\n",records[key].Host+cPath+"?"+cQuery)
		} else {
			fmt.Fprintf(w,"%s\n",records[key].Host+cPath+cQuery)
		}
		fmt.Fprintf(w,"</div></div>")
	}
}

func clipQuotes(x string) string {
	return strings.Join(strings.Split(x,"")[1:len(x)-1],"")
}