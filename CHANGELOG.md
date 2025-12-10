# Game Master's Assistant / Go Utilities
# Release Notes
## Current Version Information
 * This Package Version: 5.32.1                <!-- @@##@@ -->
 * Effective Date: 10-Dec-2025			<!-- @@##@@ -->

## Compatibility
 * GMA Core API Library Version: 6.39.1		<!-- @@##@@ -->
 * GMA Mapper Version: 4.35.2		<!-- @@##@@ -->
 * GMA Mapper Protocol: 422		<!-- @@##@@ -->
 * GMA Mapper File Format: 23		<!-- @@##@@ -->
 * GMA Mapper Preferences File Format: 12 <!-- @@##@@ -->
 * GMA User Preferences File Format: 4 <!-- @@##@@ -->

# Notice
When upgrading an existing server to version 5.31.0 or later, be sure to run `scripts/upgrade-5.31.0` on each database file to update it to the new schema.

This script requires the `sqlite3` command-line tool to be installed.

In addition, if your server didn't have the following updates installed previously, do them as well. Each of these should be applied from oldest to newest, as needed, and **only** the ones needed.  If you don't need to retain any data when setting up a new server, just delete the database file and the server will create a new one when it starts. These scripts are only needed to patch your old database for use with a newer server.

`scripts/upgrade-5.15.0` for servers older than 5.15.0

`scripts/upgrade-5.13.1` for servers older than 5.13.1

## v5.32.1
### Fixed
 * Corrects error in supported protocol number in mapper package.
 * Retracts 5.32.0 which contains this error.

## v5.32.0
### Enhanced
 * Implements server protocol 422.
 * Adds the notion of users logging in with clients but not actually playing any characters in the game.
### Fixed
 * Now logs errors when setting different `SkinSize` attributes in party members in server configuration file `AC` directives.

## v5.31.0
### Enhanced
 * Implements server protocol 421.
 * Adds tracking of client platforms.
 * Adds ability to manage a library of sound files just as the server does for images files, and play them on clients.

## v5.30.0
### Enhanced
 * Implemets server protocol 420.
 * Adds `CharacterName` (`AKA`) server message and corresponding support in other messages.
 * Adds armor class stats to creature health objects.
 * Adds targets and type info to die rolls.
 
## v5.29.0
### Added
 * Implements server protocol 419.
 * Adds support for pinned chat comments.

## v5.28.0
### Enhanced
 * Implements server protocol 418.
 * Supports client hit point requests to the GM (temporary and permanent, and reporting of temporary hit points in health data).

## v5.27.1
### Fixed
 * Added the missing `Replay` feild to the `TO` and `ROLL` messages in the mapper protocol.

## v5.27.0
### Enhanced
 * Implements server protocol 417.
   * New `Origin` field in `ROLL` and `TO` messages informs clients if they are receiving replies to messages or operations they initiated as opposed to copies or notifications of things initiatied by peers (even if by the same user).
   * New `Feature` string `GMA-MARKUP` for `ALLOW` message from client to enable markup formatting codes in chat messages. 
   * Clients can send chat messages via `TO` with the `Markup` flag set. Clients seeing this flag can then know to apply the formatting codes when displaying them. The server will filter out formatting codes for clients which have not indicated compability with this scheme.
   * Protocol includes recommended encoding to represent lookup tables in die-roll presets.
 * Adds support for system-wide global die-roll preset storage and retrieval.
 * Now prohibits usernames which begin with SYS$ prefix (now reserved for internal system identifiers).

## v5.26.0
### Enhanced
 * Implements server protocol 416.
 * Adds support for `FAILED` messages for client requests that are rejected by the GM or have errors not related to privileges.
 * Adds support for `TMRQ` and `TMACK` messages for players to request new timers in the time tracking system.

## v5.25.2
### Fixed
 * Die-roll labels may also include ‖ separators.

## v5.25.1
### Fixed
 * Die-roll label colors are stripped from output to clients which do not declare that they don't allow that feature.

## v5.25.0
### Enhanced
 * Die-roll labels now allow an extended syntax which allows users to add their own custom colors for individual modifiers and permuted rolls.
