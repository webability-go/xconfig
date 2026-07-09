// Copyright Philippe Thomassigny 2004-2026.
// Use of this source code is governed by a MIT licence.
// license that can be found in the LICENSE file.

// Package xconfig provides a thread-safe, ordered configuration manager that
// reads and writes .conf files in a structured key=value format.
//
// It extends the classic INI-file concept with automatic type inference,
// multi-value arrays, nested sub-configurations via dotted key notation, and
// full compatibility with the [xcore.XDatasetDef] interface for use in
// templates, database adapters, and other xcore-aware components.
//
// # Installation
//
//	go get -u github.com/webability-go/xconfig
//
// # Quick start
//
//	xc := xconfig.New()
//
//	if err := xc.LoadFile("app.conf"); err != nil {
//	    log.Fatal(err)
//	}
//
//	host, _ := xc.GetString("host")
//	port, _ := xc.GetInt("port")
//	debug, _ := xc.GetBool("debug")
//
//	xc.Set("build", "1.2.3")
//
//	if err := xc.SaveFile("app.conf"); err != nil {
//	    log.Fatal(err)
//	}
//
// # File format
//
// Configuration files are UTF-8 plain text. Lines beginning with # or ; are
// treated as comments and are preserved verbatim during serialisation. Blank
// lines are also preserved. All other lines must follow one of the three
// assignment forms:
//
//	key=value           # normal assignment
//	key:=value          # force-replace — overwrites even when merging
//	key+=value          # force-add    — appends even when loading
//
// Keys are case-sensitive and must not contain whitespace or the '=' character.
// Leading and trailing whitespace around both the key and the value is trimmed
// automatically. An empty key is silently ignored.
//
// A complete example file:
//
//	# app.conf
//	; semicolons are also comments
//
//	host=db.example.com
//	port=5432
//	debug=false
//
//	# repeated keys become arrays
//	tag=stable
//	tag=v2
//
//	# dotted keys create nested sub-configurations
//	cache.host=redis.example.com
//	cache.port=6379
//	cache.ttl=300
//
//	# per-entry operator overrides
//	host:=override.example.com   # always replaces, even in MergeFile
//	tag+=canary                  # always appends,  even in LoadFile
//
// # Type inference
//
// The parser infers the stored Go type from the raw value string:
//
//	Value in file              Stored Go type   Stored value
//	─────────────────────────  ───────────────  ────────────
//	true  / on  / yes         bool             true
//	false / off / no / none   bool             false
//	42, -7, 1000000           int              (integer)
//	3.14, -0.5, 1.5e-3        float64          (float)
//	anything else             string           (as-is)
//
// Because the parser tries bool → int → float → string in that order, values
// such as "1" become int, and "1.0" become float64. Use the forced-string
// prefix described below if you need to store such values as strings.
//
// # Forced-string prefix
//
// Prefixing a value with a double-quote character (") forces string storage
// regardless of the value content, and the prefix itself is stripped:
//
//	count="123          # stores the string  "123"  (not int 123)
//	flag="true          # stores the string  "true" (not bool true)
//	ratio="1.0          # stores the string  "1.0"  (not float 1.0)
//	empty="             # stores the string  ""
//
// To store a string whose first character is itself a double-quote, use two
// consecutive double-quote characters as the prefix:
//
//	label=""+quoted     # stores the string  `"+quoted`
//
// [Marshal] and [XConfig.SaveFile] restore these prefixes automatically, so a
// save-then-load cycle is always lossless.
//
// # Multi-value parameters (arrays)
//
// Repeating a key on successive lines promotes the stored value from a scalar
// to a typed slice. Values are appended in the order they appear in the file:
//
//	country=MX
//	country=US
//	country=FR
//	# → country  []string{"MX", "US", "FR"}
//
//	port=8080
//	port=8081
//	port=8082
//	# → port  []int{8080, 8081, 8082}
//
//	ratio=0.25
//	ratio=0.75
//	# → ratio  []float64{0.25, 0.75}
//
//	flag=true
//	flag=false
//	# → flag  []bool{true, false}
//
// All values for a given key must be of the same inferred type. Mixing types
// (e.g. an int followed by a bool for the same key) returns an error.
//
// # Nested sub-configurations
//
// A dot in a key name creates a hierarchy. Each segment names a sub-XConfig:
//
//	database.host=db.example.com
//	database.port=5432
//	database.tls=true
//	# → "database" is a *XConfig containing host, port, tls
//
// Nesting can be arbitrarily deep:
//
//	app.cache.redis.host=cache.example.com
//	app.cache.redis.port=6379
//	app.cache.redis.ttl=60
//
// Sub-configurations are accessed via [XConfig.GetConfig]:
//
//	db := xc.GetConfig("database")
//	if db != nil {
//	    host, _ := db.GetString("host")
//	    port, _ := db.GetInt("port")
//	}
//
// Dotted keys can also be set programmatically. [XConfig.Set] and [XConfig.Add]
// accept dotted paths and create the necessary sub-XConfig hierarchy:
//
//	xc.Set("database.host", "db.example.com")
//	xc.Set("database.port", 5432)
//
// # Assignment operators
//
// The three assignment operators control how a value combines with any existing
// entry that already has the same key, both within a single file and across
// multiple Load/Merge calls:
//
//	=   default — behaviour depends on the Load* or Merge* function used:
//	              Load* replaces existing entries; Merge* appends to them.
//
//	:=  force-replace — always overwrites the current value, regardless of
//	                    whether the caller used a Load* or Merge* function.
//	                    When := appears later in the same file than a plain =
//	                    for the same key, it replaces whatever was built up
//	                    before it in that file.
//
//	+=  force-add — always appends to the current value, regardless of
//	                whether the caller used a Load* or Merge* function.
//	                Useful for unconditionally extending an array defined
//	                in a base file.
//
// Examples of operator interactions:
//
//	# defaults.conf
//	host=primary.example.com
//	tag=stable
//
//	# local.conf  (loaded with MergeFile)
//	host:=override.example.com   # replaces host even in merge mode
//	tag+=canary                  # appends canary even in merge mode
//	# result: host="override.example.com", tag=[]string{"stable","canary"}
//
// # Loading and merging
//
// XConfig provides three families of source loaders, each with a Load variant
// (replaces existing entries) and a Merge variant (appends to existing entries):
//
//	Source      Load           Merge
//	──────────  ─────────────  ──────────────
//	File        LoadFile       MergeFile
//	String      LoadString     MergeString
//	*XConfig    LoadXConfig    MergeXConfig
//
// A common pattern for layered configuration (defaults, environment, local
// overrides):
//
//	xc := xconfig.New()
//	xc.LoadFile("config/defaults.conf")          // base values
//	xc.MergeFile("config/production.conf")       // environment overrides
//	xc.MergeFile("config/local.conf")            // developer-local tweaks
//	// Missing files are silently ignored if you wrap with os.IsNotExist.
//
// When a Merge* call is made into a non-empty XConfig, the Multiple field is
// set to true, indicating that the data originates from more than one source.
//
// # Getting values
//
// The raw value stored for a key is retrieved with [XConfig.Get], which returns
// (interface{}, bool). The concrete type of the interface is one of: string,
// int, float64, bool, []string, []int, []float64, []bool, time.Time,
// []time.Time, or *XConfig.
//
// Type-specific getters perform automatic cross-type conversions:
//
//	Getter              Returns     Converts from
//	──────────────────  ──────────  ───────────────────────────────────
//	GetString(key)      string      any type via fmt.Sprint
//	GetInt(key)         int         float64 (truncated), bool (0 or 1)
//	GetFloat(key)       float64     int (widened), bool (0.0 or 1.0)
//	GetBool(key)        bool        int (!=0), float64 (!=0)
//	GetTime(key)        time.Time   no conversion; must be exact type
//	GetConfig(key)      *XConfig    no conversion; must be exact type
//	GetDataset(key)     XDatasetDef same as GetConfig
//
// Collection getters return typed slices and wrap single scalars in a
// one-element slice, so callers can always range over the result:
//
//	GetStringCollection(key)  →  []string
//	GetIntCollection(key)     →  []int
//	GetFloatCollection(key)   →  []float64
//	GetBoolCollection(key)    →  []bool
//	GetTimeCollection(key)    →  []time.Time
//
// All getters return (zero-value, false) when the key does not exist.
//
// # Setting and deleting values
//
// [XConfig.Set] creates or replaces the entry for a key. It accepts string,
// int, float64, and bool values, and supports dotted paths:
//
//	xc.Set("host", "localhost")
//	xc.Set("port", 5432)
//	xc.Set("tls", true)
//	xc.Set("database.host", "db.example.com")   // dotted path
//
// [XConfig.Add] appends a value to a key, building a slice on the first
// repeated call. It accepts the same types as Set:
//
//	xc.Add("tag", "stable")
//	xc.Add("tag", "v2")
//	// tag is now []string{"stable", "v2"}
//
// [XConfig.Del] removes a key entirely, including from the insertion order
// so that subsequent serialisation does not include it:
//
//	xc.Del("temp")
//
// # Serialisation
//
// [XConfig.Marshal] returns the full configuration as a string in the .conf
// file format. It preserves comments, blank lines, and insertion order.
// String values that would be misinterpreted on reload are automatically
// protected with the " prefix, making the round-trip lossless:
//
//	xc.Set("label", "true")          // stored as string
//	text := xc.Marshal()             // produces:  label="true
//	xc2 := xconfig.New()
//	xc2.LoadString(text)
//	v, _ := xc2.GetString("label")   // "true"  ← string, not bool
//
// [XConfig.SaveFile] writes the same output to a file with mode 0644.
//
// [XConfig.String] returns a compact, human-readable debug dump (not suitable
// for reloading); its format is XConfig[\nkey:value\n...\n].
//
// # Cloning
//
// [XConfig.Clone] returns a fully independent deep copy of the XConfig. All
// nested sub-configurations and all array slices are recursively cloned, so
// modifications to the clone do not affect the original and vice versa:
//
//	snapshot := xc.Clone().(*xconfig.XConfig)
//	snapshot.Set("env", "test")
//	// xc is unchanged
//
// # Thread safety
//
// All public methods are safe for concurrent use from multiple goroutines.
// Read operations (Get*, Clone, Marshal, String) acquire a shared read lock.
// Write operations (Load*, Merge*, Set, Add, Del, SaveFile) acquire an
// exclusive write lock.
//
// The internal mutex is never held while sub-XConfig values are accessed
// through the pointer returned by [XConfig.GetConfig]. If callers share a
// sub-XConfig pointer across goroutines, they must provide their own
// synchronisation for those accesses.
//
// # Path safety
//
// [XConfig.LoadFile], [XConfig.MergeFile], and [XConfig.SaveFile] reject
// filenames that contain path-traversal sequences (..), returning an error
// rather than reading or writing a file outside the intended directory tree.
package xconfig

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/webability-go/xcore/v2"
)

