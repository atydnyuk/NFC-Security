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
	Path string
	RawQuery string
	Time time.Time
	Password string
}

var logTemplate = template.Must(template.New("log.html").ParseFiles("templates/log.html"))

func init() {
	http.HandleFunc("/",root)
	http.HandleFunc("/log",logprinter)
}

func root(w http.ResponseWriter, r *http.Request) {
	recordRequest(w,r)
	if err := logTemplate.Execute(w,nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func logprinter(w http.ResponseWriter, r *http.Request) {
	recordRequest(w,r)
	printLogToHTML(w,r)
}

func clipQuotes(x string) string {
	return strings.Join(strings.Split(x,"")[1:len(x)-1],"")
}

func recordRequest(w http.ResponseWriter, r *http.Request) {
	stringpath := fmt.Sprintf("%#v",r.URL.Path)
	rawquery := fmt.Sprintf("%#v",r.URL.RawQuery)
	passwordstring := fmt.Sprintf("%#v",r.FormValue("password"))
	if (stringpath=="\"/submit\"") {
		c := appengine.NewContext(r)
		
		req := RequestRecord{
		RemoteAddr:r.RemoteAddr,
		Host:r.Host,
		Path:stringpath,
		RawQuery:rawquery,
		Time:time.Now(),
		Password:passwordstring,
		}
		
		_, err := datastore.Put(c, datastore.NewIncompleteKey(c,"Record", nil), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func printLogToHTML(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Record").Order("-Time").Limit(100)
    records := make([]RequestRecord, 0, 10)
    
	if _, err := q.GetAll(c, &records); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	
	fmt.Fprintf(w,"<html><head>")
	fmt.Fprintf(w,"<link rel=\"stylesheet\" ")
	fmt.Fprintf(w,"href=\"javascript/log.css\"/>")
	fmt.Fprintf(w,"</head>")
	fmt.Fprintf(w,"<h3>Request Log</h3>")
	counter := 1
	for key := range records {
		cPath:=clipQuotes(records[key].Path)
		cQuery:=clipQuotes(records[key].RawQuery)
		if (cPath == "/submit") {
			fmt.Fprintf(w,"<div class=\"match\">")
			fmt.Fprintf(w,"<div><span class=\"label\">")
			fmt.Fprintf(w,"%d. Remote Address: %s",
				counter,records[key].RemoteAddr)
			fmt.Fprintf(w,"</span></div>")
			fmt.Fprintf(w,"<div class=\"contents\">")
			
			if (len(cQuery)!=0) {
				fmt.Fprintf(w,"Request sent for : %s\n",
					records[key].Host+cPath+"?"+cQuery)
			} else {
				fmt.Fprintf(w,"Request sent for: %s\n",
					records[key].Host+cPath+cQuery)
			}
			fmt.Fprintf(w,"Request sent at : %#v\n",
				records[key].Time.String())
			if (len(records[key].Password)>0) {
				fmt.Fprintf(w,"Password is : %s\n",
					records[key].Password)
			}
			fmt.Fprintf(w,"</div></div>")
			counter++
		}
	}
}