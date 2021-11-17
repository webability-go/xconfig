// Copyright Philippe Thomassigny 2004-2020.
// Use of this source code is governed by a MIT licence.
// license that can be found in the LICENSE file.

// Package xconfig loads a configuration file similar to a .ini file, but with some important improvements:
//
// - The xconfig recognize bool, int, float and Strings, and also collections of values and hierarchical nested sub-config sets.
//
// - The xconfig is compatible with XDataset to inject and use in templates, database records, etc.
//
// - You can load more than one file in merge mode or replacing mode in the same config object.
//
// - You can set and get parameters at any time.
//
// - You can load the config object based on a file, but also on a string and another object; this way you can serialize, unserialize, transfer by any strem, save or load on database the object, etc.
//
// - You can also save back the config string with all its comments.
//
//
// Basic use
//
// 1. Installing the package:
//
// Execute on your operating system:
//
//  go get -u github.com/webagility-go/xconfig
//
// 2. Importing the package:
//
//  import "github.com/webability-go/xconfig"
//
// 3. Then you need first to create a blank XConfig instance:
//
//  xc := xonfig.New()
//
// 4. Then, you generally load a file to fill in your XConfig definition
//
//  xc.LoadFile("/path/to/my/file.conf")
//
// 5. Use the configuration
//
//  myparam := xc.Get("myparam")
//
// myparam will take the type of the parameter: string, integer, float64, bool, or an array of string, integer or float64.
// (you should be aware of the type of your parameter before using it).
// You can also use the Get* cast functions.
//
// 6. You can also set some new parameters
//
//  xc.Set("mynewparam", "newvalue")
//
// 7. And finally save the new configuration
//
//  xc.SaveFile("/path/to/my/file.conf")
//
//
//
// File format reference
//
// The config file is a simple utf8 flat text file.
// The configuration file is a set of key=value parameters, with optional comments.
//
// The configuration file have the following syntax:
//
//  # this file is named example.conf, used in following examples
//  # the # denotes a comment.
//  ; is also a comment
//  parameter1=value1
//  parameter2=value2
//  # Repeating the same parameter will create an array of values for this parameter
//  parameter2=value3
//
//  # Creates a sub XConfig dataset for parameter, then again for subparameter
//  parameter.subparameter.subparameter2=value
//
//  # Force to add the value to the parameter (with an extra '+'), in this case to the array of string values of parameter2
//  parameter2+=value4
//
//  # Replace any value the parameter already has by this one (with an extra ':'), in this case parameter2 is a string again
//  parameter2:=value4
//
// You can add as many as parameters you wish into the file.
//
//
// 1. comments:
//
// You may add comments and also comment unused parameter with # or ; at the beginning of the line
//
//  # This is the config file for my application
//  MAINPATH=/home/var
//
//  # Unused parameter:
//  # DOMAIN=mydomain.com
//
//
// 2. Parameter keys:
//
// The parameter key is a string with characters [a-zA-Z0-9_-] only, with a minimum of 1 character.
//
// The point (.) denotes a sub set of parameters (a new sub XConfig dataset for this parameter)
//
//  database.user=username
//  database.pass=password
//  database.db=dbname
//
// In this case the database entry of the XConfig is again another XConfig with 3 parameters  into it: user, pass and db.
//
//
// 3. Assignation sign:
//
// A simple = sign is the normal assignation, the "add" or "replace" behaviour depends on the funcion called for loading the configuration (Load* or Merge* functions).
//
// In this case various asignation to the same parameter will create an array of values of the same type as the first declared parameter.
//
// An equal sign preceded by a + (+=) will always add the parameter to the array of values, never replace it (see Merge/Load).
//
// An equal sign preceded by a : (:=) will always replace the parameter and discard any already set values.
//
//
// 4. Parameter values:
//
// There are 4 types of values:
//
// - Strings
//
// - Integer
//
// - Float
//
// - Boolean
//
// The value has no restrictions except it must enter into the line (no line breaks allowed)
// The compiler accepts strings "true", "on", "yes" as a boolean 'true' and "false", "off", "no", "none" as a boolean 'false'.
// For instance, that means parameter=off is a boolean false, and parameter=yes is a boolean true in the XConfig structure.
//
// The compiler also convert all integers to an int parameter in the XConfig structure, and float values as float64 type.
// If you want a natural integer, float or boolean interpreted as a string, you must start it with a " character:
// param1="123   will be the string 123 in the XConfig structure
//
// If you want a string starting with a ", you will need to put 2 " at the beginning:
// param=""abc   will be the string "abc in the XConfig structure
//
// 3. list of values:
//
// You can repeat as many time you need the same parameter name with different values.
// This will build a list of values in the object.
// The list of values is kept as an array of values.
//
// If you have a mixed type of values, you will get an error.
//
// for instance:
//
//  # Those are booleans
//  parameter1=true
//  parameter2=on
//  parameter3=no
//
//  # Those are integers
//  parameter4=0
//  parameter5=1
//  parameter6=234
//  parameter7=-5
//  parameter8=837456783456
//
//  # Those are floats
//  parameter10=0.0
//  parameter11=1.7
//  parameter12=234.5
//  parameter13=-5.834
//  parameter14=837456783.456
//  parameter15=-5.834e7
//
//  # Those are strings
//  parameter20=asdh
//  parameter21="1
//  parameter22="false
//  parameter23="-5.834
//  parameter24=""12345
//  parameter25=something 123 true false on off
//
//  # This parameter will force parameter1 to become an array of booleans [true, false]
//  parameter1=false
//
//  # This will throw an error since parameter1 is a boolean and abc is not a boolean
//  parameter1=abc
//
//  # Note that is the first parameter is a string, all new values will should start with " to be considered as a string also:
//  parameter30=hello
//  parameter30="true
//  parameter30="123
//  # you will obtain an array []string with values ["hello", "true", "123"]
//
//  # List of authorized languages:
//  languages=es
//  languages=en
//  languages=fr
//  languages=jp
//
// The order IS important.
//
// Once loaded you will get a []string{“es”, “en”, “fr”, “jp”} assigned to the “languages” parameter.
//
//
// Merging vs Loading
//
//
// + and :
//
// You may load two config file (or more), for example when you have a master config file and a local replacement values config file.
// In this case, the values of the second file will "replace" the data already loaded with the first file. This is a replacement of config entries.
//
// In the other hand, you may merge two config files (or more), for example when you need to fragment the data into a set of simpler files.
// In this case, the values will just be "added" as if the two files were a simple file (create arrays on same id entries). This is a merging of config entries.
//
// If you want to force the behaviour of variables (merging into a replacement loading, or replace into a merging loading), you may use the two
// assignement operators + and :
//
// := will be used to replace a value by a new value, meanwhile += will be used to add a new value to the array of data.
// The rules of "same type of data" must be observed also in this case (+=)
//
//  import "github.com/webability-go/xconfig"
//
//  xc := xconfig.New()
//  xc.LoadFile("/path/to/my/file.conf")
//  xm := xconfig.New()
//  xm.LoadFile("/path/to/my/file2.conf")
//  // merge it:
//  xc.Merge(xm)
//
// With the following files:
//
//  #file.conf: global config:
//  ip=127.0.0.1
//  port=80
//  domain=test.com
//
//  #file2.conf: local config:
//  port=8080
//  title=Welcome
//
// The result config after merging local into global will be:
//
//  ip=127.0.0.1
//  port=8080
//  domain=test.com
//  title=Welcome
//
// Advanced use
//
// The XConfig object is easily usable as:
//
//  // Using the New function
//  config := xconfig.New(nil)
//
//  // Auto-new operator
//  config := &xconfig.XConfig{}
//
//  // Default new operator
//  config := new(xconfig.XConfig)
//
// or, if you load your own file by other means (remote, database etc)
//
//  config := &xconfig.XConfig{}
//  mydata := getMyParameters()   // get the whole configuration file into mydata string
//  config.LoadString(mydata)
//
// or, if you already have your configuration into a Map of Strings (unserialized, etc)
//
//  config := &xconfig.XConfig{}
//  mydata := map[string]string{"param1":"value1","param2":"value2"}
//  config.LoadXConfig(mydata)
//
// There are 3 sets of public functions:
//
// Load*: to load a file, a string dataset, or another XConfig dataset. Loading means all already existing parameters will be replaced by the new configuration.
//
// This is useful when you have a main config file, and a local config file that must replace some values
// Functions are LoadFile, LoadString and LoadXConfig
//
// Merge*: to merge a file, a string dataset, or another XConfig dataset. Merging means all new entries will be added to the already existing parameters.
//
// This is userfull then you split your config file into subset of parameters each (for instance database config, memory config, internationalization config, etc)
// Functions are MergeFile, MergeString and MergeXConfig
//
// Get/Set/Add: to read, set (replace) or add (merge) parameters to the XConfig.
//
// Once you have an instance of your configuration, you may use it like this:
//
//  // assign a local variable
//  param1 := config.Get("parameter1")
//  fmt.Println(param1)
//
//  // assign to an already casted local variable
//  var param2 string
//  param2 = config.Get("parameter2").(string)  // be careful that the parameter IS actually the same cast or an error is thrown
//  fmt.Println(param2)
//  // Or, since XConfig is XDataset extended, you may use the autocast functions:
//  param2 = config.GetString("parameter2")  // Will be converted to a string always
//
//  // use directly the parameters
//  for p, v := range config {
//    fmt.Printf("%s=%v\n", p, v)
//  }
//
//  // set a new parameter
//  config.Set("parameter3", "value3")
//  config.Set("parameter3", "new value3") // will be replaced
//  config.Add("parameter3", "another value3") // will be replaced by an array of values with both entries into it
//  config.Set("parameter4", 12345)
//  config.Set("parameter5", true)
//
//
//
// Saving configuration
//
// After loading, creating, modifying your configuration file, you may need to save your configuration.
//
// For this yoy have 2 functions:
//
//
//  config := &xconfig.XConfig{}
//  config.Set("myparam1", "My value")
//  // Create a string with the content of the new config:
//  fmt.Println(config.Marshal())
//
//  // Directly save the modified config file:
//  config.SaveFile("path/to/your/file.conf")
//
// Note: if you load your configuration file with comments in it, when you save it, the comments and presentation (new lines) will be respected.
// If you add new parameters, they will be added to the end of the file. New lines will be removed into the definition of an array of data.
//
//
package xconfig

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/webability-go/xcore/v2"
)

