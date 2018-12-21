package xconfig

import (
  "fmt"
  "errors"
  "bufio"
  "os"
  "strings"
  "strconv"
  "regexp"
)

const VERSION = "0.0.4"

/* Basic parameter. 
   The type of the value can be 0 = not set, 1 = string, 2 = int, 3 = float, 4 = bool, 11 = array of strings, 12 = array of int, 13 = array of float, 14 = array of bool, 21 = XConfig
*/
type Parameter struct
{
  paramtype int
  Value interface{}
}

func newParam() *Parameter {
  return &Parameter{0, nil}
}

func (p *Parameter) set(paramtype int, value interface{}) {
  p.paramtype = paramtype
  p.Value = value
}

func (p *Parameter) add(paramtype int, value interface{}) error {
  switch p.paramtype {
    case 0:   // not set yet
      p.paramtype = paramtype
      p.Value = value
    case 1:   // string
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
    case 2:   // integer
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
    case 3:   // float64
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
    case 4:   // boolean
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
    case 11:   // array of string
      if paramtype == 1 {
        p.Value = append(p.Value.([]string), value.(string))
      } else if paramtype == 11 {
        // concatenate array of string
        p.Value = append(p.Value.([]string), value.([]string)...)
      } else {
        return errors.New("The parameter cannot add an incompatible value to an array of strings")
      }
    case 12:   // array of int
      if paramtype == 2 {
        p.Value = append(p.Value.([]int), value.(int))
      } else if paramtype == 12 {
        // concatenate array of int
        p.Value = append(p.Value.([]int), value.([]int)...)
      } else {
        return errors.New("The parameter cannot add an incompatible value to an array of integers")
      }
    case 13:   // array of float64
      if paramtype == 3 {
        p.Value = append(p.Value.([]float64), value.(float64))
      } else if paramtype == 13 {
        // concatenate array of float64
        p.Value = append(p.Value.([]float64), value.([]float64)...)
      } else {
        return errors.New("The parameter cannot add an incompatible value to an array of integers")
      }
    case 14:   // array of float64
      if paramtype == 4 {
        p.Value = append(p.Value.([]bool), value.(bool))
      } else if paramtype == 14 {
        // concatenate array of bool
        p.Value = append(p.Value.([]bool), value.([]bool)...)
      } else {
        return errors.New("The parameter cannot add an incompatible value to an array of booleans")
      }
    case 21:   // XConfig
      // pass the addparam to the subset XConfig
    default:
      return errors.New("Unknow parameter type")
  }
  return nil
}

type XConfig struct {
  Parameters map[string]Parameter
  Comments map[string]string
  Order []string
  Multiple bool
}

func New() *XConfig {
  c := &XConfig{
          Parameters: make(map[string]Parameter),
          Comments: make(map[string]string),
          Multiple: false,
       }
  return c
}

func (c *XConfig) addcomment (line int, comment string) error {
  id := "#" + strconv.Itoa(line)
  c.Comments[id] = comment
  c.Order = append(c.Order, id)
  return nil
}

func (c *XConfig) addparam(line int, key string, typeparam int, value interface{}) error {
  // check if key contains "+" (forced array) and . (subset of config)
  // and creates a Map[] if the value already exists (or just set it)
  //  keydata, merge := analyzeKey(key)
  //  mustmerge := false
  var err error
  err = nil
  
  pospoint := strings.Index(key, ".")
  if pospoint >= 0 {
    firstkey := strings.TrimSpace(key[:pospoint])
    subkey := strings.TrimSpace(key[pospoint+1:])
    
    if val, ok := (*c).Parameters[firstkey]; ok {
      // already exists: add the sub parameters, val is an *XConfig
      val.Value.(*XConfig).addparam(line, subkey, typeparam, value)
    } else {
      // no existe
      p := newParam()
      err = p.add(21, New() )
      p.Value.(*XConfig).addparam(line, subkey, typeparam, value)
      (*c).Parameters[firstkey] = *p
      c.Order = append(c.Order, firstkey)
    }
  } else {
    if val, ok := (*c).Parameters[key]; ok {
      p := newParam()
      err = p.add(val.paramtype, val.Value)
      err = p.add(typeparam, value)
      (*c).Parameters[key] = *p
    } else {
      p := newParam()
      err = p.add(typeparam, value)
      (*c).Parameters[key] = *p
      c.Order = append(c.Order, key)
    }
  }
  return err
}


func (c *XConfig) setparam(line int, key string, typeparam int, value interface{}) error {
  // check if key contains "+" (forced array) and . (subset of config)
  // and creates a Map[] if the value already exists (or just set it)
  //  keydata, merge := analyzeKey(key)
  //  mustmerge := false
  p := newParam()
  p.add(typeparam, value)
  (*c).Parameters[key] = *p
  c.Order = append(c.Order, key)
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
  if len(key) == 0 { return c.addcomment(line, data) }

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
  return c.addparam(line, key, typeparam, value)
}

