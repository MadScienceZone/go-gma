# Changelog
## Current Version Information
 * GMA Core API Library Version: 5.0.0		<!-- @@##@@ -->
 * Supported GMA Mapper Version: 4.0.3		<!-- @@##@@ -->
 * Supported GMA Mapper Protocol: 400		<!-- @@##@@ -->
 * Supported GMA Mapper File Format: 20		<!-- @@##@@ -->
 * Effective Date: 13-Dec-2022			<!-- @@##@@ -->

## v4.7.0
 * Moved to server protocol version 400.
 * Updated file format versions to JSON-based (mapper file format 20, die-roll file format 2).

**Warning: requires Go 1.18 or later**

## v4.4.1 (alpha)
### Enhancements
 * Updated for protocol version 333.
 * Added support for Allow server command.
 * Now supports extended status marker definitions (with description of effects).

### Housekeeping
Synced version number with GMA core.

## v4.3.13 (alpha)
Added subtotals to die roll results.

**Warning: a future release of this code will require Go 1.18.**

## v4.3.12 (alpha)
Added additional core GMA functions and the map-console tool, which gives an
interactive interface from the command line to the map server.

## v4.3.10 (alpha)
Added note pointing to paizo's Community Use Policy and GMA's usage of Pathfinder
game-related information.

## v4.3.9 (alpha)
Since this is still in alpha, I'm taking the liberty to change the call to
dice.StricturedDescribeRoll to correct the weirdness of having the sfOpt
parameter, which should logically (no pun intended) be a bool value rather
than a string, and also to provide a more flexible calling paradigm which
doesn't require sending these values when they aren't needed.

Updated handling of custom bullets in the text.Render function so that common
bullet characters are translated to each output format. That means the plain text
formatter outputs Unicode bullet runes now by default, too, so if you really want
to use `*` (for example) as a bullet you'll need to specify that as a custom
bullet rune explicitly, like:

```go
formattedText, err := text.Render(s, AsPlainText, WithBullets('*', '-'))
```

## v4.3.8 (alpha)
Updated documentation. Lots of cleanup to make `golint` happier.
Added random name generation package `namegen`.
Added ability to get raw random number values from `dice.DieRoller` that use the same
PRG.

## v4.3.7 (alpha)
Moved SaveData methods for MapObjects back to being unepxorted. Users should
only save via the provided high-level SaveObjects function.

## v4.3.6 (alpha)
Updates map file support to version 17.

## v4.3.5-alpha
Adds mapper package with code to represent map objects and load/save them
from a disk file or over the client/server protocol. Also adds code for clients
to connect to a running server, with functions to send individual messages to
the server and a mechanism for the client to be notified via subscribed channels
when incoming server messages arrive.

Adds `ToDeepTclString()` to the tcllist package which converts an arbitrary
slice of values (including nested sub-lists) into a Tcl string in one step.
Otherwise one would need to convert non-string types to strings and then
call `ToTclString()` (repeatedly, in the case of sub-lists).

## v4.3.4 (alpha)
Correction to how auth objects manage byte slices. 

## v4.3.3 (alpha)
Added text processing and utility functions.

## v4.3.2 (alpha)
Cleaned up the module documentation. Un-exported some of the internals of the dice package
that weren't supposed to have been exported.

## v4.3.1 (alpha)
Initial move to its own repository. Implements the `auth`, `dice`, and `tcllist` packages.