// VERSION is the used version nombre of the XCore library.
const VERSION = "0.4.3"

// Parameter is the basic entry parameter into the configuration object
// Value is the value of the parameter.
type Parameter struct {
	// type =
	//  0: not set yet
	//  1: string, 11 = array of string
	//  2: integer, 12 = array of int
	//  3: float64, 13 = array of float64
	//  4: bool, 14 = array of boolean
	// 21: sub XConfig
	paramtype int
	// Value of the parameter ()
	Value interface{}
	// assignment =
	//  0: default behaviour
	//  1: forced :=
	//  2: forced +=
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

func (p *Parameter) add(paramtype int, value interface{}, assignment int) error {
	switch p.paramtype {
	case 0: // not set yet
		p.paramtype = paramtype
		p.Value = value
	case 1: // string
		if paramtype == 1 {
			// transform the parameter into an array and change paramtype
			sub := make([]string, 0, 2)
			p.Value = append(sub, p.Value.(string), value.(string))
			p.paramtype = 11
		} else if paramtype == 11 {
			// concatenate array of string
			p.Value = append(value.([]string), p.Value.(string))
		} else {
			return errors.New("The parameter cannot add an incompatible value to a string")
		}
	case 2: // integer
		if paramtype == 2 {
			// transform the parameter into an array and change paramtype
			sub := make([]int, 0, 2)
			p.Value = append(sub, p.Value.(int), value.(int))
			p.paramtype = 12
		} else if paramtype == 12 {
			// concatenate array of int
			p.Value = append(value.([]int), p.Value.(int))
		} else {
			return errors.New("The parameter cannot add an incompatible value to an integer")
		}
	case 3: // float64
		if paramtype == 3 {
			// transform the parameter into an array and change paramtype
			sub := make([]float64, 0, 2)
			p.Value = append(sub, p.Value.(float64), value.(float64))
			p.paramtype = 13
		} else if paramtype == 13 {
			// concatenate array of float64
			p.Value = append(value.([]float64), p.Value.(float64))
		} else {
			return errors.New("The parameter cannot add an incompatible value to a float")
		}
	case 4: // boolean
		if paramtype == 4 {
			// transform the parameter into an array and change paramtype
			sub := make([]bool, 0, 2)
			p.Value = append(sub, p.Value.(bool), value.(bool))
			p.paramtype = 14
		} else if paramtype == 14 {
			// concatenate array of bool
			p.Value = append(value.([]bool), p.Value.(bool))
		} else {
			return errors.New("The parameter cannot add an incompatible value to a boolean")
		}
	case 11: // array of string
		if paramtype == 1 {
			p.Value = append(p.Value.([]string), value.(string))
		} else if paramtype == 11 {
			// concatenate array of string
			p.Value = append(p.Value.([]string), value.([]string)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to an array of strings")
		}
	case 12: // array of int
		if paramtype == 2 {
			p.Value = append(p.Value.([]int), value.(int))
		} else if paramtype == 12 {
			// concatenate array of int
			p.Value = append(p.Value.([]int), value.([]int)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to an array of integers")
		}
	case 13: // array of float64
		if paramtype == 3 {
			p.Value = append(p.Value.([]float64), value.(float64))
		} else if paramtype == 13 {
			// concatenate array of float64
			p.Value = append(p.Value.([]float64), value.([]float64)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to an array of integers")
		}
	case 14: // array of float64
		if paramtype == 4 {
			p.Value = append(p.Value.([]bool), value.(bool))
		} else if paramtype == 14 {
			// concatenate array of bool
			p.Value = append(p.Value.([]bool), value.([]bool)...)
		} else {
			return errors.New("The parameter cannot add an incompatible value to an array of booleans")
		}
	case 21: // XConfig
		// pass the addparam to the subset XConfig
	default:
		return errors.New("Unknow parameter type")
	}
	return nil
}

// Clone is the parameter method to call to make a full clone of the information
func (p *Parameter) Clone() *Parameter {
	cloned := newParam()
	clonedval := p.Value
	if cloneable, ok := clonedval.(interface{ Clone() xcore.XDatasetDef }); ok {
		clonedval = cloneable.Clone()
	}
	cloned.set(p.paramtype, clonedval, p.assignment)
	return cloned
}

// XConfigDef is the config definition type
type XConfigDef interface {
	xcore.XDatasetDef
}

// XConfig is the main config object
type XConfig struct {
	Parameters  map[string]Parameter
	Comments    map[string]string
	Order       []string
	Multiple    bool
	multithread bool
	mutex       sync.RWMutex
}

// New is called to create a new empty XConfig object
func New() *XConfig {
	c := &XConfig{
		Parameters: make(map[string]Parameter),
		Comments:   make(map[string]string),
		Multiple:   false,
	}
	return c
}

// ========================
//  The private functions used to control the XConfig structre and load strings and files
func (c *XConfig) addcomment(line int, comment string) error {
	id := "#" + strconv.Itoa(line)
	c.Comments[id] = comment
	c.Order = append(c.Order, id)
	return nil
}

func (c *XConfig) addparam(line int, key string, typeparam int, value interface{}, assignment int) error {
	// check if key contains "+" (forced array) and . (subset of config)
	// and creates a Map[] if the value already exists (or just set it)
	//  keydata, merge := analyzeKey(key)
	//  mustmerge := false
	pospoint := strings.Index(key, ".")
	if pospoint >= 0 {
		firstkey := strings.TrimSpace(key[:pospoint])
		subkey := strings.TrimSpace(key[pospoint+1:])

		if val, ok := c.Parameters[firstkey]; ok {
			// already exists: add the sub parameters, val is an *XConfig
			val.Value.(*XConfig).addparam(line, subkey, typeparam, value, assignment)
		} else {
			// no existe
			p := newParam()
			err := p.add(21, New(), assignment)
			if err != nil {
				return err
			}
			p.Value.(*XConfig).addparam(line, subkey, typeparam, value, assignment)
			c.Parameters[firstkey] = *p
			c.Order = append(c.Order, firstkey)
		}
	} else {
		if val, ok := c.Parameters[key]; ok {
			p := newParam()
			err := p.add(val.paramtype, val.Value, assignment)
			if err != nil {
				return err
			}
			err = p.add(typeparam, value, assignment)
			if err != nil {
				return err
			}
			c.Parameters[key] = *p
		} else {
			p := newParam()
			err := p.add(typeparam, value, assignment)
			if err != nil {
				return err
			}
			c.Parameters[key] = *p
			c.Order = append(c.Order, key)
		}
	}
	return nil
}

func (c *XConfig) setparam(line int, key string, typeparam int, value interface{}, assignment int) error {
	// check if key contains "+" (forced array) and . (subset of config)
	// and creates a Map[] if the value already exists (or just set it)
	//  keydata, merge := analyzeKey(key)
	//  mustmerge := false
	p := newParam()
	p.add(typeparam, value, assignment)
	if _, ok := c.Parameters[key]; !ok {
		c.Order = append(c.Order, key)
	}
	c.Parameters[key] = *p
	return nil
}

func (c *XConfig) parseline(line int, data string, merge bool) error {

	posequal := strings.Index(data, "=")

	// we ignore empty and comments lines, no key=value lines too
	if len(data) == 0 || data[0] == '#' || data[0] == ';' || posequal < 0 {
		return c.addcomment(line, data)
	}

	// we separate the key. if there is no key, we ignore the data
	key := strings.TrimSpace(data[:posequal])
	if len(key) == 0 {
		return c.addcomment(line, data)
	}

	// analyse si := or +=
	assignment := 0

	// we capture the value if it exists. If not, the key entry is initialized with a nil value
	var value interface{}
	var typeparam = 1
	if len(data) > posequal {
		strvalue := strings.TrimSpace(data[posequal+1:])
		value = strvalue
		if len(strvalue) > 0 && strvalue[0] == '"' {
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
				} else {
					if floatvalue, err := strconv.ParseFloat(strvalue, 64); err == nil {
						value = floatvalue
						typeparam = 3
					}
				}
			}
		}
	}
	return c.addparam(line, key, typeparam, value, assignment)
}

func (c *XConfig) parsemap(data *XConfig, merge bool) error {
	if len(c.Parameters) == 0 {
		c.Parameters = data.Parameters
		c.Comments = data.Comments
		c.Order = data.Order
	} else {
		line := len(c.Order)
		for p, v := range (*data).Parameters {
			if merge {
				c.addparam(line, p, v.paramtype, v.Value, v.assignment)
			} else {
				c.setparam(line, p, v.paramtype, v.Value, v.assignment)
			}
		}
		c.Multiple = true
	}
	return nil
}

func (c *XConfig) loadandparse(filename string, merge bool) error {
	// No filename: we let the config object as is
	if len(filename) == 0 {
		return nil
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
		err := tempConfig.parseline(ln, line, merge)
		if err != nil {
			return err
		}
		ln++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// We need a temporal xconfig and inject at the end because of the merge flag and the + and * flags (hard to change on the fly based on the existante of the old variable vs new variable)
	c.parsemap(tempConfig, merge)
	return nil
}

func (c *XConfig) parsestring(data string, merge bool) error {
	// No filename: we let the config object as is
	if len(data) == 0 {
		return nil
	}

	tempConfig := New()
	scanner := bufio.NewScanner(strings.NewReader(data))
	line := 0
	for scanner.Scan() {
		data := scanner.Text()
		err := tempConfig.parseline(line, data, merge)
		if err != nil {
			return err
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	//  fmt.Println(tempConfig)

	c.parsemap(tempConfig, merge)
	return nil
}

// String will create a string of the ordered content of the XConfig
func (c *XConfig) String() string {
	sdata := []string{}
	for _, val := range c.Order {
		if val[0] == '#' {
			sdata = append(sdata, c.Comments[val])
		} else {
			sdata = append(sdata, val+":"+fmt.Sprint(c.Parameters[val].Value))
		}
	}
	return "XConfig[\n" + strings.Join(sdata, "\n") + "\n]\n"
}

// GoString will create a string of the ordered content of the XConfig (based on String)
func (c *XConfig) GoString() string {
	return "#" + c.String()
}

// Set will replace or create the value of the key entry
func (c *XConfig) Set(key string, value interface{}) {
	// check if key contains "+" (forced array) and . (subset of config)
	// and just replace the value
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
	c.setparam(0, key, valuetype, value, 1)
}

// Add will adds a value to the structure. If the key entry already exists, then try to build a collection of it
func (c *XConfig) Add(key string, value interface{}) error {
	// check if key contains "+" (forced array) and . (subset of config)
	// and creates a Map[] if the value already exists (or just set it)
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
		return errors.New("The XConfig.Add function only accept string, integer, float64 and boolean values")
	}
	return c.addparam(0, key, valuetype, value, 0)
}

// Get will return the value of the key entry
// return false as second parameter if the entry does not exists (remember a value can be NIL and exists)
func (c *XConfig) Get(key string) (interface{}, bool) {
	// check if key contains "." (subset of config)
	if val, ok := c.Parameters[key]; ok {
		return val.Value, true
	}
	return nil, false
}

// GetDataset will return the key entry data as an XDataset if it exists and is a XDatasetDef
// return false as second parameter if the entry does not exists (remember a value can be NIL and exists)
func (c *XConfig) GetDataset(key string) (xcore.XDatasetDef, bool) {
	// check if key contains "." (subset of config)
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case *XConfig:
			return val.Value.(*XConfig), true
		}
	}
	return nil, false
}

// GetCollection will return the key entry data as an XDatasetCollectionDef if it exists and is a XDatasetCollectionDef
// return false as second parameter if the entry does not exists (remember a value can be NIL and exists)
func (c *XConfig) GetCollection(key string) (xcore.XDatasetCollectionDef, bool) {
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case xcore.XDatasetCollectionDef:
			return val.Value.(xcore.XDatasetCollectionDef), true
		}
	}
	return nil, false
}

