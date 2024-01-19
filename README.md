[![Coverage Status](https://coveralls.io/repos/github/MadScienceZone/go-gma/badge.svg?branch=main)](https://coveralls.io/github/MadScienceZone/go-gma?branch=main)
![GitHub](https://img.shields.io/github/license/MadScienceZone/go-gma)
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

While we intend to open source GMA **later in ~~2023~~ 2024**, it's not quite ready for
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
API docs may be viewed at [pkg.go.dev](https://pkg.go.dev/github.com/MadScienceZone/go-gma/v5).

## Building
Running `make` in the top-level directory will build all the program binaries
under the `cmd` directory.

By default, this will build the server without instrumentation to collect runtime performance metrics.
If you wish to compile the server with telemetry instrumentation enabled, run `make telemetry`.

## Versioning
We started out with a desire to keep the version numbers for this project in sync with the main GMA project, so version _x_._y_._z_ of `go-gma` would always be compatible with version _x_._y_._z_ of `gma`. (This is why `go-gma` started at at v4. It moved to v5 when the JSON formats were introduced as a breaking change.)
This is no longer the case, so the version numbers of each don't necessarily match, since each project has encountered breaking changes at different times.

## Game System Neutral

The GMA software and the go-gma client in this repository are intended to be game-system-neutral. They are not written for, nor necessarily intended for use with, the Dungeons & Dragons game from Wizards of the Coast, nor do they rely upon OGL-licensed intellectual property from Wizards of the Coast. Where any game system was in mind for these tools, it was the Pathfinder role-playing game system from Paizo, Inc.

## Legal Notice
GMA uses trademarks and/or copyrights owned by Paizo Inc., used under Paizo's 
Community Use Policy ([paizo.com/communityuse]()). We are expressly prohibited from 
charging you to use or access this content. GMA is not published, endorsed, or 
specifically approved by Paizo. For more information about Paizo Inc. and Paizo 
products, visit [paizo.com]().
