@UTF-8

XConfig for GO v0
=============================

The XConfig library is used to easily build a config object based on a descriptor file.


TO DO:
======

- Merge vs load to load more than 1 file (pending)
- implement + and :
- implement Save
- Add a flag when it's a multiple load to warn a "save"
- Add error control con Get* for type conversions (if the type is different as expected)
- Integrate with xcore.XDataset


Version Changes Control
=======================

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




Manual:
=======================

XConfig loads a configuration file similar to a .ini file, but with some important improvements:
- You can load more than one file in merge mode or replacing mode in the same config object.
- You can set and get parameters at any time.
- You can load the config object based on a file, but also on a string and another object; this way you can serialize, unserialize, transfer by any strem, save or load on database the object, etc.
- You can also save ack the config string with all its comments

Basic use:
----------

1. Installing the package:

Execute on your operating system:

```
go get github.com/webagility-go/xconfig
```

2. Importing the package:

Adds to the import section:

```
import "github.com/webability-go/xconfig"
```

3. The configuration files:

3.1 General syntax

The configuration file is a set of key=value parameters, with optional comments.

The configuration file have the following syntax:

```
# this file is named example.conf, used in following examples
# the # denotes a comment.
; is also a comment
parameter1=value1
parameter2=value2
# Repeating the same parameter will create an array of values for this parameter
parameter2=value3   

# Creates a sub XConfig dataset for parameter, then again for subparameter
parameter.subparameter.subparameter2=value

# Force to add the value to the parameter (with an extra '+'):
parameter2+=value4

# Replace any value the parameter already has by this one (with an extra ':'):
parameter2:=value4

```

3.2 Parameter keys:

The parameter key is a string with characters [a-zA-Z0-9_-] only, with a minimum of 1 character.
The parameters may have a + or a * after their name, to denote "add the value" or "replace the value" forced rule.
Without the + and * (in normal operation), the "add" or "replace" behaviour depends on the funcion called for loading the configuration (Load* or Merge* functions)
The point (.) denotes a sub set of parameters (a new sub XConfig dataset for this parameter)

3.3 Parameter values:

There are 4 types of values:
- Strings
- Integer
- Float
- Boolean

The value has no restrictions except it must enter into the line (no line breaks allowed)
The compiler accepts strings "true", "on", "yes" as a boolean 'true' and "false", "off", "no", "none" as a boolean 'false'.
For instance, that means parameter=off is a boolean false, and parameter=yes is a boolean true in the XConfig structure.

The compiler also convert all integers to an int parameter in the XConfig structure, and float values as float64 type.
If you want a natural integer, float or boolean interpreted as a string, you must start it with a " character:
param1="123   will be the string **123** in the XConfig structure

If you want a string starting with a ", you will need to put 2 " at the begining:
param=""abc   will be the string **"abc** in the XConfig structure

When you insert more than one value for a parameter, it creates an array of the type of the first found value. If you have a mixed type of values, you will get an error

for instance:
```
parameter1=true
parameter1=123
parameter1=hello
```
"123" and "hello" are not boolean so you get an error reading your definition file.

Note that is the first parameter is a string, all new values will should start with " to be considered as a string also:
```
parameter1=hello
parameter1="true
parameter1="123
```
you will obtain an array []string with values ["hello", "true", "123"]

The order IS important.

4. Invoking the XConfig in your code:

First of all, you need to import the library in your code:

```
import (
  "github.com/webagility-go/xconfig"
)
```

Then you need first to create a blank XConfig instance:

```
xc := xonfig.New()
```

Then, you generally load a file to fill in your XConfig definition

```
xc.LoadFile("/path/to/my/file.conf")
```

And finally use the configuration

```
myparam := xc.Get("myparam")
```

myparam will take the type of the parameter: string, integer, float64, bool, or an array of string, integer or float64
(you should be aware of the type of your parameter before using it)


5. Merging vs Loading


Advanced use:
-------------



Package reference:
------------------

1. New, Declaration

2. Set

3. Add

4. Get

5. LoadFile

6. MergeFile

7. LoadString

8. MergeString

9. LoadXConfig

10. MergeXConfig

11. Clone


Internals:
----------







The XConfig object is easily usable as:
```
# Using the New function
config := xconfig.New(nil)

# Auto-new operator
config := &xconfig.XConfig{}

# Default new operator
config := new(xconfig.XConfig)
```

or, if you load your own file by other means (remote, database etc)
```
config := &xconfig.XConfig{}
mydata := getMyParameters()   // get the whole configuration file into mydata
confif.LoadString(mydata)
```

or, if you already have your configuration into a Map of Strings (unserialized, etc)
```
config := &xconfig.XConfig{}
mydata := map[string]string{"param1":"value1","param2":"value2"}
confif.LoadXConfig(mydata)
```

There are 3 sets of public functions:
Load*: to load a file, a string dataset, or another XConfig dataset. Loading means all already existing parameters will be replaced by the new configuration.
  This is usefull when you have a main config file, and a local config file that must replace some values
  Functions are LoadFile, LoadString and LoadXConfig
Merge*: to merge a file, a string dataset, or another XConfig dataset. Merging means all new entries will be added to the already existing parameters.
  This is userfull then you split your config file into subset of parameters each (for instance database config, memory config, internationalization config, etc)
  Functions are MergeFile, MergeString and MergeXConfig
Get/Set/Add: to read, set (replace) or add (merge) parameters to the XConfig.

Once you have an instance of your configuration, you may use it like this:

```
// assign a local variable
param1 := config.Get("parameter1")
fmt.Println(param1)

// assign to an already casted local variable
var param2 string
param2 = config.Get("parameter2").(string)  // be carefull that the parameter IS actually the same cast or an error is thrown
fmt.Println(param2)

// use directly the parameters
for p, v := range config {
  fmt.Printf("%s=%v\n", p, v)

// set a new parameter
config.Set("parameter3", "value3")
config.Set("parameter3", "new value3") // will be replaced
config.Add("parameter3", "another value3") // will be replaced by an array of values with both entries into it
config.Set("parameter4", 12345)
config.Set("parameter5", true)
```

Advanced topics:
================

1. Default values:
------------------

You may pass a map of default values to the XCOnfig so if the parameter is *not present* into the config file, it will take the default value.
Note: default values will be taken only if the parameter DOES NOT EXIST into the config file. This means an empty value is considerated as a value

Something like this:
parameter1=
will not fire the default value because the parameter is present into the config file

Example:
--------

class myConfig extends XConfig
{
  private $default = array(
    'parameter1' => 'default1'
  );

  public function __construct($data)
  {
    parent::__construct($data, $this->default);
  }
}

Merging vs Loading:
-------------------

+ and :

You may merge two config file (or more), for example when you have a master config file and a local replacement values config file:
```
include_once 'include/xconfig/XConfig.class.php');
$globalconfig = new XConfig(file_get_contents('myglobalconfig.conf'));
$localconfig = new XConfig(file_get_contents('mylocalconfig.conf'));
$globalconfig->merge($localconfig);
```
with files:
```
#global config:
ip=127.0.0.1
port=80
domain=test.com
```
```
#local config:
port=8080
title=Welcome
```

The result config after merging local into global will be:
```
ip=127.0.0.1
port=8080
domain=test.com
title=Welcome
```

Sub dataset:
------------





---
