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
	Accepted string
	ResponsePassword string
}

var logTemplate = template.Must(template.New("log.html").ParseFiles("templates/log.html"))

var tagstring string
var lastValidPassword string

type TagPass struct {
    Password string
}

func init() {
	http.HandleFunc("/",root)
	http.HandleFunc("/log",logprinter)
	http.HandleFunc("/submit",submitcode)

}

func root(w http.ResponseWriter, r *http.Request) {
	if err := logTemplate.Execute(w,nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func logprinter(w http.ResponseWriter, r *http.Request) {
	printLogToHTML(w,r)
}

func submitcode(w http.ResponseWriter, r *http.Request) {
	//get the password we expect from the datastore
	checkpass(r)
	checkLastPass(r)
	//check if the submitted password is correct
	passwordstring := fmt.Sprintf("%#v",r.FormValue("password"))
	responsestring,accepted := generateResponse(passwordstring)
	//put the request recieved in the log
	
	if (accepted) {
		recordRequest(w,r,true,responsestring)
		tagstring=responsestring;
		fmt.Fprintf(w,"ACCEPTED. Please write this to the tag: %s\n",
			responsestring)
		lastValidPassword=responsestring;
		//If we are guaranteed no MITM attacks via a on-phone TPM or 
		//some magical timing mechanism on the tag, we can do this:
		got_valid_forgive_fixer(w,r,passwordstring)
		
		writeNewPassword(r)
		writeLastValidPassword(r)
	} else {
		if (trimQuotes(passwordstring) == lastValidPassword) {
			//we will accept the last old password as well
			//in case something fishy happened
			recordRequest(w,r,true,responsestring)
			fmt.Fprintf(w,"ACCEPTED. Please write this to the tag: %s\n",
				responsestring)
			tagstring=responsestring;
			lastValidPassword=responsestring;
			writeNewPassword(r)
			writeLastValidPassword(r)
			return
		}
		recordRequest(w,r,false,responsestring)
		fmt.Fprintf(w,"REJECTED.\n")
		fmt.Fprintf(w,"write this to the tag: %s\n",responsestring)
		tagstring=responsestring;
		writeNewPassword(r)
	}
}

/**
 * If we can guarantee that the codes that the webserver gives
 * can't be accessed in plaintext, we can use this method to 
 * guarantee that the person that fixes the tag that a malicious
 * user breaks gets their request accepted.
 *
 * This guarantees that NO good users get their requests rejected, 
 * while at the same time, maintaining the same security guarantees
 * as before. 
 **/
func got_valid_forgive_fixer(w http.ResponseWriter, 
	r *http.Request, password string) {
	c := appengine.NewContext(r)
	//we set a limit of 1000...because that sounds reasonable for a 
	//demo. Definitely not production quality code here.
	q := datastore.NewQuery("Record").Order("-Time").Limit(1000)
    records := make([]RequestRecord, 0, 10)
    
	if _, err := q.GetAll(c, &records); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}
	password = trimQuotes(password)
	//we want to find the one that had a submitted password
	//equal to the one that was passed to this function
	for key := range records {
		if (records[key].ResponsePassword == password) {
			req := RequestRecord{
			RemoteAddr:records[key].RemoteAddr,
			Host:records[key].Host,
			Path:records[key].Path,
			RawQuery:records[key].RawQuery,
			Time:records[key].Time,
			Password:records[key].Password,
			Accepted:"true",
			ResponsePassword:records[key].ResponsePassword,
			}
			fmt.Printf("Found someone that wrote it\n")
			//err := datastore.Delete(c, records[key].Key)
			//if err != nil {
			//	http.Error(w, err.Error(), http.StatusInternalServerError)
			//return
			//}
			
			_, err := datastore.Put(c, datastore.NewIncompleteKey(c,"Record", nil), &req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
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
 * Writes the last password given to valid request to the datastore
 */
func writeLastValidPassword(r *http.Request) {
	c := appengine.NewContext(r)
    k := datastore.NewKey(c, "TagPass", "lastvalidpass", 0, nil)
    e := TagPass{lastValidPassword}

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
		if _, err := datastore.Put(c, k, &e); err != nil {
			fmt.Printf("Failed to put password in datastore\n")
			return
		}
	}
	tagstring = e.Password
}

/*
 * Checks if we have the last valid password in the datastore. 
 * If we don't then we set it to an empty string, otherwise
 * we fetch it.
 */
func checkLastPass(r *http.Request) {
	c := appengine.NewContext(r)
    k := datastore.NewKey(c, "TagPass", "lastvalidpass", 0, nil)
    e := new(TagPass)
    if err := datastore.Get(c, k, e); err != nil {
        fmt.Printf("No password in datastore. Get failed")
    }

    if (len(e.Password) == 0) {
		fmt.Printf("No password in datastore. Putting in default\n")
		e.Password = ""
		if _, err := datastore.Put(c, k, &e); err != nil {
			fmt.Printf("Failed to put password in datastore\n")
			return
		}
	}
	lastValidPassword = e.Password
}

/*
 * Generates cryptographically random string
 * The string is converted to base64 so it can be 
 * printed, and then it becomes the new password
*/
func generateResponse(pw string) (string,bool) {
	//we make a new password regardless of whether we get
	//the right password or a wrong one
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

	if (trimQuotes(pw)==tagstring) {
		return newpass,true
	} 
	return newpass,false
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
func recordRequest(w http.ResponseWriter, r *http.Request, a bool, resp string) {
	stringpath := fmt.Sprintf("%#v",r.URL.Path)
	rawquery := fmt.Sprintf("%#v",r.URL.RawQuery)
	passwordstring := fmt.Sprintf("%#v",r.FormValue("password"))
	c := appengine.NewContext(r)

	var acceptedString string
	if (a) { 
		acceptedString = "true"
	} else {
		acceptedString = "false"
	}
	
	req := RequestRecord{
	RemoteAddr:r.RemoteAddr,
	Host:r.Host,
	Path:stringpath,
	RawQuery:rawquery,
	Time:time.Now(),
	Password:passwordstring,
	Accepted:acceptedString,
	ResponsePassword:resp,
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
			if (records[key].Accepted == "true" ) {
				fmt.Fprintf(w,"<div class=\"accept\">")
			} else {
				fmt.Fprintf(w,"<div class=\"reject\">")
			}
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
			fmt.Fprintf(w,"</div></div></div>")
			counter++
		}
	}
}