package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/http"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

//go:embed assets
var staticFiles embed.FS

var dbpool *sqlitex.Pool
var assets http.Handler

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		assets.ServeHTTP(w, r)
		return
	}

	tmpl, err := template.ParseFS(staticFiles, "assets/index.html")
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(w, nil)

	if err != nil {
		slog.Error("error", "err", err)
	}
}

func redirect(w http.ResponseWriter, r *http.Request) {
	var redir string
	conn, _ := dbpool.Take(r.Context())
	if conn == nil {
		return
	}
	defer dbpool.Put(conn)

	id := r.PathValue("id")

	err := sqlitex.ExecuteTransient(conn, "SELECT urlparams FROM links WHERE public_id=?;",
		&sqlitex.ExecOptions{
			Args: []interface{}{id},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				redir = stmt.ColumnText(0)
				return nil
			},
		})
	if err != nil {
		slog.Error("error", "err", err)
		http.Error(w, "error", http.StatusInternalServerError)
	}

	slog.Info("redirect", "id", id)
	http.Redirect(w, r, fmt.Sprintf("/%s", redir), http.StatusSeeOther)
}

func get(w http.ResponseWriter, r *http.Request) {
	conn, _ := dbpool.Take(r.Context())
	if conn == nil {
		return
	}
	defer dbpool.Put(conn)

	id := r.PathValue("id")

	var state string
	results := false
	err := sqlitex.ExecuteTransient(conn, "SELECT data FROM links WHERE public_id=?;",
		&sqlitex.ExecOptions{
			Args: []interface{}{id},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				state = stmt.ColumnText(0)
				results = state != "null"
				return nil
			},
		})
	if err != nil {
		slog.Error("error", "err", err)
		http.Error(w, "error", http.StatusInternalServerError)
	}
	if !results {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	type Output struct {
		ID   string
		Data string
	}

	tmpl, err := template.ParseFS(staticFiles, "assets/index.html")
	if err != nil {
		log.Fatal(err)
	}
	err = tmpl.Execute(w, Output{ID: id, Data: state})

	if err != nil {
		slog.Error("error", "err", err)
	}

}

func save(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://iiif.link")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
	if r.Method == "POST" {
		conn, _ := dbpool.Take(r.Context())
		if conn == nil {
			return
		}
		defer dbpool.Put(conn)

		err := r.ParseMultipartForm(1)
		if err != nil {
			slog.Error("save", "err", err)
			http.Error(w, "error", http.StatusBadRequest)
			return
		}

		params := r.FormValue("tifyParams")
		jsonparams, _ := Tify(params)

		public_id, err := gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyz", 12)
		if err != nil {
			slog.Error("save", "err", err)
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		err = sqlitex.Execute(conn, "INSERT INTO links (public_id, urlparams, data) VALUES (?,?,?)",
			&sqlitex.ExecOptions{
				Args: []interface{}{public_id, params, jsonparams.Json()},
			})
		if err != nil {
			slog.Error("save", "err", err)
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		slog.Info("new link", "id", public_id)
		fmt.Fprint(w, public_id)
	}
}

func main() {
	var err error
	dbpool, err = sqlitex.NewPool("file:iiiflink.db?cache=shared&mode=rwc&_journal_mode=WAL", sqlitex.PoolOptions{
		PoolSize: 10,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer dbpool.Close()

	var staticFS = fs.FS(staticFiles)
	htmlContent, err := fs.Sub(staticFS, "assets")
	if err != nil {
		log.Fatal(err)
	}
	assets = http.FileServer(http.FS(htmlContent))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", index)
	mux.HandleFunc("POST /save", save)
	mux.HandleFunc("GET /id/{id}", get)
	mux.HandleFunc("GET /r/{id}", redirect)

	slog.Info("iiif.link # http://localhost:3000")
	err = http.ListenAndServe(":3000", mux)
	if err != nil {
		log.Fatal(err)
	}
}
