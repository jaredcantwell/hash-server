# hash-server
Coding assignment that implements an http server that asynchronously hashes passwords.

#Getting Started
To run the server:

```bash
go get github.com/jaredcantwell/hash-server
cd ~/go/src/github.com/jaredcantwell/hash-server
go build *.go
./main --port 8081
```

To run tests:

```bash
go test -v ./server
go test -v ./hasher
```

#Design Overview
##Server
The Server (package server) wraps all the logic around launching the http server, registering handlers, parsing inputs, formating responses, and returning errors.  All hashing logic is in the AsyncHasher (package hasher).

##Hasher


#Notes
###Design Decisions / Assumptions
 - Getting the hash will remove the id from the cache.  This keeps the memory from growing unbounded.  But it means that multiple calls to GET /hash/### will fail after the first one.  This could be adjusted to keep a result for some duration too.
 - The /stats endpoint only returns the average time of the hash computation since this is the most expensive part.
 - The 5 second delay is intended to mimic a hash computation (or other operation) that takes a long time.  I assume these operations are not cancelable, and we must wait for them to complete once they begin.

###Notes
 - First Go app ever.. tried to keep it idiomatic Go (i.e. no mutexes).  In real world, I'd have to learn more to know if these are the best decisions.

###Improvements
 - use net/http/httptest
 - lots and lots more testing - didn't have time for all edge conditions.  Test code contains documentations that I _would_ write with more time.
 - better abstractions to enhance testing - like all async stuff w/o having to run 5 second delay