// GetString will return the key entry data as a string, or ""
// return false as second parameter if the entry does not exists (remember a value can be "" and exists)
func (c *XConfig) GetString(key string) (string, bool) {
	// check if key contains "." (subset of config)
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case string:
			return val.Value.(string), true
		default:
			return fmt.Sprint(val.Value), true
		}
	}
	return "", false
}

// GetInt will return the key entry data as an int, or 0
// return false as second parameter if the entry does not exists (remember a value can be 0 and exists)
func (c *XConfig) GetInt(key string) (int, bool) {
	// check if key contains "." (subset of config)
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case int:
			return val.Value.(int), true
		case float64:
			return int(val.Value.(float64)), true
		case bool:
			if val.Value.(bool) {
				return 1, true
			}
			return 0, true
		}
	}
	return 0, false
}

// GetFloat will return the key entry data as a float64, or 0
// return false as second parameter if the entry does not exists (remember a value can be 0 and exists)
func (c *XConfig) GetFloat(key string) (float64, bool) {
	// check if key contains "." (subset of config)
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case float64:
			return val.Value.(float64), true
		case int:
			return float64(val.Value.(int)), true
		case bool:
			if val.Value.(bool) {
				return 1.0, true
			}
			return 0.0, true
		}
	}
	return 0, false
}

