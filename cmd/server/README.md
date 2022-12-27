# GMA Server
This is an example of a server which implements the GMA mapper service using the
routines in the go-gma package. 

It is currently an alpha preview so is lightly documented but better documentation
is coming.

# Build notes
* Adding `-tags instrumentation` to the `go build` command will include New Relic APM instrumentation in the server.
   * ...when that's implemented, which at the moment it is not.
