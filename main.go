package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/julienschmidt/httprouter"
	"github.com/segmentio/ksuid"
	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/go-playground/validator.v9"
)

// View ....
type View struct {
	Label    string  `json:"label"`
	Manifest string  `json:"manifest" validate:"required,url"`
	Canvas   string  `json:"canvas" validate:"required,url"`
	Page     *int    `json:"page" validate:"required"`
	Zoom     float64 `json:"zoom" validate:"required"`
	Viewport struct {
		X float64 `json:"x" validate:"required"`
		Y float64 `json:"y" validate:"required"`
	} `json:"viewport"`
	Bounds struct {
		X *int `json:"x" validate:"required"`
		Y *int `json:"y" validate:"required"`
		W *int `json:"w" validate:"required"`
		H *int `json:"h" validate:"required"`
	} `json:"bounds"`
}

// Store ...
type Store struct {
	DB *leveldb.DB
}

// NewStore ...
func NewStore(path string) (*Store, error) {
	var err error
	store := Store{}
	store.DB, err = leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	return &store, nil
}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tmpl, _ := template.New("index.html").Delims("[[", "]]").ParseFiles("index.html")
	tmpl.Execute(w, nil)
}

func (store *Store) get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	data, err := store.DB.Get([]byte(id), nil)

	if err == leveldb.ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404"))
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500"))
		return
	}

	var view View
	json.Unmarshal(data, &view)

	type Output struct {
		Data  string
		Label string
		Image string
	}
	image := fmt.Sprintf("%s/%d,%d,%d,%d/,100/0/default.jpg", view.Canvas,
		*view.Bounds.X, *view.Bounds.Y, *view.Bounds.W, *view.Bounds.H)

	tmpl, _ := template.New("index.html").Delims("[[", "]]").ParseFiles("index.html")
	tmpl.Execute(w, Output{Data: string(data), Label: view.Label, Image: image})
}

func (store *Store) header(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	data, err := store.DB.Get([]byte(id), nil)

	if err == leveldb.ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404"))
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500"))
		return
	}
	var view View
	json.Unmarshal(data, &view)

	w.Header().Add("X-IIIF-label", view.Label)
	w.Header().Add("X-IIIF-manifest", view.Manifest)
	w.Header().Add("X-IIIF-canvas", view.Canvas)
	w.Header().Add("X-IIIF-page", strconv.Itoa(*view.Page))
	w.Header().Add("X-IIIF-Image", fmt.Sprintf("%s/%d,%d,%d,%d/,100/0/default.jpg", view.Canvas,
		*view.Bounds.X, *view.Bounds.Y, *view.Bounds.W, *view.Bounds.H))
}

func (store *Store) save(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500"))
		return
	}

	var view View
	json.Unmarshal(body, &view)
	validate := validator.New()
	err = validate.Struct(view)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
		return
	}

	id := ksuid.New()
	err = store.DB.Put([]byte(id.String()), []byte(body), nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500"))
		return
	}
	fmt.Fprintf(w, id.String())
}

func main() {

	store, err := NewStore("./data")
	defer store.DB.Close()

	if err != nil {
		log.Fatal(err)
	}

	router := httprouter.New()
	router.GET("/", index)
	router.GET("/id/:id", store.get)
	router.HEAD("/id/:id", store.header)
	router.POST("/save", store.save)
	log.Fatal(http.ListenAndServe(":8080", router))
}