// GetTime will return the key entry data as a time, or 0
// return false as second parameter if the entry does not exists (remember a value can be 0 and exists)
func (c *XConfig) GetTime(key string) (time.Time, bool) {
	// check if key contains "." (subset of config)
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case time.Time:
			return val.Value.(time.Time), true
		}
	}
	return time.Time{}, false
}

// GetBool will return the key entry data as a boolean, or false
// return false as second parameter if the entry does not exists (remember a value can be false and exists)
func (c *XConfig) GetBool(key string) (bool, bool) {
	// check if key contains "." (subset of config)
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case bool:
			return val.Value.(bool), true
		case int:
			return val.Value.(int) != 0, true
		case float64:
			return val.Value.(float64) != 0, true
		}
	}
	return false, false
}

// GetStringCollection will return the key entry data as a []string, or nil
// return false as second parameter if the entry does not exists (remember a value can be nil and exists)
func (c *XConfig) GetStringCollection(key string) ([]string, bool) {
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case []string:
			return val.Value.([]string), true
		case string:
			return []string{val.Value.(string)}, true
		}
	}
	return nil, false
}

// GetBoolCollection will return the key entry data as a []bool, or nil
// return false as second parameter if the entry does not exists (remember a value can be nil and exists)
func (c *XConfig) GetBoolCollection(key string) ([]bool, bool) {
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case []bool:
			return val.Value.([]bool), true
		case bool:
			return []bool{val.Value.(bool)}, true
		}
	}
	return nil, false
}

