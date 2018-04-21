package ktl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// DefaultListen is host/port ktl will listen on by default.
const DefaultListen = "localhost:8543"

// Database describes the interface implemented by database backends such as mongo,
// providing access to storing and retrieving batches.
type Database interface {
	CreateBatch(context.Context, *Batch) error
	ListBatches(context.Context, *BatchListOptions) ([]*Batch, error)
	GetBatch(ctx context.Context, id string) (*Batch, error)
	UpdateBatch(context.Context, *Batch) error
}

// ErrNotFound is returned by database backends when an item can't be found.
var ErrNotFound = fmt.Errorf("not found")

// Serve runs a HTTP server, serving the ktl REST API.
func Serve(db Database) error {
	s := newServer(db)
	ui := newUI()
	http.Handle("/ui/", http.StripPrefix("/ui", ui))
	http.Handle("/v0/", http.StripPrefix("/v0", s))
	log.Println("Listening on", DefaultListen)
	return http.ListenAndServe(DefaultListen, nil)
}

func newUI() http.Handler {
	return http.FileServer(http.Dir("ui"))
}

// server provides REST API endpoint handling.
type server struct {
	db     Database
	router *mux.Router
}

func newServer(db Database) *server {
	s := &server{db: db}
	r := mux.NewRouter().StrictSlash(true)

	r.Path("/batch").
		Methods("POST").
		Name("CreateBatch").
		HandlerFunc(s.createBatch)

	r.Path("/batch").
		Methods("GET").
		Name("ListBatches").
		HandlerFunc(s.listBatches)

	r.Path("/batch/{id}").
		Methods("GET").
		Name("GetBatch").
		HandlerFunc(s.getBatch)

	r.Path("/batch/{batchID}/step/{stepID}:restart").
		Methods("POST").
		Name("RestartStep").
		HandlerFunc(s.restartStep)

	s.router = r
	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

// TODO there's a bit of boilerplate in these handlers that could probably
//      be usefully abstracted.

// createBatch creates a new Batch entity in the database and returns its ID.
func (s *server) createBatch(w http.ResponseWriter, req *http.Request) {

	batch := &Batch{}
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()

	err := dec.Decode(batch)
	if err != nil {
		http.Error(w, "error decoding Batch from request body", http.StatusBadRequest)
		return
	}

	err = ValidateBatch(batch)
	if err != nil {
		http.Error(w, fmt.Sprintf("validation error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	batch.ID = NewBatchID()
	batch.CreatedAt = time.Now()

	err = s.db.CreateBatch(req.Context(), batch)
	if err != nil {
		http.Error(w, fmt.Sprintf("error saving batch: %s", err.Error()),
			http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(CreateBatchResponse{ID: batch.ID})
	if err != nil {
		http.Error(w, "error encoding response body", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}

func (s *server) listBatches(w http.ResponseWriter, req *http.Request) {

	batches, err := s.db.ListBatches(req.Context(), &BatchListOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("error listing batches: %s", err), http.StatusInternalServerError)
		return
	}
	// TODO test
	if batches == nil {
		batches = []*Batch{}
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(BatchList{Batches: batches})
	if err != nil {
		http.Error(w, "error encoding response body", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}

func (s *server) getBatch(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	batch, err := s.db.GetBatch(req.Context(), id)
	if err == ErrNotFound {
		http.Error(w, "batch not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting batch: %s", err), http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)
	err = enc.Encode(batch)
	if err != nil {
		http.Error(w, "error encoding response body", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}

func (s *server) restartStep(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	batchID := vars["batchID"]
	stepID := vars["stepID"]
	err := RestartStep(req.Context(), s.db, batchID, stepID)
	if err == ErrNotFound {
		http.Error(w, "batch not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("error restarting step: %s", err), http.StatusInternalServerError)
		return
	}
	return
}
