package main

import (
	"fmt"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"errors"
	"time"
	"bytes"
	"bufio"
	"os"
)

type Page struct {
	Title string
	Body  []byte
	Author string
	LastEdited string
}
var addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
var templates = template.Must(template.ParseFiles("data/templates/edit.html", "data/templates/view.html", "data/templates/testWiki.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

/*
 This function will save the given page and its metadata to file

 <br><br>

 p - A pointer to the page reference you want to save
 */
func (p *Page) save() error {
	//Set the last edit time of the page to the current time as is we are saving we will probably
	//have just edited it.
	p.LastEdited = time.Now().String()
	//Create the content filename for the page
	filename := "data/contents/" + p.Title + ".txt"
	//Write the current page to file.
	return ioutil.WriteFile(filename, p.Body, 0600)
}

/*

 */
func loadPage (title string) (*Page, error){
	filename := "data/contents/" + title + ".txt"
	metadata, err := loadFileMetadata(title)
	var lastEdit string = "Unknown"
	var author string = "Unknown"
	if err != nil {
		lastEdit = metadata[0]
		author = metadata[1]
	}
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body, LastEdited: lastEdit, Author: author}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	if title == "testWiki" {
		renderTemplate(w, "testWiki", p, true)
	}else {
		renderTemplate(w, "view", p, true)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p, false)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page, isView bool) {
	var v map[string]interface{}
	if isView {
		v = map[string]interface{}{
			"Title":    p.Title,
			"Body":        template.HTML(processPage(p.Body)),
			"LastMod": p.LastEdited,
		}
	}else{
		v = map[string]interface{}{
			"Title":    p.Title,
			"Body":        template.HTML(reverseProcessPage(p.Body)),
			"LastMod": p.LastEdited,
		}
	}
	var err error = templates.ExecuteTemplate(w, tmpl+".html", v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}

func reverseProcessPage(body []byte) ([]byte){
	boldStartRegex := regexp.MustCompile("<strong>")
	boldFinishRegex := regexp.MustCompile("<\\/strong>")

	italicsStartRegex := regexp.MustCompile("<em>")
	italicsFinishRegex := regexp.MustCompile("<\\/em>")

	linkBeginRegex := regexp.MustCompile("<a href=\"/view/")
	linkCenterRegex := regexp.MustCompile(">")
	linkFinishRegex := regexp.MustCompile("<\\/a>")

	body = bytes.Replace(body, []byte("<br>"), []byte("\n"), -1)

	body = boldStartRegex.ReplaceAll(body, []byte("*\\"))
	body = boldFinishRegex.ReplaceAll(body, []byte("/*"))

	body = italicsStartRegex.ReplaceAll(body, []byte("_\\"))
	body = italicsFinishRegex.ReplaceAll(body, []byte("/_"))

	body = linkBeginRegex.ReplaceAll(body, []byte("\\["))
	body = linkFinishRegex.ReplaceAll(body, []byte("]/"))
	body = linkCenterRegex.ReplaceAll(body, []byte("]["))


	return body
}

func processPage(body []byte) ([]byte){
	boldStartRegex := regexp.MustCompile("(\\*)\\\\")
	boldFinishRegex := regexp.MustCompile("\\/(\\*)")

	italicsStartRegex := regexp.MustCompile("_\\\\")
	italicsFinishRegex := regexp.MustCompile("\\/_")

	linkBeginRegex := regexp.MustCompile("\\\\\\[")
	linkCenterRegex := regexp.MustCompile("\\]\\[")
	linkFinishRegex := regexp.MustCompile("\\]\\/")

	body = bytes.Replace(body, []byte("\\n"), []byte("<br>"), -1)

	body = boldStartRegex.ReplaceAll(body, []byte("<strong>"))
	body = boldFinishRegex.ReplaceAll(body, []byte("</strong>"))

	body = italicsStartRegex.ReplaceAll(body, []byte("<em>"))
	body = italicsFinishRegex.ReplaceAll(body, []byte("</em>"))

	body = linkBeginRegex.ReplaceAll(body, []byte("<a href=\"/view/"))
	body = linkCenterRegex.ReplaceAll(body, []byte("\">"))
	body = linkFinishRegex.ReplaceAll(body, []byte("</a>"))

	fmt.Println(string(body))

	return body
}

func loadFileMetadata(pageName string) (metaData []string, err error) {
	filename := "data/meta/" + pageName + ".txt"
	reader, error := os.Open(filename)
	if error != nil {
		return nil, error
	}
	bufReader := bufio.NewReader(reader)
	lastEdited, _, error := bufReader.ReadLine()
	author, _, error := bufReader.ReadLine()

	var returnVal []string
	returnVal[0] = string(lastEdited)
	returnVal[1] = string(author)

	return returnVal, nil
}

func main() {
	fmt.Println("Go WIKI Http Server....")
	fmt.Println("\nInitializing HTTP server..")
	flag.Parse()
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	if *addr {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("final-port.txt", []byte(l.Addr().String()), 0644)
		if err != nil {
			log.Fatal(err)
		}
		s := &http.Server{}
		s.Serve(l)
		return
	}

	fmt.Println("Starting HTTP server..")
	http.ListenAndServe(":8080", nil)

}
