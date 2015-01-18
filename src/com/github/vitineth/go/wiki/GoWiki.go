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
	"com/github/vitineth/go/wiki/ResourceUtils"
	"com/github/vitineth/go/wiki/PageMarkdownUtils"
)

type Page struct {
	Title string
	Body  []byte
	Metadata *ResourceUtils.MetaData
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
	year, month, day := time.Now().Date()
	p.Metadata.LastSaveDate = string(day)+" "+month.String()+" "+string(year)

	ResourceUtils.SaveFileMetadata(p.Metadata, p.Title)

	//Create the content filename for the page
	filename := "data/contents/" + p.Title + ".txt"
	//Write the current page to file.
	return ioutil.WriteFile(filename, p.Body, 0600)
}

/*

 */
func loadPage (title string) (*Page, error){
	//Construct the filename from the page title
	filename := "data/contents/" + title + ".txt"
	//Get the metadata
	metadata, err := ResourceUtils.LoadFileMetadata(title)

	var meta *ResourceUtils.MetaData = &ResourceUtils.MetaData{
		LastSaveDate: "Unknown",
		LastSaveTime: "Unknown",
		PageCreationDate: "Unknown",
		PageCreationTime: "Unknown",
		Author: "Unknown",
		Views: -1}

	if error(err) == nil {
		meta = &ResourceUtils.MetaData{
			LastSaveDate: metadata.LastSaveDate,
			LastSaveTime: metadata.LastSaveTime,
			PageCreationDate: metadata.PageCreationDate,
			PageCreationTime: metadata.PageCreationTime,
			Author: metadata.Author,
			Views: metadata.Views}
	}
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body, Metadata: meta}, nil
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
	v := map[string]interface{}{
		"Title":    p.Title,
		"Body":        template.HTML(PageMarkdownUtils.ProcessPage(p.Body, isView)),
		"LastMod": p.Metadata.LastSaveDate,
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
