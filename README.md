@UTF-8

[![Go Report Card](https://goreportcard.com/badge/github.com/webability-go/xconfig)](https://goreportcard.com/report/github.com/webability-go/xconfig)
[![GoDoc](https://godoc.org/github.com/webability-go/xconfig?status.png)](https://godoc.org/github.com/webability-go/xconfig)
[![GolangCI](https://golangci.com/badges/github.com/webability-go/xconfig.svg)](https://golangci.com)

XConfig for GO v1
=============================

Package xconfig loads a configuration file similar to a .ini file, but with some important improvements:
- The xconfig recognize bool, int, float and Strings, and also collections of values and hierarchical nested sub-config sets.
- The xconfig is compatible with XDataset to inject and use in templates, database records, etc.
- You can load more than one file in merge mode or replacing mode in the same config object.
- You can set and get parameters at any time.
- You can load the config object based on a file, but also on a string and another object; this way you can serialize, unserialize, transfer by any strem, save or load on database the object, etc.
- You can also save back the config string with all its comments.

Manuals are available on godoc.org [![GoDoc](https://godoc.org/github.com/webability-go/xconfig?status.png)](https://godoc.org/github.com/webability-go/xconfig)


TO DO:
======

- Merge vs load to load more than 1 file (pending)
- implement + and : *
- implement Save
- Add a flag when it's a multiple load to warn a "save"
- Add error control con Get* for type conversions (if the type is different as expected)

Version Changes Control
=======================

v0.3.0 - 2020-02-11
-----------------------
- Documentation enhanced, Configuration Syntax Reference added
- Licence and Tests Added

v0.2.1 - 2020-02-10
-----------------------
- Documentation enhanced
- Bug corrected on String and GoString()

v0.2.0 - 2020-02-10
-----------------------
- Modification to XConfig to meet xcore v1.0.0 (.String and .GoString functions added, .Stringify function removed)

v0.1.0 - 2019-12-06
-----------------------
- Code formated before sending to github (gofmt -s)

v0.0.9 - 2019-07-12
-----------------------
- Incode println removed

v0.0.8 - 2019-06-25
-----------------------
- Added Clone functions to meet xcore.XDatasetDef definition

v0.0.7 - 2019-03-06
-----------------------
- Added functions of Time to get time data.

v0.0.6 - 2019-02-05
-----------------------
- Added conversion between types con Get* functions

v0.0.5 - 2019-01-06
-----------------------
- XConfig object aligned with XDataset with missing set of functions
- XConfig csn now be injected into an XTemplate

V0.0.4 - 2018-12-21
-----------------------
- Added GetString, GetInt, GetBool and GetFloat to directly get the known value type (and do not have to convert types)
- Error corrected on Set (type float added)

V0.0.3 - 2018-12-05
-----------------------
> Branch "late-night" added to github
- Added support for sub-XConfig with entry1.entry2.entry3=xxx

V0.0.2 - 2018-11-27
-----------------------
- Added support for strings starting with "
- Added support for array of booleans
- Added support for float parameters
- Work for a full normal load of a file, suporting also repeated parameters (array of parameters)
- VERSION constant added
- XConfig.Get function implemented for basic parameters (without + or :)

V0.0.1 - 2018-11-14
-----------------------
- First commit, only work to load basic parameters