// VERSION is the semantic version of the XConfig library.
const VERSION = "0.5.0"

// Parameter is an internal storage unit for a single configuration entry.
//
// It pairs a typed value with the type code inferred by the parser and the
// assignment operator recorded from the source file. The type code determines
// how subsequent Add or Merge operations combine this entry with new values.
//
// Type codes:
//
//	0  = unset (zero value; not yet assigned)
//	1  = string       11 = []string
//	2  = int          12 = []int
//	3  = float64      13 = []float64
//	4  = bool         14 = []bool
//	21 = *XConfig (nested sub-configuration)
//
// Assignment codes:
//
//	0  = default (follows the Load*/Merge* call semantics)
//	1  = force-replace (:= operator — always overwrites)
//	2  = force-add     (+= operator — always appends)
type Parameter struct {
	paramtype  int
	Value      interface{}
	assignment int
}

func newParam() *Parameter {
	return &Parameter{0, nil, 0}
}

func (p *Parameter) set(paramtype int, value interface{}, assignment int) {
	p.paramtype = paramtype
	p.Value = value
	p.assignment = assignment
}

// add merges a new value into the parameter, promoting scalars to slices when
// the same key appears more than once.
//
// When a scalar (type 1–4) receives a second value of the same scalar type,
// both values are combined into a two-element slice (type 11–14), with the
// existing value first so that insertion order is preserved.
//
// When a scalar receives a slice of the same element type, the existing scalar
// is prepended to the incoming slice.
//
// When a slice (type 11–14) receives a compatible value, the new element or
// elements are appended to the end of the slice.
//
// Passing incompatible types (e.g. adding a bool to an int parameter) returns
// a descriptive error; no mutation occurs.
//
// Sub-XConfig entries (type 21) are recursively merged rather than replaced.
func (p *Parameter) add(paramtype int, value interface{}, assignment int) error {
	switch p.paramtype {
	case 0: // not set yet — first assignment
		p.paramtype = paramtype
		p.Value = value
	case 1: // string
		if paramtype == 1 {
			sub := make([]string, 0, 2)
			p.Value = append(sub, p.Value.(string), value.(string))
			p.paramtype = 11
		} else if paramtype == 11 {
			// Existing scalar goes first to maintain chronological order.
			p.Value = append([]string{p.Value.(string)}, value.([]string)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to a string")
		}
	case 2: // integer
		if paramtype == 2 {
			sub := make([]int, 0, 2)
			p.Value = append(sub, p.Value.(int), value.(int))
			p.paramtype = 12
		} else if paramtype == 12 {
			p.Value = append([]int{p.Value.(int)}, value.([]int)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to an integer")
		}
	case 3: // float64
		if paramtype == 3 {
			sub := make([]float64, 0, 2)
			p.Value = append(sub, p.Value.(float64), value.(float64))
			p.paramtype = 13
		} else if paramtype == 13 {
			p.Value = append([]float64{p.Value.(float64)}, value.([]float64)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to a float")
		}
	case 4: // boolean
		if paramtype == 4 {
			sub := make([]bool, 0, 2)
			p.Value = append(sub, p.Value.(bool), value.(bool))
			p.paramtype = 14
		} else if paramtype == 14 {
			p.Value = append([]bool{p.Value.(bool)}, value.([]bool)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to a boolean")
		}
	case 11: // array of string
		if paramtype == 1 {
			p.Value = append(p.Value.([]string), value.(string))
		} else if paramtype == 11 {
			p.Value = append(p.Value.([]string), value.([]string)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to an array of strings")
		}
	case 12: // array of int
		if paramtype == 2 {
			p.Value = append(p.Value.([]int), value.(int))
		} else if paramtype == 12 {
			p.Value = append(p.Value.([]int), value.([]int)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to an array of integers")
		}
	case 13: // array of float64
		if paramtype == 3 {
			p.Value = append(p.Value.([]float64), value.(float64))
		} else if paramtype == 13 {
			p.Value = append(p.Value.([]float64), value.([]float64)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to an array of floats")
		}
	case 14: // array of bool
		if paramtype == 4 {
			p.Value = append(p.Value.([]bool), value.(bool))
		} else if paramtype == 14 {
			p.Value = append(p.Value.([]bool), value.([]bool)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to an array of booleans")
		}
	case 21: // nested sub-XConfig — merge recursively rather than replacing
		if paramtype == 21 {
			existingSub := p.Value.(*XConfig)
			newSub := value.(*XConfig)
			existingSub.parsemap(newSub, true)
		}
	default:
		return errors.New("Unknown parameter type")
	}
	return nil
}

// Clone returns a fully independent deep copy of the Parameter.
//
// Slice backing arrays ([]string, []int, []float64, []bool) are copied into
// freshly allocated slices so that the original and the clone never share
// underlying storage. Nested *XConfig values are cloned recursively via their
// own Clone method.
func (p *Parameter) Clone() *Parameter {
	cloned := newParam()
	var clonedval interface{}

	switch v := p.Value.(type) {
	case []string:
		dst := make([]string, len(v))
		copy(dst, v)
		clonedval = dst
	case []int:
		dst := make([]int, len(v))
		copy(dst, v)
		clonedval = dst
	case []float64:
		dst := make([]float64, len(v))
		copy(dst, v)
		clonedval = dst
	case []bool:
		dst := make([]bool, len(v))
		copy(dst, v)
		clonedval = dst
	default:
		clonedval = p.Value
		if cloneable, ok := clonedval.(interface{ Clone() xcore.XDatasetDef }); ok {
			clonedval = cloneable.Clone()
		}
	}
	cloned.set(p.paramtype, clonedval, p.assignment)
	return cloned
}

// XConfigDef is the interface satisfied by XConfig. It extends
// xcore.XDatasetDef with no additional methods, providing a named type alias
// that callers can use to accept any XConfig-compatible value.
type XConfigDef interface {
	xcore.XDatasetDef
}

// XConfig is a thread-safe, ordered key-value configuration store.
//
// Parameters maps each key to its typed Parameter value.
// Comments maps sentinel keys of the form "#<n>" to the corresponding comment
// or blank line text, preserving them across serialisation round-trips.
// Order records the insertion sequence of both parameters and comments, which
// is reproduced faithfully by Marshal and SaveFile.
// Multiple is set to true the first time a second source is merged into this
// config, indicating that the data originates from more than one file or string.
//
// An XConfig must be created via New; a zero value is not ready for use.
type XConfig struct {
	Parameters map[string]Parameter
	Comments   map[string]string
	Order      []string
	Multiple   bool
	mutex      sync.RWMutex
}

// New creates and returns an empty XConfig ready for use.
//
//	xc := xconfig.New()
//	xc.Set("host", "localhost")
//	xc.Set("port", 5432)
func New() *XConfig {
	return &XConfig{
		Parameters: make(map[string]Parameter),
		Comments:   make(map[string]string),
		Multiple:   false,
	}
}

// ── Private helpers ───────────────────────────────────────────────────────────

// validatePath rejects filenames that contain path-traversal sequences (..),
// preventing callers from accidentally reading or writing files outside the
// intended directory tree. An empty filename is accepted and treated as a
// no-op by the callers that invoke this helper.
func validatePath(filename string) error {
	if filename == "" {
		return nil
	}
	cleaned := filepath.Clean(filename)
	for _, part := range strings.Split(filepath.ToSlash(cleaned), "/") {
		if part == ".." {
			return fmt.Errorf("xconfig: path traversal detected in filename %q", filename)
		}
	}
	return nil
}

// isValidKeyPart reports whether s is a non-empty string that does not contain
// the '=' character. It is used to distinguish a := or += operator that
// appears in the key position from a literal occurrence of those character
// sequences inside a value string (e.g. a URL containing ":=").
func isValidKeyPart(s string) bool {
	return len(s) > 0 && !strings.Contains(s, "=")
}

// needsQuotePrefix reports whether value would be misinterpreted on reload —
// that is, whether the parser would infer a type other than string for it.
// When this function returns true, Marshal and SaveFile prepend a '"' character
// so that a subsequent LoadFile or LoadString call restores the original string
// type without data loss.
//
// The following values require protection: the boolean keywords (yes/true/on,
// no/none/false/off), bare integers, bare floats, and strings that already
// begin with a double-quote character.
func needsQuotePrefix(s string) bool {
	lower := strings.ToLower(s)
	switch lower {
	case "yes", "true", "on", "no", "none", "false", "off":
		return true
	}
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	// A string whose first character is '"' would lose that character if it
	// were written unprotected, because the parser strips a leading '"'.
	if len(s) > 0 && s[0] == '"' {
		return true
	}
	return false
}

func (c *XConfig) addcomment(line int, comment string) error {
	id := "#" + strconv.Itoa(line)
	c.Comments[id] = comment
	c.Order = append(c.Order, id)
	return nil
}

func (c *XConfig) addparam(line int, key string, typeparam int, value interface{}, assignment int) error {
	pospoint := strings.Index(key, ".")
	if pospoint >= 0 {
		firstkey := strings.TrimSpace(key[:pospoint])
		subkey := strings.TrimSpace(key[pospoint+1:])

		if val, ok := c.Parameters[firstkey]; ok {
			val.Value.(*XConfig).addparam(line, subkey, typeparam, value, assignment)
		} else {
			p := newParam()
			if err := p.add(21, New(), assignment); err != nil {
				return err
			}
			p.Value.(*XConfig).addparam(line, subkey, typeparam, value, assignment)
			c.Parameters[firstkey] = *p
			c.Order = append(c.Order, firstkey)
		}
		return nil
	}

	if val, ok := c.Parameters[key]; ok {
		p := newParam()
		if err := p.add(val.paramtype, val.Value, assignment); err != nil {
			return err
		}
		if err := p.add(typeparam, value, assignment); err != nil {
			return err
		}
		p.assignment = assignment // preserve operator code for parsemap routing
		c.Parameters[key] = *p
	} else {
		p := newParam()
		if err := p.add(typeparam, value, assignment); err != nil {
			return err
		}
		p.assignment = assignment // preserve operator code for parsemap routing
		c.Parameters[key] = *p
		c.Order = append(c.Order, key)
	}
	return nil
}

// setparam replaces (or creates) the entry for key with the given typed value.
// When key contains a dot, the first segment names a sub-XConfig and the
// remainder is resolved recursively, creating intermediate sub-configs as
// needed. This allows callers such as Set("db.host", "localhost") to build the
// correct nested hierarchy instead of creating a literal "db.host" flat entry.
func (c *XConfig) setparam(line int, key string, typeparam int, value interface{}, assignment int) error {
	pospoint := strings.Index(key, ".")
	if pospoint >= 0 {
		firstkey := strings.TrimSpace(key[:pospoint])
		subkey := strings.TrimSpace(key[pospoint+1:])

		if val, ok := c.Parameters[firstkey]; ok {
			return val.Value.(*XConfig).setparam(line, subkey, typeparam, value, assignment)
		}
		p := newParam()
		if err := p.add(21, New(), 0); err != nil {
			return err
		}
		if err := p.Value.(*XConfig).setparam(line, subkey, typeparam, value, assignment); err != nil {
			return err
		}
		c.Parameters[firstkey] = *p
		c.Order = append(c.Order, firstkey)
		return nil
	}

	p := newParam()
	if err := p.add(typeparam, value, assignment); err != nil {
		return err
	}
	p.assignment = assignment // preserve operator code for parsemap routing
	if _, ok := c.Parameters[key]; !ok {
		c.Order = append(c.Order, key)
	}
	c.Parameters[key] = *p
	return nil
}

// parseline parses a single line from a configuration file or string into
// tempConfig, applying type inference and operator detection.
//
// The := and += operators are detected before the plain = sign is searched for,
// so that key names are never corrupted. For example, "host:=db.example.com"
// correctly yields key="host", assignment=1 (force-replace), value="db.example.com".
// A sequence such as "url=http://host:=path" has the operator inside the value
// portion and is treated as a normal string assignment.
func (c *XConfig) parseline(line int, data string, merge bool) error {
	// Empty lines and comment lines are stored verbatim in the Comments map.
	if len(data) == 0 || data[0] == '#' || data[0] == ';' {
		return c.addcomment(line, data)
	}

	// Detect := and += before searching for plain =, so the operator
	// characters are not included in the key or value.
	assignment := 0
	parsedata := data // operate on a copy; keep the original for comment fallback

	if idx := strings.Index(parsedata, ":="); idx > 0 {
		potKey := strings.TrimSpace(parsedata[:idx])
		if isValidKeyPart(potKey) {
			assignment = 1                                  // force replace
			parsedata = parsedata[:idx] + parsedata[idx+1:] // remove ':', keep '='
		}
	} else if idx := strings.Index(parsedata, "+="); idx > 0 {
		potKey := strings.TrimSpace(parsedata[:idx])
		if isValidKeyPart(potKey) {
			assignment = 2                                  // force add
			parsedata = parsedata[:idx] + parsedata[idx+1:] // remove '+', keep '='
		}
	}

	posequal := strings.Index(parsedata, "=")
	if posequal < 0 {
		return c.addcomment(line, data) // no '=' found → treat as comment
	}

	key := strings.TrimSpace(parsedata[:posequal])
	if len(key) == 0 {
		return c.addcomment(line, data)
	}

	var value interface{}
	var typeparam = 1
	if len(parsedata) > posequal {
		strvalue := strings.TrimSpace(parsedata[posequal+1:])
		value = strvalue
		if len(strvalue) > 0 && strvalue[0] == '"' {
			// Leading '"' forces string interpretation; strip the prefix.
			value = strvalue[1:]
		} else {
			if strvalue == "yes" || strvalue == "true" || strvalue == "on" {
				value = true
				typeparam = 4
			} else if strvalue == "no" || strvalue == "none" || strvalue == "false" || strvalue == "off" {
				value = false
				typeparam = 4
			} else {
				if intvalue, err := strconv.Atoi(strvalue); err == nil {
					value = intvalue
					typeparam = 2
				} else if floatvalue, err := strconv.ParseFloat(strvalue, 64); err == nil {
					value = floatvalue
					typeparam = 3
				}
			}
		}
	}
	// := (force-replace) must call setparam so that a later entry in the same
	// file/string overwrites any earlier value accumulated in tempConfig.
	// += (force-add) and plain = both accumulate via addparam.
	if assignment == 1 {
		return c.setparam(line, key, typeparam, value, assignment)
	}
	return c.addparam(line, key, typeparam, value, assignment)
}

// parsemap integrates all entries from data into c.
//
// When c is empty (first load), every Parameter is deep-cloned via
// Parameter.Clone before insertion, preventing the original and the new config
// from sharing slice backing arrays or sub-XConfig pointers.
//
// When c already contains entries (subsequent load or merge), each parameter
// from data is routed through either setparam (replace semantics) or addparam
// (append semantics). The routing respects per-entry assignment operators (:=
// and +=) that were recorded during parsing, overriding the global merge flag
// where necessary.
func (c *XConfig) parsemap(data *XConfig, merge bool) error {
	if len(c.Parameters) == 0 {
		// Deep-clone every parameter to avoid aliasing.
		c.Parameters = make(map[string]Parameter, len(data.Parameters))
		for k, v := range data.Parameters {
			c.Parameters[k] = *(&v).Clone()
		}
		c.Comments = make(map[string]string, len(data.Comments))
		for k, v := range data.Comments {
			c.Comments[k] = v
		}
		c.Order = make([]string, len(data.Order))
		copy(c.Order, data.Order)
	} else {
		line := len(c.Order)
		for p, v := range data.Parameters {
			// Per-entry operators override the global merge flag.
			switch v.assignment {
			case 1: // := — force replace regardless of merge mode
				c.setparam(line, p, v.paramtype, v.Value, v.assignment)
			case 2: // += — force add regardless of merge mode
				c.addparam(line, p, v.paramtype, v.Value, v.assignment)
			default:
				if merge {
					c.addparam(line, p, v.paramtype, v.Value, v.assignment)
				} else {
					c.setparam(line, p, v.paramtype, v.Value, v.assignment)
				}
			}
		}
		c.Multiple = true
	}
	return nil
}

// loadandparse opens filename, reads it line by line, and integrates its
// entries into c according to merge. Path traversal is rejected before the
// file is opened.
func (c *XConfig) loadandparse(filename string, merge bool) error {
	if len(filename) == 0 {
		return nil
	}
	if err := validatePath(filename); err != nil {
		return err
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	tempConfig := New()
	scanner := bufio.NewScanner(file)
	ln := 1
	for scanner.Scan() {
		line := scanner.Text()
		if err := tempConfig.parseline(ln, line, merge); err != nil {
			return err
		}
		ln++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	c.parsemap(tempConfig, merge)
	return nil
}

func (c *XConfig) parsestring(data string, merge bool) error {
	if len(data) == 0 {
		return nil
	}
	tempConfig := New()
	scanner := bufio.NewScanner(strings.NewReader(data))
	line := 1
	for scanner.Scan() {
		row := scanner.Text()
		if err := tempConfig.parseline(line, row, merge); err != nil {
			return err
		}
		line++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	c.parsemap(tempConfig, merge)
	return nil
}

// ── String representation ─────────────────────────────────────────────────────

// String returns a human-readable ordered summary of the XConfig contents.
// Comments and blank lines are included verbatim; parameter entries are
// rendered as "key:value". This output is intended for debugging and logging;
// use Marshal or SaveFile to produce a configuration text that can be reloaded.
//
//	xc := xconfig.New()
//	xc.Set("host", "localhost")
//	xc.Set("port", 5432)
//	fmt.Print(xc)
//	// Output:
//	// XConfig[
//	// host:localhost
//	// port:5432
//	// ]
func (c *XConfig) String() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	sdata := []string{}
	for _, val := range c.Order {
		if len(val) == 0 {
			continue // guard against empty sentinel keys
		}
		if val[0] == '#' {
			sdata = append(sdata, c.Comments[val])
		} else {
			sdata = append(sdata, val+":"+fmt.Sprint(c.Parameters[val].Value))
		}
	}
	return "XConfig[\n" + strings.Join(sdata, "\n") + "\n]\n"
}

// GoString returns the same content as String, prefixed with "#". It is
// invoked by the %#v formatting verb and by Go's default struct printer.
func (c *XConfig) GoString() string {
	return "#" + c.String()
}

// ── Public API — write operations ────────────────────────────────────────────

// Set creates or replaces the entry for key with value.
//
// Accepted value types are string, int, float64, and bool. Values of any
// other type are stored with type code 0 and serialised via fmt.Sprint.
//
// Dotted key paths create or update nested sub-configurations:
//
//	xc.Set("host", "db.example.com")
//	xc.Set("port", 5432)
//	xc.Set("tls", true)
//	xc.Set("timeout", 30.5)
//
//	// Dotted path — creates or updates a "database" sub-config:
//	xc.Set("database.host", "db.example.com")
//	xc.Set("database.port", 5432)
//
//	db := xc.GetConfig("database")
//	// db.GetString("host") → "db.example.com"
//	// db.GetInt("port")    → 5432
//
// Set is a no-op when key is the empty string.
func (c *XConfig) Set(key string, value interface{}) {
	if len(key) == 0 {
		return // empty keys are silently ignored
	}
	var valuetype int
	switch value.(type) {
	case string:
		valuetype = 1
	case int:
		valuetype = 2
	case float64:
		valuetype = 3
	case bool:
		valuetype = 4
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.setparam(0, key, valuetype, value, 1)
}

// Add appends value to the entry for key, building a typed slice when the key
// already holds a scalar value of the same type.
//
// Accepted value types are string, int, float64, and bool. Passing a value of
// any other type returns an error. Appending a value whose type is incompatible
// with the existing entry (e.g. adding a bool to an int key) also returns an
// error; the configuration is not modified in that case.
//
//	xc.Add("country", "MX")
//	xc.Add("country", "US")
//	xc.Add("country", "FR")
//	countries, _ := xc.GetStringCollection("country")
//	// countries == []string{"MX", "US", "FR"}
//
//	xc.Add("port", 8080)
//	xc.Add("port", 8081)
//	ports, _ := xc.GetIntCollection("port")
//	// ports == []int{8080, 8081}
//
// Add returns a non-nil error when key is the empty string.
func (c *XConfig) Add(key string, value interface{}) error {
	if len(key) == 0 {
		return errors.New("xconfig: key must not be empty")
	}
	var valuetype int
	switch value.(type) {
	case string:
		valuetype = 1
	case int:
		valuetype = 2
	case float64:
		valuetype = 3
	case bool:
		valuetype = 4
	default:
		return errors.New("The XConfig.Add function only accepts string, integer, float64 and boolean values")
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.addparam(0, key, valuetype, value, 0)
}

// Del removes the entry for key from the configuration. If the key does not
// exist, Del is a no-op. The key is removed from Parameters, Comments, and
// Order, so subsequent calls to Marshal, SaveFile, String, and iteration via
// Order will not include it.
//
//	xc.Set("temp", "scratch")
//	xc.Del("temp")
//	_, exists := xc.Get("temp")  // exists == false
func (c *XConfig) Del(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.Parameters, key)
	delete(c.Comments, key)

	// Build a new slice instead of modifying Order in-place during iteration.
	newOrder := c.Order[:0]
	for _, v := range c.Order {
		if v != key {
			newOrder = append(newOrder, v)
		}
	}
	c.Order = newOrder
}

// ── Public API — read operations ─────────────────────────────────────────────

// Get returns the raw value stored under key and a boolean indicating whether
// the key was found. The concrete type of the returned interface{} is one of:
// string, int, float64, bool, []string, []int, []float64, []bool, or *XConfig.
//
// Returns (nil, false) when the key does not exist.
//
//	if v, ok := xc.Get("port"); ok {
//	    fmt.Println(v)        // e.g. 5432 (int)
//	    fmt.Printf("%T\n", v) // "int"
//	}
func (c *XConfig) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		return val.Value, true
	}
	return nil, false
}

// GetDataset returns the value of key as an xcore.XDatasetDef. Only nested
// sub-XConfig entries satisfy this interface. Returns (nil, false) when the
// key does not exist or holds a scalar value.
//
//	// Config file:
//	//   database.host=localhost
//	//   database.port=5432
//
//	if ds, ok := xc.GetDataset("database"); ok {
//	    // ds implements xcore.XDatasetDef and can be passed to xcore templates.
//	    host, _ := ds.GetString("host")  // "localhost"
//	}
func (c *XConfig) GetDataset(key string) (xcore.XDatasetDef, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		if xc, ok := val.Value.(*XConfig); ok {
			return xc, true
		}
	}
	return nil, false
}

// GetCollection returns the value of key as an xcore.XDatasetCollectionDef.
// Returns (nil, false) when the key does not exist or its value does not
// implement that interface.
func (c *XConfig) GetCollection(key string) (xcore.XDatasetCollectionDef, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		if col, ok := val.Value.(xcore.XDatasetCollectionDef); ok {
			return col, true
		}
	}
	return nil, false
}

// GetString returns the value of key as a string. If the stored value is not
// already a string, it is converted with fmt.Sprint. Returns ("", false) when
// the key does not exist.
//
//	xc.Set("version", "2.3.1")
//	ver, ok := xc.GetString("version")   // "2.3.1", true
//
//	xc.Set("port", 8080)
//	port, ok := xc.GetString("port")     // "8080", true (int → string)
func (c *XConfig) GetString(key string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		switch v := val.Value.(type) {
		case string:
			return v, true
		default:
			return fmt.Sprint(v), true
		}
	}
	return "", false
}

// GetInt returns the value of key as an int. Float64 values are truncated
// toward zero; bool values are converted to 1 (true) or 0 (false). Returns
// (0, false) when the key does not exist or the value cannot be converted.
//
//	xc.Set("workers", 4)
//	n, ok := xc.GetInt("workers")   // 4, true
//
//	xc.Set("ratio", 1.9)
//	r, ok := xc.GetInt("ratio")     // 1, true (float64 truncated toward zero)
//
//	xc.Set("enabled", true)
//	e, ok := xc.GetInt("enabled")   // 1, true (bool → int)
func (c *XConfig) GetInt(key string) (int, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		switch v := val.Value.(type) {
		case int:
			return v, true
		case float64:
			return int(v), true
		case bool:
			if v {
				return 1, true
			}
			return 0, true
		}
	}
	return 0, false
}

// GetFloat returns the value of key as a float64. Int values are widened
// without loss; bool values are converted to 1.0 (true) or 0.0 (false).
// Returns (0, false) when the key does not exist or the value cannot be
// converted.
//
//	xc.Set("ratio", 3.14)
//	r, ok := xc.GetFloat("ratio")   // 3.14, true
//
//	xc.Set("count", 7)
//	n, ok := xc.GetFloat("count")   // 7.0,  true (int widened)
func (c *XConfig) GetFloat(key string) (float64, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		switch v := val.Value.(type) {
		case float64:
			return v, true
		case int:
			return float64(v), true
		case bool:
			if v {
				return 1.0, true
			}
			return 0.0, true
		}
	}
	return 0, false
}

// GetBool returns the value of key as a bool. Non-zero int and float64 values
// are true; zero values are false. Returns (false, false) when the key does
// not exist or the value cannot be converted.
//
//	xc.Set("debug", true)
//	d, ok := xc.GetBool("debug")    // true,  true
//
//	xc.Set("retries", 0)
//	r, ok := xc.GetBool("retries")  // false, true (zero int → false)
//
//	xc.Set("factor", 2.5)
//	f, ok := xc.GetBool("factor")   // true,  true (non-zero float → true)
func (c *XConfig) GetBool(key string) (bool, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		switch v := val.Value.(type) {
		case bool:
			return v, true
		case int:
			return v != 0, true
		case float64:
			return v != 0, true
		}
	}
	return false, false
}

// GetTime returns the value of key as a time.Time. Returns (time.Time{}, false)
// when the key does not exist or its value is not a time.Time.
//
// Note: the parser does not parse dates from configuration files; time.Time
// values can only be stored programmatically via Set.
//
//	deadline := time.Now().Add(24 * time.Hour)
//	xc.Set("deadline", deadline)
//	t, ok := xc.GetTime("deadline")  // deadline, true
func (c *XConfig) GetTime(key string) (time.Time, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		if t, ok := val.Value.(time.Time); ok {
			return t, true
		}
	}
	return time.Time{}, false
}

// GetStringCollection returns the value of key as a []string slice. A single
// string scalar is wrapped in a one-element slice so that callers can always
// range over the result without a type switch. Returns (nil, false) when the
// key does not exist or holds an incompatible type.
//
//	// Config file:
//	//   tag=stable
//	//   tag=canary
//	tags, ok := xc.GetStringCollection("tag")
//	// tags == []string{"stable", "canary"}, ok == true
//
//	// Single-value promotion:
//	xc.Set("env", "production")
//	envs, ok := xc.GetStringCollection("env")
//	// envs == []string{"production"}, ok == true
func (c *XConfig) GetStringCollection(key string) ([]string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		switch v := val.Value.(type) {
		case []string:
			return v, true
		case string:
			return []string{v}, true
		}
	}
	return nil, false
}

// GetBoolCollection returns the value of key as a []bool slice. A single bool
// scalar is wrapped in a one-element slice. Returns (nil, false) when the key
// does not exist or holds an incompatible type.
//
//	// Config file:
//	//   flag=true
//	//   flag=false
//	//   flag=true
//	flags, ok := xc.GetBoolCollection("flag")
//	// flags == []bool{true, false, true}, ok == true
func (c *XConfig) GetBoolCollection(key string) ([]bool, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		switch v := val.Value.(type) {
		case []bool:
			return v, true
		case bool:
			return []bool{v}, true
		}
	}
	return nil, false
}

