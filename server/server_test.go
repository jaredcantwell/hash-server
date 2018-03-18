package server

import (
	"testing"
	"net/http"
	"net/url"
	"time"
	"fmt"
	"bytes"
)

func TestStress(t *testing.T) {
	go New(8080).Run()

	// Sleep 2 seconds to let the server startup.. this obviously needs fixed/improved
	time.Sleep(2 * time.Second)

	responses := make(chan string)

	// note: this can scale much higher, but requires changing ulimit and
	// I didn't want to require that to run the tests
	for i := 0 ; i < 100 ; i++ {
		go func() {
			resp, err := http.PostForm("http://localhost:8080/hash",
				url.Values{"password": {"angryMonkey"}})

			if err != nil {
				t.Fail()
				return
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			responses <- buf.String()
		}()
	}

	time.Sleep(10 * time.Second)


	resp, err := http.Post("http://localhost:8080/shutdown", "", nil)
	fmt.Println(resp)
	fmt.Println(err)


	time.Sleep(60 * time.Second)
}

// bad port input
// lots of shutdowns in a row
// error cases
//  invalid request path
//  hash not found
//  password param permutations (error conditions)
//  invalid methods
//    GET /hash
//    POST /hash/1 - should not work
//    DELETE /hash
//    POST /stats
//    GET /shutdown
//    DELETE /shutdown