func (c *XConfig) parsemap(data *XConfig, merge bool) error {
  if (*c).Parameters == nil {
    (*c).Parameters = (*data).Parameters
    (*c).Comments = (*data).Comments
    (*c).Order = (*data).Order
  } else {
    line := len((*c).Order)
    for p, v := range (*data).Parameters {
      if (merge) {
        c.addparam(line, p, v.paramtype, v.Value)
      } else {
        c.setparam(line, p, v.paramtype, v.Value)
      }
    }
    (*c).Multiple = true
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
  for scanner.Scan() {
    line := scanner.Text()
    err := tempConfig.parseline(2, line, merge)
    if (err != nil) {
      return err
    }
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
    if (err != nil) {
      return err
    }
    line += 1
  }

  if err := scanner.Err(); err != nil {
    return err
  }
  
  fmt.Println(tempConfig)
  
  c.parsemap(tempConfig, merge)
  return nil
}

func (c *XConfig) Set(key string, value interface{}) error {
  // check if key contains "+" (forced array) and . (subset of config)
  // and just replace the value
  var valuetype int
  switch value.(type) {
    case string: valuetype = 1
    case int: valuetype = 2
    case float64: valuetype = 3
    case bool: valuetype = 4
  }
  return c.setparam(0, key, valuetype, value)
}

/*
 If the existance of the key entry does matter, use Get. The second parameter is set to false when the entry does not existance
 You will have to cast your result
*/
func (c *XConfig) Get(key string) (interface{}, bool) {
  // check if key contains "." (subset of config)
  if val, ok := (*c).Parameters[key]; ok {
    return val.Value, true
  }
  return nil, false
}

/*
 Get the string value of a string param. If the value is not string or does not exists, return ""
*/
func (c *XConfig) GetString(key string) string {
  // check if key contains "." (subset of config)
  if val, ok := (*c).Parameters[key]; ok {
    switch val.Value.(type) {
      case string: return val.Value.(string)
    }
  }
  return ""
}

/*
 Get the integer value of an int param. If the value is not int or does not exists, return 0
*/
func (c *XConfig) GetInt(key string) int {
  // check if key contains "." (subset of config)
  if val, ok := (*c).Parameters[key]; ok {
    switch val.Value.(type) {
      case int: return val.Value.(int)
    }
  }
  return 0
}

/*
 Get the float value of a float param. If the value is not float or does not exists, return 0
*/
func (c *XConfig) GetFloat(key string) float64 {
  // check if key contains "." (subset of config)
  if val, ok := (*c).Parameters[key]; ok {
    switch val.Value.(type) {
      case float64: return val.Value.(float64)
    }
  }
  return 0
}

/*
 Get the boolean value of a bool. If the value is not bool or does not exists, return false
*/
func (c *XConfig) GetBool(key string) bool {
  // check if key contains "." (subset of config)
  if val, ok := (*c).Parameters[key]; ok {
    switch val.Value.(type) {
      case bool: return val.Value.(bool)
    }
  }
  return false
}

// Accept only string, int, float64 and boolean values
func (c *XConfig) Add(key string, value interface{}) error {
  // check if key contains "+" (forced array) and . (subset of config)
  // and creates a Map[] if the value already exists (or just set it)
  var valuetype int
  switch value.(type) {
    case string: valuetype = 1
    case int: valuetype = 2
    case float64: valuetype = 3
    case bool: valuetype = 4
    default: 
      return errors.New("The XConfig.Add function only accept string, integer, float64 and boolean values")
  }
  return c.addparam(0, key, valuetype, value)
}

func (c *XConfig) LoadFile(filename string) error {
  return c.loadandparse(filename, false)
}

func (c *XConfig) MergeFile(filename string) error {
  return c.loadandparse(filename, true)
}

func (c *XConfig) LoadString(data string) error {
  return c.parsestring(data, false)
}

func (c *XConfig) MergeString(data string) error {
  return c.parsestring(data, true)
}

func (c *XConfig) LoadXConfig(data *XConfig) error {
  return c.parsemap(data, false)
}

func (c *XConfig) MergeXConfig(data *XConfig) error {
  return c.parsemap(data, true)
}











func analyzeKey(key string) (interface{}, bool) {
  mustmerge := false
  if key[len(key)-1] == '+' {
    mustmerge = true
    key = key[0:len(key)-1]
  }
  if key[len(key)-1] == '*' {
    key = key[0:len(key)-1]
  }

  
  match, _ := regexp.MatchString("[a-zA-Z0-9_-]+", key)
  return match, mustmerge

  
}