// GetIntCollection returns the value of key as a []int slice. A single int
// scalar is wrapped in a one-element slice. Returns (nil, false) when the key
// does not exist or holds an incompatible type.
//
//	// Config file:
//	//   port=8080
//	//   port=8081
//	//   port=8082
//	ports, ok := xc.GetIntCollection("port")
//	// ports == []int{8080, 8081, 8082}, ok == true
func (c *XConfig) GetIntCollection(key string) ([]int, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		switch v := val.Value.(type) {
		case []int:
			return v, true
		case int:
			return []int{v}, true
		}
	}
	return nil, false
}

// GetFloatCollection returns the value of key as a []float64 slice. A single
// float64 scalar is wrapped in a one-element slice. Returns (nil, false) when
// the key does not exist or holds an incompatible type.
//
//	// Config file:
//	//   ratio=0.25
//	//   ratio=0.50
//	//   ratio=0.75
//	ratios, ok := xc.GetFloatCollection("ratio")
//	// ratios == []float64{0.25, 0.50, 0.75}, ok == true
func (c *XConfig) GetFloatCollection(key string) ([]float64, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		switch v := val.Value.(type) {
		case []float64:
			return v, true
		case float64:
			return []float64{v}, true
		}
	}
	return nil, false
}

// GetTimeCollection returns the value of key as a []time.Time slice. A single
// time.Time value is wrapped in a one-element slice. Returns (nil, false) when
// the key does not exist or holds an incompatible type.
func (c *XConfig) GetTimeCollection(key string) ([]time.Time, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		switch v := val.Value.(type) {
		case []time.Time:
			return v, true
		case time.Time:
			return []time.Time{v}, true
		}
	}
	return nil, false
}