// GetIntCollection will return the key entry data as a []int, or nil
// return false as second parameter if the entry does not exists (remember a value can be nil and exists)
func (c *XConfig) GetIntCollection(key string) ([]int, bool) {
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case []int:
			return val.Value.([]int), true
		case int:
			return []int{val.Value.(int)}, true
		}
	}
	return nil, false
}

// GetFloatCollection will return the key entry data as a []float64, or nil
// return false as second parameter if the entry does not exists (remember a value can be nil and exists)
func (c *XConfig) GetFloatCollection(key string) ([]float64, bool) {
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case []float64:
			return val.Value.([]float64), true
		case float64:
			return []float64{val.Value.(float64)}, true
		}
	}
	return nil, false
}

// GetTimeCollection will return the key entry data as a []Time, or nil
// return false as second parameter if the entry does not exists (remember a value can be nil and exists)
func (c *XConfig) GetTimeCollection(key string) ([]time.Time, bool) {
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case []time.Time:
			return val.Value.([]time.Time), true
		case time.Time:
			return []time.Time{val.Value.(time.Time)}, true
		}
	}
	return nil, false
}

// Del will delete then entry key it exists
func (c *XConfig) Del(key string) {
	delete(c.Parameters, key)
	delete(c.Comments, key)
	// deletes from Order and comments too
	for idx, v := range c.Order {
		if v == key {
			c.Order = append(c.Order[:idx], c.Order[idx+1:]...)
		}
	}
}