### Added
 * `LineWrap` function to break long lines up on embedded newlines.
 * `WithoutNewline` option for hex dump output.
 * `YorN` yes/no prompt function.

## v5.24.0
### Enhanced
 * The standalone die-roller command `roll` was refactored and the ability to report statistics of the result set were added. *Note:* This changes the JSON output format (`-json` option).

## v5.23.0
### Added
 * Changes to authenticator and protocol to make the challenge/response exchange a little more secure, such as it is, by decoupling the nonce from the number of hash rounds. It is backward compatible with previous protocol versions but a server which uses protocol 415 and later will insist on a client that understands the new authentication scheme. Deprecates old methods which didn't understand this in favor of new ones, which is part of what makes this not be a breaking change:
    * `GenerateChallenge` replaced with `GenerateChallengeWithIterations`
    * `GenerateChallengeBytes` replaced with `GenerateChallengeBytesWithIterations`
    * `AcceptChallenge` replaced with `AcceptChallengeWithIterations`
    * `AcceptChallengeBytes` replaced with `AcceptChallengeBytesWithIterations`
### Enhanced
 * Text handling code picks up latest changes from GMA core to PostScript preambles.

## v5.22.0
### Added
 * Added new option `|total n` to the die roller syntax. This repeats the die roll until the cumulative total of all the rolls meets or exceeds a target value.
 * Added new CLI tool `markup` (invoked as `gma go markup` if using the whole GMA tool suite) which renders GMA markup text to plain text, PostScript, or HTML.
    * This is still a work in progress. It functions, but it's quite finished yet for PostScript output. The other output formats seem to be about ready.
 * Added text processing functions `CenterPrefix`, `CenterSuffix`, `CenterText`, `CenterTextPadding` to the `text` package.
### Enhanced
 * Moved description of markup and die-roll syntax to constants that can be read from applications.
 * Enhanced the markup language.
 * Changed GMA PostScript preamble data to be exported from the `text` package (was private before).
### Fixed
 * `cmd/roll` reports the default seed in its JSON output if one wasn't given to it via `-seed`.
 * Rearranged manpages so they all begin with `gma-go-` to disambiguate them from commands on the system with the same names. This also reflects the fact that they are intended to be invoked as subcommands of `gma` by starting the command line with `gma go CMDNAME ARGS...` (although they can be run independently too).


## v5.21.1
### Fixed
 * `cmd/roll` was left out of the git commit.

## v5.21.0
### Added
 * Standalone die roller program, suitable for interactive use or to serve back-end code for web interfaces and such.
 * Die-roll syntax help text now maintained in the `dice` package.

## v5.20.1
### Fixed
 * Corrected bug in handling of `DR` server request. It was not unmarshalling the JSON payload, making it impossible to request delegated die-roll presets.

## v5.20.0
### Enhanced
 * Added the `image-audit` program. This checks the server's database for any images it's advertising to clients which don't actually exist on the web server. It can also update the server's database to remove any such invalid entries.

## v5.19.0
### Enhanced
 * Added Quality-of-service (QoS) limits. By default these are not enabled, but by adding a `QOS` entry in the server's `.init` file, one or more of the following limits may be enabled, which allow the server to eject misbehaving clients, helping to prevent denial of service situations where a client can overburden or deadlock the server:
    * Limit on the number of times a client can ask for the same image *after* the server has told it where to find it. This indicates a problem with the client's ability to obtain images which may result in a packet storm as it keeps asking for the same image forever.
    * Limit on the overall volume of packets received in a given span of time.
    * A log entry which shows the QoS metrics gathered so far from each client.
 * Implements protocol 414, which now allows `DENIED` messages to be sent from the server at any time (not just during the authentication phase), for the server to indicate that it now denies access to the client (say, for QoS rule violations).
 * Updates mapper client preferences format to include a new field to set the default timer visibility for the client.

## v5.18.0
### Enhanced
 * Improves functionality of `session-stats` command to use expanded JSON input file and generate the full game synopsis and video link list.