// Clone returns a fully independent deep copy of the XConfig as an
// xcore.XDatasetDef. All nested sub-configurations and all array slices are
// recursively cloned; modifying the clone does not affect the original, and
// modifying the original does not affect the clone.
//
// The returned value can be type-asserted to *XConfig when a concrete type is
// needed:
//
//	snapshot := original.Clone().(*xconfig.XConfig)
//	snapshot.Set("env", "test")
//	// original remains unchanged
func (c *XConfig) Clone() xcore.XDatasetDef {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	cloned := New()
	for id, val := range c.Parameters {
		cloned.Parameters[id] = *(&val).Clone()
	}
	for id, val := range c.Comments {
		cloned.Comments[id] = val
	}
	cloned.Order = make([]string, len(c.Order))
	copy(cloned.Order, c.Order)
	cloned.Multiple = c.Multiple
	return cloned
}

// GetConfig returns the nested sub-XConfig stored under key, or nil when the
// key does not exist or its value is not a *XConfig.
//
//	// Config file:
//	//   database.host=db.example.com
//	//   database.port=5432
//	//   database.tls=true
//
//	db := xc.GetConfig("database")
//	if db != nil {
//	    host, _ := db.GetString("host")  // "db.example.com"
//	    port, _ := db.GetInt("port")     // 5432
//	    tls,  _ := db.GetBool("tls")     // true
//	}
func (c *XConfig) GetConfig(key string) *XConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.Parameters[key]; ok {
		if xc, ok := val.Value.(*XConfig); ok {
			return xc
		}
	}
	return nil
}

