@UTF-8

XConfig for GO v0.0.1
=============================
This is the build 1

The XConfig library is used to easily build a config object based on a descriptor file.

Version Changes Control
=======================

V0.0.1 - 2018-11-06
-----------------------
- First commit, still not fully working


Manual
=======================

XConfig loads a configuration file similar to a .ini file, but with some important improvements.

Basic use:
----------

The configuration file have the following syntax for example:

```
# this file is named example.conf, used in following examples
# the # denotes a comment.
; is also a comment
parameter1=value1
parameter2=value2
parameter2=value3
```

XConfig accept "true", "on", "yes" as a boolean 'true' and "false", "off", "no", "none" as a boolean 'false'.
For instance, that means parameter=off is a boolean false, and parameter=yes is a boolean true in the XConfig structure.

XConfig also convert all integers to an int parameter in the XConfig structure

The XConfig object is easily usable as:
```
config := &xconfig.XConfig{}
confif.Load("./example.conf")
confif.Merge("./mergeme.conf")
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
Merge*: to merge a file, a string dataset, or another XConfig dataset. Merging means all new entries will be added to the already existing parameters.
  This is userfull then you split your config file into subset of parameters each (for instance database config, memory config, internationalization config, etc)
Get/Set/Add: to read, set (replace) or add (merge) parameters to the XConfig.

Once you have an instance of your configuration, you may use it like this:

```
// assign a local variable
param1 := config.Get("parameter1")
fmt.Println(param1)

// assign a casted local variable
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
----------------

Default values:
---------------

You may pass an array of default values to the constructor so if the parameter is *not present* into the config file, it will take the default value.
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

Merging:
--------

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

---
