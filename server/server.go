// Package server implements an http server that hashes passwords.
//
// A basic http server that provides an interface to asynchronously submit
// requests for passwords to be hashed with sha512, and then to retrieve the
// results at a later time using an id.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jaredcantwell/hash-server/hasher"
)

// Server implements the functionality of this package.
type Server struct {
	shutdownChan chan interface{}
	hasher       *hasher.AsyncHasher
	srv          *http.Server
}

// New creates and initializes a new Server that will listen on the supplied
// port and provide the http functionality of this package.  The server will
// not begin listening though.  Call Run to startup the server for incoming
// connections.
func New(port int) *Server {
	var server Server
	server.srv = &http.Server{Addr: fmt.Sprintf(":%d", port)}
	server.shutdownChan = make(chan interface{}, 1)
	server.hasher = hasher.New()
	return &server
}

// Run starts up the underlying http server and begins listening for new connections.
// Run is a blocking call and will not return until a POST /shutdown request is
// made, at which point everything will be cleaned up and Run will return.
func (s *Server) Run() {
	http.HandleFunc("/hash/", mux(s.hashGETHandler, s.hashPOSTHandler))
	http.HandleFunc("/stats", mux(s.statsHandler, nil))
	http.HandleFunc("/shutdown", mux(nil, s.shutdownHandler))

	// Startup the server in the background so that we can perform the shutdown
	// in this routine asynchronously
	listenErr := make(chan error)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Printf("Httpserver: ListenAndServe() error: %s", err)
				listenErr <- err
			}
		}
	}()

	// Wait until the /shutdown handler signals that its been called (at least once)
	// OR an error happened in the startup
	select {
	case <-listenErr:
		// Don't shutdown the server because it never started
	case <-s.shutdownChan:
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		if err := s.srv.Shutdown(ctx); err != nil {
			panic(err) // failure/timeout shutting down the server gracefully
		}
	}

	// Now that we can guarantee no new requests will go into the hasher,
	// let outstanding requests drain so we get a clean shutdown
	s.hasher.Drain()

	fmt.Println("Server shutdown.")
}

// parsePathParamInt attempts to parse out a trailing int64 from the provided
// URL.  To handle error cases, the prefix must be provided in order to catch
// "extra" parts in the path.
//
//   parsePathParamInt("/some/path/123", "/some/path/") -> 123
//
// Ideally, we wouldn't have to parse the path parameters ourselves,
// but the frameworks that handle this for you aren't in the standard libraries
// so they are off limits for the assignment.  Thus, we do it ourselves.
func parsePathParamInt(path string, prefix string) (int64, error) {
	idStr := strings.TrimPrefix(path, prefix)
	return strconv.ParseInt(idStr, 10, 64)
}

// hashGETHandler is invoked on a GET request to retrieve the hash for an id provided in the URL.
func (s *Server) hashGETHandler(w http.ResponseWriter, r *http.Request) {
	// First parse out the id being requested
	id, err := parsePathParamInt(r.URL.Path, "/hash/")
	if err != nil {
		http.Error(w, "Invalid request path.  id is not an integer.", 400)
		return
	}

	hash, err := s.hasher.GetAndRemoveHash(id)
	if err != nil {
		http.Error(w, "Hash not found.", 404)
		return
	}

	fmt.Fprintln(w, hash)
}

// hashPOSTHandler is invoked on a POST request to compute a new password hash
func (s *Server) hashPOSTHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hash" {
		http.Error(w, "Invalid path.", 404)
		return
	}

	password := r.FormValue("password")
	if password == "" {
		http.Error(w, "Invalid password parameter.", 400)
		return
	}

	id := s.hasher.Compute(password)
	fmt.Fprintln(w, id)
}

// statsHandler serves up the json stats requests
func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.hasher.Stats())
}

// shutdownHandler signals for the server to be shutdown when a POST /shutdown request is made.
func (s *Server) shutdownHandler(w http.ResponseWriter, r *http.Request) {
	// When a request to /shutdown is made, we can either return immediately and
	// shutdown in the background, or wait for everything to clean up before
	// returning from the request.  Since cleanup involves shutting down the server,
	// it will be hard to respond after we've shutdown the server, so we simply
	// begin the shutdown process with the /shutdown call, but do not wait.
	// With a lot more coordination this could be improved.

	// This select allows multiple calls to shutdown that will all simply
	// just return.  The first called will add to the channel (of size 1),
	// but if a future caller tries to add when the channel is full, that
	// means someone else called shutdown already, so the default branch
	// will just do nothing.
	select {
	case s.shutdownChan <- nil:
	default:
	}
}

// mux is a simple helper demux out GET and POST functions from the single handler that
// you must register with the http code.  It reduces code duplication and hides annoying
// boiler plate code around checking if a request is a GET/POST/etc. and returning an
// error if that method is not supported.
func mux(get func(http.ResponseWriter, *http.Request),
	post func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			if post == nil {
				http.Error(w, "Invalid request method.", 405)
			}

			post(w, r)
		case "GET":
			if get == nil {
				http.Error(w, "Invalid request method.", 405)
			}

			get(w, r)
		default:
			http.Error(w, "Invalid request method.", 405)
		}
	}
}
