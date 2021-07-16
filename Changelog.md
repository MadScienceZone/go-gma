# Changelog
## Current Version Information
 * GMA Core API Library Version: 4.3.6		<!-- @@##@@ -->
 * Supported GMA Mapper Version: 3.40.9		<!-- @@##@@ -->
 * Supported GMA Mapper Protocol: 332		<!-- @@##@@ -->
 * Supported GMA Mapper File Format: 17		<!-- @@##@@ -->
 * Effective Date: 16-Jul-2021			<!-- @@##@@ -->

## UNRELEASED 
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
