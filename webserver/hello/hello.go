package hello

import (
	"fmt"
	"net/http"
	"html/template"
	"appengine"
	"appengine/datastore"
	"time"
	"strings"
	"crypto/rand"
	"io"
	"encoding/base64" 
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
var tagstring string

type TagPass struct {
    Password string
}

func init() {
	http.HandleFunc("/",root)
	http.HandleFunc("/log",logprinter)
	http.HandleFunc("/submit",submitcode)

}

func root(w http.ResponseWriter, r *http.Request) {
	recordRequest(w,r)
	if err := logTemplate.Execute(w,nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func logprinter(w http.ResponseWriter, r *http.Request) {
	printLogToHTML(w,r)
}

func submitcode(w http.ResponseWriter, r *http.Request) {
	//put the request recieved in the log
	recordRequest(w,r)

	//get the password we expect from the datastore
	checkpass(r)

	//check if the submitted password is correct
	passwordstring := fmt.Sprintf("%#v",r.FormValue("password"))
	fmt.Fprintf(w,"You submitted the password: %s\n",passwordstring)
	responsestring,accepted := generateResponse(passwordstring)
	if (accepted) {
		fmt.Fprintf(w,"ACCEPTED. Please write this to the tag: %s\n",
			responsestring)
		writeNewPassword(r)
	} else {
		fmt.Fprintf(w,"REJECTED we want %s.\n",tagstring)
	}
}

/*
 * Writes the password to the datastore
 */
func writeNewPassword(r *http.Request) {
	c := appengine.NewContext(r)
    k := datastore.NewKey(c, "TagPass", "pass", 0, nil)
    e := TagPass{tagstring}

    if _, err := datastore.Put(c, k, &e); err != nil {
        fmt.Printf("Failed to put password in datastore\n")
        return
    }
}

/*
 * Checks if we have a password in the datastore. If we don't
 * then we set a default (NOT SECURE. Demo purposes), otherwise
 * we fetch it.
 */
func checkpass(r *http.Request) {
	c := appengine.NewContext(r)
    k := datastore.NewKey(c, "TagPass", "pass", 0, nil)
    e := new(TagPass)
    if err := datastore.Get(c, k, e); err != nil {
        fmt.Printf("No password in datastore. Get failed")
    }

    if (len(e.Password) == 0) {
		fmt.Printf("No password in datastore. Putting in default\n")
		e.Password = "lemurtwelve"
	}
	tagstring = e.Password

    if _, err := datastore.Put(c, k, &e); err != nil {
        fmt.Printf("Failed to put password in datastore\n")
        return
    }
}

/*
 * Generates cryptographically random string
 * The string is converted to base64 so it can be 
 * printed, and then it becomes the new password
*/
func generateResponse(pw string) (string,bool) {
	if (trimQuotes(pw)==tagstring) {
		b := make([]byte, 15)
		n, err := io.ReadFull(rand.Reader, b)
		if n != len(b) || err != nil {
			fmt.Println("error:", err)
			return "",false
		}		
		en := base64.StdEncoding
        d := make([]byte, en.EncodedLen(len(b))) 
        en.Encode(d, b) 
        newpass := string(d)

		//we expect the newpass to be written to the tag
		//so we set it as the next password that we expect
		tagstring=newpass
		return newpass,true
	} 
	return "",false
}

/*
 * Function trims the quotes from a string
 */
func trimQuotes(x string) string {
	return strings.Join(strings.Split(x,"")[1:len(x)-1],"")
}

/*
 * Records the important parts of a submit request and 
 * puts it into the datastore so that it can be seen in the log
 */
func recordRequest(w http.ResponseWriter, r *http.Request) {
	stringpath := fmt.Sprintf("%#v",r.URL.Path)
	rawquery := fmt.Sprintf("%#v",r.URL.RawQuery)
	passwordstring := fmt.Sprintf("%#v",r.FormValue("password"))
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


/*
 * Goes through the last 100 requests in the datastore and 
 * prints them to the screen in log format.
 */
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
		cPath:=trimQuotes(records[key].Path)
		cQuery:=trimQuotes(records[key].RawQuery)
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