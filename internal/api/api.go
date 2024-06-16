package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/root-gabriel/ya/internal/handlers"
	"github.com/root-gabriel/ya/internal/storage"
)

// APIServer представляет сервер API
type APIServer struct {
	Router  *mux.Router
	Storage *storage.MemStorage
}

// NewServer создает новый экземпляр сервера
func NewServer() *APIServer {
	storage := storage.NewMem()
	server := &APIServer{
		Router:  mux.NewRouter(),
		Storage: storage,
	}
	server.routes()
	return server
}

// routes инициализирует маршруты сервера
func (s *APIServer) routes() {
	s.Router.HandleFunc("/update/counter/{name}/{value}", handlers.UpdateCounterHandler(s.Storage)).Methods("POST")
	s.Router.HandleFunc("/value/counter/{name}", handlers.GetCounterValueHandler(s.Storage)).Methods("GET")
	s.Router.HandleFunc("/ping", handlers.PingHandler(s.Storage)).Methods("GET")
}

// ServeHTTP реализует интерфейс http.Handler для APIServer
func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

