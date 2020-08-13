package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"

	"github.com/gorilla/handlers"
)

var (
	qLock *sync.RWMutex
	queue = []string{}
)

// request type that we expect from a post request
type request struct {
	Name string
}

func main() {
	qLock = &sync.RWMutex{}

	r := GetRouter()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})

	//todo (agreen) replace with actual origin list
	//originsOk := handlers.AllowedOrigins([]string{"http://localhost:3000", "http://dev.katamarijr.com", "https://dev.katamarijr.com"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	//listenPort := viper.GetInt(config.ListenPort)
	listenPort := 4444
	log.Printf("Listening on port %v\n", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", listenPort), handlers.CORS(originsOk, headersOk, methodsOk)(r)))
}

// GetRouter returns the route tree.
func GetRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	r.Use(handlers.RecoveryHandler())

	r.HandleFunc("/drain", Get).Methods("GET")
	r.HandleFunc("/", Post).Methods("POST")

	return r
}

// Post will handle a post request containing a name to save in the queue.
func Post(w http.ResponseWriter, r *http.Request) {
	var req request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		return
	}

	qLock.Lock()
	queue = append(queue, req.Name)
	qLock.Unlock()

	w.WriteHeader(200)
}

// Get will handle an get request to server the contents of the queue and reset it.
func Get(w http.ResponseWriter, r *http.Request) {
	qLock.Lock()

	response := struct {
		Names []string
	}{}

	response.Names = queue

	b, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(b)
	if err != nil {
		panic(err)
	}

	//empty the queue
	queue = []string{}

	qLock.Unlock()
}