// Clone will perform a full clone of the whole structure
func (c *XConfig) Clone() xcore.XDatasetDef {
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

// GetConfig will return the key entry data as a XConfig, or nil
// This is similar to the GetDataset function
func (c *XConfig) GetConfig(key string) *XConfig {
	// check if key contains "." (subset of config)
	if val, ok := c.Parameters[key]; ok {
		switch val.Value.(type) {
		case *XConfig:
			return val.Value.(*XConfig)
		}
	}
	return nil
}

// LoadFile will try to load the file and parse it into the XConfig structure
func (c *XConfig) LoadFile(filename string) error {
	return c.loadandparse(filename, false)
}

// MergeFile will try to load the file and parse it into the XConfig structure, merging the entries to the existing ones
func (c *XConfig) MergeFile(filename string) error {
	return c.loadandparse(filename, true)
}

// LoadString will parse the string into the XConfig structure
func (c *XConfig) LoadString(data string) error {
	return c.parsestring(data, false)
}

// MergeString will parse the string into the XConfig structure, merging the entries to the existing ones
func (c *XConfig) MergeString(data string) error {
	return c.parsestring(data, true)
}

// LoadXConfig will load the new XConfig into the existing one
func (c *XConfig) LoadXConfig(data *XConfig) error {
	return c.parsemap(data, false)
}

// MergeXConfig will merge the new XConfig to the existing one
func (c *XConfig) MergeXConfig(data *XConfig) error {
	return c.parsemap(data, true)
}

func (c *XConfig) buildLevel(prefix string) string {
	sdata := []string{}
	for _, val := range c.Order {
		if val[0] == '#' {
			sdata = append(sdata, c.Comments[val])
		} else {
			switch c.Parameters[val].paramtype {
			case 11:
				a := c.Parameters[val].Value.([]string)
				for _, v := range a {
					sdata = append(sdata, prefix+val+"="+v)
				}
			case 12:
				a := c.Parameters[val].Value.([]int)
				for _, v := range a {
					sdata = append(sdata, prefix+val+"="+strconv.Itoa(v))
				}
			case 13:
				a := c.Parameters[val].Value.([]float64)
				for _, v := range a {
					sdata = append(sdata, prefix+val+"="+fmt.Sprint(v))
				}
			case 14:
				a := c.Parameters[val].Value.([]bool)
				for _, v := range a {
					sdata = append(sdata, prefix+val+"="+fmt.Sprint(v))
				}
			case 21:
				a := c.Parameters[val].Value.(*XConfig)
				sdata = append(sdata, a.buildLevel(prefix+val+"."))
			default:
				sdata = append(sdata, prefix+val+"="+fmt.Sprint(c.Parameters[val].Value))
			}
		}
	}
	return strings.Join(sdata, "\n")
}

func (c *XConfig) Marshal() string {
	return c.buildLevel("") + "\n"
}

func (c *XConfig) SaveFile(filename string) error {
	data := c.Marshal()
	return ioutil.WriteFile(filename, []byte(data), 0x644)
}

func analyzeKey(key string) (interface{}, bool) {
	mustmerge := false
	if key[len(key)-1] == '+' {
		mustmerge = true
		key = key[0 : len(key)-1]
	}
	if key[len(key)-1] == ':' {
		key = key[0 : len(key)-1]
	}

	match, _ := regexp.MatchString("[a-zA-Z0-9_-]+", key)
	return match, mustmerge

}
