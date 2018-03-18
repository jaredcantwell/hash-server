# hash-server
Coding assignment that implements an http server that asynchronously hashes passwords.

To run the server:

```bash
cd ~/go/github.com/jaredcantwell/hash-server
go build *.go
./main
```

To run tests:

```bash
go test -v ./server
go test -v ./hasher
```

Assumptions:
 - 

Notes:
 - 

Improvements:
 - use net/http/httptest
 - lots and lots more testing - didn't have time for all edge conditions
 - better abstractions to enhance testing - like all async stuff w/o having to run 5 second delay
