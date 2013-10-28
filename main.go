package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

const HTTP_OK = "HTTP/1.1 200 OK"
const HTTP_OK_NO_CONTENT = "HTTP/1.1 204 No Content"
const HTTP_404 = "HTTP/1.1 404 Not Found"

var currpath, _ = os.Getwd()

type Page struct {
	Title string
	Body  []byte
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	fakefile := r.URL.Path[utf8.RuneCountInString(api_path):]
	filename := strings.Replace(fakefile, "/", "_", -1) + ".html"
	path := filepath.Clean(currpath + "/.." + api_path + filename)
	body, err := ioutil.ReadFile(path)
	log.Println("Getting API: " + path)
	if err != nil {
		http.Error(w, fmt.Sprintf("API, 404 Not Found : %s", err), 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		w.Write(body)
	case "PUT":
		w.Write(body)
	case "POST":
		w.Write(body)
	case "DELETE":
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Clean(r.URL.Path[1:])
	if filename == "." {
		filename = "index.html"
	}
	body, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		filename = "../" + filename
		body, err = ioutil.ReadFile(filename)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("File, 404 Not Found : %s", err), 404)
		return
	}
	if filepath.Ext(filename) == ".html" {
		log.Print("Found HTML file, will be parsing for include")
		fmt.Fprintf(w, "%s", includeFiles(body))
		return
	}
	fmt.Fprintf(w, "%s", body)
}
func loadInclude(filename string) string {
	body, err := ioutil.ReadFile(".." + filename)
	if err != nil {
		return "Could not find view: " + filename
	}
	return string(body[:len(body)])
}

func includeFiles(body []byte) string {
	includeRegex, err := regexp.Compile(`^\s*include\("(.+?)"\);`)
	if err != nil {
		log.Fatal(err)
	}
	bodys := string(body[:len(body)])
	toreturn := bodys
	for _, line := range strings.Split(bodys, "\n") {
		if includeRegex.Match([]byte(line)) {
			includeFile := includeRegex.FindStringSubmatch(line)
			log.Println("INCLUDE", includeFile[1])
			content := loadInclude(includeFile[1])
			toreturn = strings.Replace(toreturn, includeFile[0], content, -1)
		}
	}
	return toreturn
}

var api_path = "/api/v1/"

func main() {
	log.SetOutput(os.Stderr)
	http.HandleFunc(api_path, apiHandler)
	http.HandleFunc("/", fileHandler)
	log.Println("Starting server on port 8080, go to http://localhost:8080/index.html")
	http.ListenAndServe(":8080", nil)
}
