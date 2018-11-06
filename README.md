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


Manuals
=======================

The configuration file have the following syntax for example:

```
# this file is named myconfig.conf, used in following examples
# the # denotes a comment.
; is also a comment
parameter1=value1
parameter2=value2
parameter2=value3
```

As version 1.1, xconfig now accept true, on, yes as a boolean 'true' and false, off, no, none as a boolean 'false'.
For instance, that means parameter=off is now a boolean false, and parameter=yes is now a boolean true.

Before verion 1.0, note the config file is always read as a STRING.
That means parameter=0, parameter=false, parameter=123
will be caught as "0", "false", "123", not integers or booleans

This will be converted into the XConfig object.
The XConfig object is easily usable as:
```
$config = new XConfig('String of the config file');
```
or
```
$config = new XConfig(Array of parameters);
```
Concrete Example 1:
```
include_once 'include/xconfig/XConfig.class.php');
$config = new XConfig(file_get_contents('myconfig.conf'));
```
Concrete Example 2:
```
include_once 'include/xconfig/XConfig.class.php');
$config = new XConfig(array(
    'parameter1' => 'value1',
    'parameter2' => array('value2', 'value3')
  ));
```
Once you have an instance of your configuration, you may use it like this:
```
// assign a local variable
$param1 = $config->parameter1;
echo $param1 . '<br />';

// use directly the variable
foreach($config->parameter2 as $p)
  echo $p . '<br />';

// set a new parameter
$config->parameter3 = 'value3';

// iterate the config
foreach($config as $parameter => $value)
  echo $parameter . ' = ' . $value . '<br />';
```  

Advanced topic
--------------

Default values:
---------------

You may pass an array of default values to the constructor so if the parameter is *not present* into the config file, it will take the default value.
Note: default values will be taken only if the parameter DOES NOT EXIST into the config file. This means an empty value is considerated as a value

Something like this:
parameter1=
will not fire the default value because the parameter is present into the config file

You may encapsulate the config object into a specific personal object with a local default set of parameters. 

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
