package rest

import (
	"encoding/json"
	"fmt"
	"github.com/fafeitsch/Open-Traffic-Sandbox/pkg/model"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type api struct {
	lineModel model.LineModel
	gps       model.RouteService
}

func headers(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				log.Printf("Panic during request \"%s\": %v", r.URL, p)
				errorResponse(w, http.StatusInternalServerError, "internal server error, please see log")
			}
		}()
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Accept", "application/json")
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

const apiPrefix = "/api"

// NewRouter creates an http router for the REST Api.
func NewRouter(lineModel model.LineModel, gps model.RouteService) http.Handler {
	api := api{lineModel: lineModel, gps: gps}
	router := mux.NewRouter()
	router.Handle(apiPrefix+"/lines", headers(api.getLines))
	router.Handle(apiPrefix+"/lines/{key}", headers(api.getLine))
	router.Handle(apiPrefix+"/lines/{key}/route", headers(api.getRoute))
	return router
}

type restError struct {
	Error string `json:"error"`
}

func errorResponse(w http.ResponseWriter, status int, text string, args ...interface{}) {
	message := restError{Error: fmt.Sprintf(text, args...)}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(message)
}
