package main

import (
	. "fmt"
	"net/http"
	"net/url"
	"io"
	"bytes"
	"html/template"
	"log"
	"strings"
	"path/filepath"
	"os"
	"mime"
)

// Shorthand
type M map[string]interface {}

func render(w http.ResponseWriter, tmpl string, data map[string]interface{}) {
	// Get full file path.
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))

	if err != nil {
		log.Fatal(err)
	}

	templateFile := dir + string(filepath.Separator) + tmpl

	Println(templateFile)

	// Attempt to parse the file.
	tmpl = Sprintf("%s", templateFile)
	t, err := template.ParseFiles(tmpl)

	if err != nil{
		log.Print("template parsing error: ", err)
	}

	// We pass our data map to the template
	err = t.Execute(w, data)

	if err != nil{
		log.Print("template executing error: ", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	render(w, "form.html", map[string]interface{} {
		"Title": "JSON to PDF",
	})
}

func requestBroker(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	requestUrl := strings.Join(r.Form["url"], "")

	if requestUrl == "" {
		render(w, "form.html", map[string]interface{} {
			"Error": template.HTML(`<div class="alert alert-danger" role="alert">Url required.</div>`),
		})
		return
	}

	apiUrl := requestUrl
	resource := "/v0/fetch"
	var jsonString = strings.Join(r.Form["json"], "")

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := Sprintf("%v", u)

	client := &http.Client{}

	req, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(jsonString))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := client.Do(req)
	download := strings.Join(r.Form["download"], "");

	if download == "on" && resp.Header.Get("Content-Type") != "application/json" {
		disposition := resp.Header.Get("Content-Disposition");

		// Handle the Content-Disposition another way.
		if disposition  == "" {
			for _, value := range resp.Header {
				value := value[0];
				if len(value) > 19 && value[:19] == "Content-Disposition" {
					disposition = value[20:len(value)] // 20 to include the colon
				}
			}
		}

		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
		_, params, _ := mime.ParseMediaType(disposition)
		filename := params["filename"]
		w.Header().Set("Content-Disposition", "attachment; filename=" + filename + ";")
	}

	io.Copy(w, resp.Body)
}

func main() {
	Println("Running server @ localhost:6800")
	http.HandleFunc("/", handler)
	http.HandleFunc("/processor", requestBroker)
	http.ListenAndServe(":6800", nil)
}
