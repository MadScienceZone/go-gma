[![Coverage Status](https://coveralls.io/repos/github/fizban-of-ragnarok/go-gma/badge.svg?branch=main)](https://coveralls.io/github/fizban-of-ragnarok/go-gma?branch=main)
![GitHub](https://img.shields.io/github/license/fizban-of-ragnarok/go-gma)
# go-gma
Golang port of the GMA core API libraries.

This is a work in progress. Only a small portion of the GMA
API core library has been ported to this point.

## GMA?
This is part of a larger project called GMA (Game Master's Assistant)
which is a suite of tools to facilitate the play of table-top fantasy
role-playing games. It provides a GM toolset for planning encounters,
tracking character state, and running encounters in a comprehensive way.
This includes a multi-user interactive tactical battle map where players
can move their tokens around the map, initiative is managed automatically,
etc.

While we intend to open source GMA eventually, it's not quite ready for
general use (mostly because it needs to be generalized more to be playable
on multiple game systems and less tied to the author's game group).
The manual describing the full GMA product may be downloaded 
[here](https://www.madscience.zone/gma/gma.pdf) (PDF, 61Mb).

In the mean time, we're porting one part of the GMA suite (the interactive
map server) to Go for increased stability and performance, which we will 
release as open source ahead of the rest of GMA, since it's independent
of the generalization issues that are gating the release of GMA.

To support this, the `go-gma` repository holds parts of the core GMA API
needed by the map server.


## Documentation
API docs may be viewed at [pkg.go.dev](https://pkg.go.dev/github.com/MadScienceZone/go-gma/v4).

## Versioning
We are keeping the version numbers for this project in sync with
the main GMA project, so version _x_._y_._z_ of `go-gma` will always
be compatible with version _x_._y_._z_ of `gma`.

## Legal Notice
GMA uses trademarks and/or copyrights owned by Paizo Inc., used under Paizo's 
Community Use Policy ([paizo.com/communityuse]()). We are expressly prohibited from 
charging you to use or access this content. GMA is not published, endorsed, or 
specifically approved by Paizo. For more information about Paizo Inc. and Paizo 
products, visit [paizo.com]().