// ── Load / Merge ──────────────────────────────────────────────────────────────

// LoadFile reads the file at filename and populates the XConfig, replacing any
// existing entry that shares a key with one from the file. Entries that exist
// in the config but are absent from the file are left untouched.
//
// To add entries from the file without replacing existing ones, use MergeFile.
//
//	xc := xconfig.New()
//	if err := xc.LoadFile("/etc/myapp/app.conf"); err != nil {
//	    log.Fatal(err)
//	}
//	host, _ := xc.GetString("host")
//	port, _ := xc.GetInt("port")
//
// LoadFile returns an error if filename contains a path-traversal sequence
// (".."), if the file cannot be opened, or if any line cannot be parsed.
func (c *XConfig) LoadFile(filename string) error {
	if err := validatePath(filename); err != nil {
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.loadandparse(filename, false)
}

// MergeFile reads the file at filename and merges its entries into the XConfig.
// When a key already exists, the new value is appended to form an array;
// existing entries are not replaced. To replace existing entries instead, use
// LoadFile.
//
//	xc := xconfig.New()
//	xc.LoadFile("defaults.conf")
//	xc.MergeFile("local.conf")   // extends or supplements defaults
//
//	// Per-entry operators still take effect inside a merge:
//	// "host:=override.example.com" in local.conf will always replace host.
//
// MergeFile returns an error if filename contains a path-traversal sequence,
// if the file cannot be opened, or if any line cannot be parsed.
func (c *XConfig) MergeFile(filename string) error {
	if err := validatePath(filename); err != nil {
		return err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.loadandparse(filename, true)
}

// LoadString parses the configuration text in data and replaces any existing
// entries that share a key with entries from data. The format is identical to
// a configuration file. Use MergeString to add entries without replacing
// existing ones.
//
//	xc := xconfig.New()
//	xc.LoadString("host=localhost\nport=5432\ndebug=false\n")
//	host, _ := xc.GetString("host")   // "localhost"
//	port, _ := xc.GetInt("port")      // 5432
//	debug, _ := xc.GetBool("debug")   // false
func (c *XConfig) LoadString(data string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.parsestring(data, false)
}

// MergeString parses the configuration text in data and merges its entries
// into the XConfig. Existing entries are extended rather than replaced.
//
//	xc := xconfig.New()
//	xc.LoadString("tag=stable\n")
//	xc.MergeString("tag=latest\n")
//	tags, _ := xc.GetStringCollection("tag")
//	// tags == []string{"stable", "latest"}
func (c *XConfig) MergeString(data string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.parsestring(data, true)
}

// LoadXConfig copies all entries from src into c, replacing existing entries.
// src is deep-cloned before insertion, so subsequent modifications to either
// object do not affect the other. Use MergeXConfig to add entries without
// replacing existing ones.
//
// LoadXConfig is a no-op when src is nil.
//
//	base := xconfig.New()
//	base.LoadFile("base.conf")
//
//	active := xconfig.New()
//	active.LoadXConfig(base)         // independent copy of base
//	active.Set("env", "production")  // does not affect base
func (c *XConfig) LoadXConfig(data *XConfig) error {
	if data == nil {
		return nil
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.parsemap(data, false)
}

// MergeXConfig merges all entries from src into c. Existing entries are
// extended rather than replaced; src is not modified.
//
// MergeXConfig is a no-op when src is nil.
//
//	defaults := xconfig.New()
//	defaults.LoadFile("defaults.conf")
//
//	xc := xconfig.New()
//	xc.LoadFile("app.conf")
//	xc.MergeXConfig(defaults)   // fills in any keys absent from app.conf
func (c *XConfig) MergeXConfig(data *XConfig) error {
	if data == nil {
		return nil
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.parsemap(data, true)
}

// ── Serialisation ─────────────────────────────────────────────────────────────

// buildLevel serialises the XConfig at the given key prefix, recursing into
// nested sub-configurations. String values that would be misinterpreted by the
// parser on reload (boolean keywords, numeric literals, strings starting with
// '"') are written with a leading '"' character to preserve their type.
// Empty sentinel keys in Order are skipped to prevent index-out-of-range panics.
func (c *XConfig) buildLevel(prefix string) string {
	sdata := []string{}
	for _, val := range c.Order {
		if len(val) == 0 {
			continue // skip empty sentinel keys
		}
		if val[0] == '#' {
			sdata = append(sdata, c.Comments[val])
		} else {
			param := c.Parameters[val]
			switch param.paramtype {
			case 1: // string — protect ambiguous values with a leading '"'
				strval := param.Value.(string)
				if needsQuotePrefix(strval) {
					sdata = append(sdata, prefix+val+"=\""+strval)
				} else {
					sdata = append(sdata, prefix+val+"="+strval)
				}
			case 11: // []string — protect each element individually
				for _, v := range param.Value.([]string) {
					if needsQuotePrefix(v) {
						sdata = append(sdata, prefix+val+"=\""+v)
					} else {
						sdata = append(sdata, prefix+val+"="+v)
					}
				}
			case 12: // []int
				for _, v := range param.Value.([]int) {
					sdata = append(sdata, prefix+val+"="+strconv.Itoa(v))
				}
			case 13: // []float64
				for _, v := range param.Value.([]float64) {
					sdata = append(sdata, prefix+val+"="+fmt.Sprint(v))
				}
			case 14: // []bool
				for _, v := range param.Value.([]bool) {
					sdata = append(sdata, prefix+val+"="+fmt.Sprint(v))
				}
			case 21: // nested sub-XConfig — recurse with extended prefix
				sdata = append(sdata, param.Value.(*XConfig).buildLevel(prefix+val+"."))
			default:
				sdata = append(sdata, prefix+val+"="+fmt.Sprint(param.Value))
			}
		}
	}
	return strings.Join(sdata, "\n")
}

// Marshal returns the complete serialised text of the XConfig in the .conf
// format. The output preserves comments, blank lines, and the original
// insertion order. It can be written to a file and reloaded with LoadFile or
// LoadString without any loss of data or type information.
//
//	xc := xconfig.New()
//	xc.Set("host", "localhost")
//	xc.Set("port", 5432)
//	fmt.Print(xc.Marshal())
//	// Output:
//	// host=localhost
//	// port=5432
//	//
//
// String values that would be misinterpreted by the parser are automatically
// protected with the '"' prefix, so a round-trip is always lossless:
//
//	xc.Set("label", "true")      // string, not bool
//	text := xc.Marshal()         // produces: label="true
//	xc2 := xconfig.New()
//	xc2.LoadString(text)
//	v, _ := xc2.GetString("label")  // "true"
func (c *XConfig) Marshal() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.buildLevel("") + "\n"
}

// SaveFile serialises the XConfig and writes it to filename.
//
// The file is written with mode 0644. Comments, blank lines, and the original
// insertion order are preserved. String values that would be misinterpreted by
// the parser on reload are automatically protected with the '"' prefix, making
// the save-then-load cycle lossless.
//
//	xc := xconfig.New()
//	xc.LoadFile("app.conf")
//	xc.Set("version", "1.2.3")
//	if err := xc.SaveFile("app.conf"); err != nil {
//	    log.Fatal(err)
//	}
//
// SaveFile returns an error if filename contains a path-traversal sequence or
// if the file cannot be written.
func (c *XConfig) SaveFile(filename string) error {
	if err := validatePath(filename); err != nil {
		return err
	}
	c.mutex.RLock()
	data := c.buildLevel("") + "\n"
	c.mutex.RUnlock()
	return os.WriteFile(filename, []byte(data), 0644)
}
