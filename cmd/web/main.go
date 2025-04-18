package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"soulgame/pkg/models/postgresql"

	_ "github.com/lib/pq"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	quests        *postgresql.QuestModel
	templateCache map[string]*template.Template
}

func main() {

	addr := flag.String("addr", ":443", "Сетевой адрес HTTPS")
	certFile := flag.String("certfile", "certificate.pem", "certificate PEM file")
	keyFile := flag.String("keyfile", "private_key.pem", "key PEM file")
	//certFile := "/etc/letsencrypt/live/soulgame-bp.ru/fullchain.pem"
	//keyFile := "/etc/letsencrypt/live/soulgame-bp.ru/privkey.pem"
	dsn := flag.String(
		"dsn",
		"host=79.174.88.25 port=18561 user=soulgame password=b0712200sV__ dbname=soulgame sslmode=disable",
		"Параметры подключения к БД",
	)

	//dsn := flag.String("dsn", "user=soulgame password=b0712200sV__ dbname=soulgame sslmode=disable", "Параметры подключения к бд")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		errorLog.Fatal(err)
	}

	ScheduleDailyUpdate(db)

	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		quests:        &postgresql.QuestModel{DB: db},
		templateCache: templateCache,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS13,
			PreferServerCipherSuites: false,
		},
	}

	infoLog.Println("Запуск сервера на", *addr)
	err = srv.ListenAndServeTLS(*certFile, *keyFile)
	errorLog.Fatal(err)
}
