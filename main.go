package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

//go:embed assets
var staticFiles embed.FS

var dbpool *sqlitex.Pool
var assets http.Handler

var tmpl = template.Must(template.ParseFS(staticFiles, "assets/index.html"))

// linkMeta is our own derived metadata — NOT viewer state. Captured by the
// client at save time and stored verbatim in the meta column. Title/Thumbnail
// fill OG/Twitter tags; Canvas/ContentState are replayed into the X-IIIF-*
// discovery headers. Region xywh in ContentState depends on the live browser
// viewport, so it must be computed client-side, not reconstructed in Go.
type linkMeta struct {
	Title        string `json:"title"`
	Thumbnail    string `json:"thumbnail"`
	Canvas       string `json:"canvas,omitempty"`
	ContentState string `json:"contentState,omitempty"`
}

type pageData struct {
	ID   string
	Data string
	Meta *linkMeta
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		assets.ServeHTTP(w, r)
		return
	}

	if err := tmpl.Execute(w, pageData{}); err != nil {
		slog.Error("error", "err", err)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	conn, _ := dbpool.Take(r.Context())
	if conn == nil {
		return
	}
	defer dbpool.Put(conn)

	id := r.PathValue("id")

	var state, metaRaw, manifest string
	var page int
	results := false
	err := sqlitex.ExecuteTransient(conn, "SELECT data, meta, json_extract(data,'$.manifestUrl'), json_extract(data,'$.pages[0]') FROM links WHERE public_id=?;",
		&sqlitex.ExecOptions{
			Args: []interface{}{id},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				state = stmt.ColumnText(0)
				metaRaw = stmt.ColumnText(1)
				manifest = stmt.ColumnText(2)
				page = stmt.ColumnInt(3)
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

	var m linkMeta
	hasMeta := metaRaw != "" && metaRaw != "null" && json.Unmarshal([]byte(metaRaw), &m) == nil

	// X-IIIF-* discovery headers
	if manifest != "" {
		w.Header().Set("X-IIIF-Manifest", manifest)
	}
	if page > 0 {
		w.Header().Set("X-IIIF-Page", strconv.Itoa(page))
	}
	if hasMeta && m.Canvas != "" {
		w.Header().Set("X-IIIF-Canvas", m.Canvas)
	}
	if hasMeta && m.ContentState != "" {
		w.Header().Set("X-IIIF-Content-State", m.ContentState)
	}

	var meta *linkMeta
	if hasMeta && (m.Title != "" || m.Thumbnail != "") {
		if m.Title == "" {
			m.Title = "IIIF document"
		}
		meta = &m
	}

	if err = tmpl.Execute(w, pageData{ID: id, Data: state, Meta: meta}); err != nil {
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

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			slog.Error("save", "err", err)
			http.Error(w, "error", http.StatusBadRequest)
			return
		}

		var payload struct {
			State json.RawMessage `json:"state"`
			Meta  json.RawMessage `json:"meta"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			slog.Error("save", "err", err)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		var probe interface{}
		if err := json.Unmarshal(payload.State, &probe); err != nil {
			slog.Error("save", "err", err)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		state := payload.State

		var meta []byte
		if len(payload.Meta) > 0 && string(payload.Meta) != "null" {
			meta = []byte(payload.Meta)
		}

		public_id, err := gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyz", 12)
		if err != nil {
			slog.Error("save", "err", err)
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		err = sqlitex.Execute(conn, "INSERT INTO links (public_id, data, meta) VALUES (?,?,?)",
			&sqlitex.ExecOptions{
				Args: []interface{}{public_id, []byte(state), meta},
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

	slog.Info("iiif.link # http://localhost:3000")
	err = http.ListenAndServe(":3000", mux)
	if err != nil {
		log.Fatal(err)
	}
}
