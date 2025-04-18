package main

import (
	"net/http"
	"path/filepath"
)

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/snippet", app.showSnippet)
	mux.HandleFunc("/new", app.newUser)
	mux.HandleFunc("/login", app.login)
	mux.HandleFunc("/shop", app.shop)
	mux.HandleFunc("/admin", app.admin)
	mux.HandleFunc("/unlogin", app.unlog)
	mux.HandleFunc("/create/quest", app.createquest)
	mux.HandleFunc("/completequest", app.complete)
	mux.HandleFunc("/profile", app.profile)
	mux.HandleFunc("/purchase", app.purchase)
	mux.HandleFunc("/repass", app.repass)
	mux.HandleFunc("/complete", app.ready)
	mux.HandleFunc("/checkin", app.checkin)
	mux.HandleFunc("/log", app.log)
	mux.HandleFunc("/create/item", app.createitem)
	mux.HandleFunc("/helpclient", app.helpclient)
	mux.HandleFunc("/take/items", app.takeItems)
	mux.HandleFunc("/completeitem", app.completeitem)
	mux.HandleFunc("/log/item", app.logitem)
	mux.HandleFunc("/delete/item", app.deleteitem)
	mux.HandleFunc("/delitem", app.delitem)
	mux.HandleFunc("/delete/quest", app.deletequest)
	mux.HandleFunc("/delquest", app.delquest)

	fileServer := http.FileServer(neuteredFileSystem{http.Dir("./ui/static/")})
	mux.Handle("/static", http.NotFoundHandler())
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	return mux
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, _ := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}