## v5.17.0
### Enhanced
 * Implements protocol 413.
    * Expands fields in the `PROGRESS` message to track game timers as well as operation progress.
    * Adds message definitions for `CORE`, `CORE/`, and `CORE=` messages. Note that the core database is not yet implemented, so while the protocol messages are understood, no actual data can be retrieved using them yet.
 * Die rolls are now evaluated using floating-point instead of integer values. This means that constant values in die-roll expressions may include a fractional part, including formats like `123`, `1.23`, and `.123` but not `12.` (trailing periods aren't recognized unless following digits).
    * The result of binary math operators involving real numbers are truncated to the greatest integer value less than or equal to the result, as per typical d20 game mechanics.
### Fixed
 * Corrects issues with die-roll syntax parsing.
    * `//` inside a `{ ... / ... }` permutation was confused as a separator. It is now correctly seen as a division operator.
    * Labels are now allowed where operators are expected, so something like `(2+2) bonus` will now work as expected.
    * The definition of a label has been restricted since the looser definition had a tendency to swallow invalid expressions as labels. A label must now be a space-separated series of one or more words, where a word begins with a Unicode letter or underscore followed by zero or more Unicode letters, digits, underscores, periods (full stops), and/or commas.

## v5.16.0
### Enhanced
 * Adds `flash_updates` preferences item for mapper configuraiton files.

## v5.15.0
### Enhanced
 * Implements protocol 412.
   * Adds the ability for users to designate authorized delegates to manage their die-roll presets.
 * Adds support for mapper preferences file format version 5.

### Fixed
 * Corrected an error in the client library which lost subscription information when following a `REDIRECT` directive from a server.

## v5.14.0
### Enhanced
 * Implements protocol 411.
   * Adds timestamps to chat messages and die-roll results.

## v5.13.2
### Fixed
 * Now immediately disconnects clients after sending `REDIRECT` to them.

## v5.13.1
### Fixed
 * The move to protocol version 410 introduced an error in how the chat history database was managed. This release includes a script `scripts/upgrade-5.13.1` which repairs the database, as well as new server code to prevent this from happening again.
### Enhanced
 * When the server receives the `USR1` signal to reload its configuration files, it now also jumps the chat/die-roll message IDs to the current UNIX timestamp value, which should put it ahead of other concurrently-running servers (unless you have a server that's been spewing a message per second since it started, which is really unlikely, or your server's clock is wrong).

## v5.13.0
 * Implements protocol 410.
   * Adds `REDIRECT` command to protocol and server init file
   * Adds server-side configuration extension to `WORLD` command to allow server admin/GM to set a limited number of client preferences, overriding local user preferences.
      * `MkdirPath`, `SCPDestination`, `ServerHostname` GM settings for uploading content to the server.
      * `ImageBaseURL` setting which tells clients where to find images and maps on the server.
      * `ModuleCode` setting which specifies the adventure module in play.
   * Adds client code to accept `REDIRECT` and server-side configuration, implemented in `map-console`.
 * Server now interprets the `HUP` signal as a request to hang up on all connected clients but leave the server running and accepting new connections. `INT` remains as the signal to shut down the server itself (previously, `HUP` and `INT` both shut down the server).

## v5.12.0
### Enhancements
 * Adds `Stipple` field to map elements to specify a pattern fill.
 * Moves to protocol 409.

## v5.11.1
### Fixes
 * Corrects die-roll syntax error where spaces between open parentheses was not parsed correctly (`((42))` worked but not `(  (  42  )  )`)

## v5.11.0
### Enhancements
 * Servers can now filter clients to require a minimum client version number that is allowed to connect.
    * This is accomplished by adding `MinimumVersion` and `VersionPattern` fields to each `Package` in a server's init file `UPDATES` section.
    * See the protocol documentation for details on these fields.

## v5.10.0
### Enhancements
 * Adds preferences option to run curl in insecure mode (mapper preferences file v4)

## v5.9.1
### Fixes
 * Doesn't allow `d0` in die rolls. This caused the server to panic.
 * Doesn't allow dividing by 0 in die roll expressions, which also caused the server to panic.

## v5.9.0
### Enhancements
 * Moved protocol to version 408
   * Adds `PolyGM` attribute to `PS`.
   * Adds `ReceivedTime` and `SentTime` to `ECHO`.
   * Deprecates `Size` in favor of expanded and generalized `SkinSize` in `PS`.
 * Added New Relic APM instrumentation to the major server functions.
   * This is just the start of a work in progress.
 * Renamed the server's `-profile` option to `-cpuprofile`
 * Changed the semantics of `-telemetry-log` so that by default it does not log at all; now give `"-"` as the log path to have it log to standard output.
### TODO
 * Add custom attributes to transactions (client info)
 * Add error reporting

## v5.8.3
### Fixes
 * Added missing source files needed for mapper clients and server.

## v5.8.2
### Fixes
 * The server and client connection code included a spin loop that sent CPU usage through the roof when clients were connected. This has been fixed.

## v5.8.1
### Fixes
 * Correction to `coredb` feat import code.

## v5.8.0
### Enhancements
 * Added support for animated image files.

### IMPORTANT UPGRADE NOTE
When moving to version 5.8.0, a change is needed to the database file in use by the server.
You can either delete the database file so that the 5.8.0 server will create a new one, or run the following
command after shutting down your old server to make the necessary schema change before starting your 5.8.0 server:
```
scripts/upgrade5.7.0-5.8.0
```

## v5.7.0
### Enhancements
 * Added `coredb` program and supporting functions and types in the `util` package to import/export entries to/from the core game database (which will be) introducted in GMA Core 7.0.
 * Added the GMA PostScript preamble file as `string` constants `text.commonPostScriptPreamble` and `text.gmaPostScriptPreamble`
 * Added data structures and functions to access the global GMA preferences settings (which will be) introduced in GMA Core 7.0.

## v5.6.0
### Enhancements
 * Updated to file format version 21
 * Removed redundant `Area` field from creatures
 * Added new `CustomReach` field for creatures
 * Added new `DispSize` field for creatures

## v5.5.2
### Fixes
 * `map-console` didn't work if no `preferences.json` or `*.conf` file was found.
 * clients didn't see `Transparency` attribute of status markers.
 * `map-console` didn't list transparency attributes.

## v5.5.1
### Enhancements
 * Implemented protocol 405.
 * `map-console` shows server version number upon connect.

## v5.5.0
### Enhancements
 * Updated `UserPreferences` structure.
 * Improved `Makefile`
 * Implements server protocol 404
   * Adds `Transparent` attribute to status marker definitions
   * Adds `Hidden` attribute to creatures

### Fixed
 * Typos in documentation

## v5.4.0
### Enhancements
 * Added <= and >= operators for die rolls to constrain values to be within defined limits.

## v5.3.1
### Fixes
 * Die roll expressions got confused with spaces between parentheses and operators (e.g., `(d20 + 3) * 2`). Fixes [issue #19](https://github.com/MadScienceZone/go-gma/issues/19)

## v5.3.0
### Enhancements
 * The `dice` module now respects algebraic order of operations and use of parentheses in die-roll expressions.
 * The `map-console` program now reads from mapper preferences settings as does the latest mapper version.
 * Added code to the `util` module to parse mapper preferences files.

## v5.2.2
### Enhancements
 * Implemented protocol 403 which expands the OK message and adds AI/.

### Fixes
 * Corrected a bug where receipts weren't sent when a player make a die roll to GM only with multiple results, such as with a permutation.

## v5.2.1
### Enhancements
 * Added shorter CLI options to `map-console`.
 * Added doc comments for commands.
 * Added `Makefile` to make building command binaries easier.

## v5.2.0
 * Updated to server protocol version 402.
 * Reports "to-GM-only" die rolls with more explicit details for clients to display.

## v5.1.1
 * Changed how peer connections and disconnections are announced (removed race condition).

## v5.1
 * Moved to server protocol version 401.

## v5.0.0
 * Moved to server protocol version 400.
 * Updated file format versions to JSON-based (mapper file format 20, die-roll file format 2).
 * Refactored client/server code.

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
