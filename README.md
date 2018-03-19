# hash-server
Coding assignment that implements an http server that asynchronously hashes passwords.

## Getting Started
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

## Design Overview
The entire implementation is in 3 files:
 - server/server.go
 - hasher/hasher.go
 - hasher/stats.go

This project implements the following APIs:

Method | Description
-------|------------
POST /hash | Accepts a password parameter and returns an integer id that can be used with the GET method to retrieve the hash of the password at a later time.
GET /hash/{hashId} | Retrieves the hash of a password requested by a previous call to POST /hash. A hash can only be retrieved once.
GET /stats | Gets stats about the total number of hash requests and the average hash processing time.
POST /shutdown | Requests the server to cleanly shutdown.  NOTE: This method will return immediately, but shutdown may take longer to complete if there are many in-flight requests.

### Server
The Server (package server) wraps all the logic around launching the http server, registering handlers, parsing inputs, formating responses, and returning errors.  The Server also handles cleanly shutting down when requested.  All hashing logic is in the AsyncHasher (package hasher).  Ther Server can be run on any port, and an error will be returned if the port is not usable.

### Hasher
The AsyncHasher (package hasher) handles mangement of the async hashing operations.  It coordinates background requests, tracks stats, and can cleanly shutdown when requested.

AsyncHasher is an interface.  There are two concrete implementations:

Class | Description
------|------------
AsyncHasherChannel | Uses channels as the primary means of synchronization for get/set operations on the map of hashes and accessing the stats.
AsyncHasherMutex | Uses mutexes to protect the map of hashes and the stats.  No channels are used.

## Notes
### Design Decisions / Assumptions
 - I am assuming the purpose of the project is to mimic long-running operations and return "async handles" to the user, which they can use to poll for the completion.  Therefore, I decided to not store the hash results in memory indefinitely, which could cause unbounded memory growth.  Instead, getting the hash will remove the id from the cache.  But it means that multiple calls to GET /hash/### will fail after the first one.  This could be adjusted to keep a result for some duration too.
 - I assume that for /stats the user is most interested in the expensive operation of hashing.  Therefore, the /stats endpoint only returns the average time of the hash computation since this is the most expensive part. It is reported in milliseconds per the instructions, however on my machine it typically only takes a couple hundred microseconds to perform the hash.
 - The 5 second delay is intended to mimic a hash computation (or other operation) that takes a long time.  I assume these operations are not cancelable, and we must wait for them to complete once they begin.

### Notes
 - This is my first Go app ever.  I tried to keep it idiomatic Go (i.e. no mutexes).  In the real world, I'd have to learn more to know if these are the best decisions or not.
 - I implemented an AsyncHasherMutex that uses mutexes instead of channels too.  This feels more straightforward, but I'm unsure if its "good" Go code.  If the hasher were more complicated the locking coule get to the point that channels would start to make more sense.

### Improvements
 - Improve documentation for the REST API.  I would love to use something like swagger, but that requires packages outside of the standard library.  Regardless, since this provides an API, that API should be well documented somewhere that is ideally programatically accessbile and documented close to the code.
 - GET /hash should distinguish between not yet complete and already retrieved.
 - Much better testing
   - The net/http/httptest library looks very powerful for doing more in depth API testing.  I did not have time to integrate this into my unit tests.
   - More edge case and stress testing.  Good tests usually take longer to write than the code they're testing.  I didn't have to write all these tests, but I did document the tests that I _would_ write if I did have more time.  Hopefully this can suffice in showing the edge cases that should be tested with more time.
   - The code could make better use of interfaces to faciliate more targeted testing.  For example, it would be nice to test all the async background task logic without having to actually execute operations that take 5 seconds each, which kills test time.  Unfortuantely, I didn't have time to hammer these good testing interfaces out.
