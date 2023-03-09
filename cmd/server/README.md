# GMA Server
This is an example of a server which implements the GMA mapper service using the
routines in the go-gma package. 

This replaces the older Python implementation found in the `MadScienceZone.GMA.Mapper.MapService` module
in the GMA Core package's library. The Go version presented here is a more advanced design which implements
the newer 400-series server protocol

# Build notes
* Adding `-tags instrumentation` to the `go build` command will include New Relic APM instrumentation in the server.
   * ...when that's implemented, which at the moment it is not.
