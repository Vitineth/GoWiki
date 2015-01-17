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
	"io"
)

type Page struct {
	Title string
	Body  []byte
	Author string
	LastEdited string
}
var addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
var templates = template.Must(template.ParseFiles("edit.html", "view.html", "testWiki.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func (p *Page) save() error {
	p.LastEdited = time.Now().String()
	filename := p.Title + ".txt"
	data := append([]byte(p.LastEdited + "\n"), p.Body...)
	return ioutil.WriteFile(filename, data, 0600)
}

func loadPage (title string) (*Page, error){
	filename := title + ".txt"
	reader, err := os.Open(filename)
	dateLine, body := read(reader)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body, LastEdited: string(dateLine)}, nil
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

	body = boldStartRegex.ReplaceAll(body, []byte("*\\"))
	body = boldFinishRegex.ReplaceAll(body, []byte("/*"))

	body = italicsStartRegex.ReplaceAll(body, []byte("_\\"))
	body = italicsFinishRegex.ReplaceAll(body, []byte("/_"))

	body = linkBeginRegex.ReplaceAll(body, []byte("\\["))
	body = linkFinishRegex.ReplaceAll(body, []byte("]/"))
	body = linkCenterRegex.ReplaceAll(body, []byte("]["))

	body = bytes.Replace(body, []byte("<br>"), []byte("\n"), -1)

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

	body = bytes.Replace(body, []byte("\n"), []byte("<br>"), -1)

	body = boldStartRegex.ReplaceAll(body, []byte("<strong>"))
	body = boldFinishRegex.ReplaceAll(body, []byte("</strong>"))

	body = italicsStartRegex.ReplaceAll(body, []byte("<em>"))
	body = italicsFinishRegex.ReplaceAll(body, []byte("</em>"))

	body = linkBeginRegex.ReplaceAll(body, []byte("<a href=\"/view/"))
	body = linkCenterRegex.ReplaceAll(body, []byte("\">"))
	body = linkFinishRegex.ReplaceAll(body, []byte("</a>"))

	return body
}

func read(reader io.Reader)(date []byte, text []byte){
	text = []byte("zx")
	var newReader *bufio.Reader = bufio.NewReader(reader)
	var isFirst bool = true
	for true {
		if isFirst {
			date, _, _ = newReader.ReadLine()
			isFirst = false
		}
		temp , _, err:= newReader.ReadLine()
		if err != nil {
			break
		}
		text = append(text, temp...)
	}
	return date, text
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
