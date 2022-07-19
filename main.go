package main

import (
	"crypto/md5"
	"log"
	"net/http"
	"time"

	"github.com/hoodyman/httpdbexample/commonvars"
	"github.com/hoodyman/httpdbexample/db"
	"github.com/hoodyman/httpdbexample/handlers"
	"github.com/hoodyman/simplewebtools"
)

func main() {

	err := db.InitDb()
	if err != nil {
		log.Fatalln(err)
	}
	defer db.CloseDb()

	commonvars.Templ = simplewebtools.TemplateHolder{}
	err = commonvars.Templ.LoadTemplates("templates")
	if err != nil {
		log.Fatalln(err)
	}

	commonvars.Csrf = simplewebtools.TokenHolder{}
	commonvars.Csrf.Start(time.Minute*60, time.Second, 64, md5.New())

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handlers.HandlerIndex)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, req *http.Request) {})

	log.Fatal(http.ListenAndServe(":8080", nil))

}
