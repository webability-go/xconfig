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

/* Basic parameter. 
   The type of the value can be 0 = not set, 1 = string, 2 = int, 3 = bool, 4 = array of strings, 5 = array of int, 6 = XConfig
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
    case 0: 
      p.paramtype = paramtype
      p.Value = value
    case 1:
      if paramtype == 1 {
        // transform the parameter into an array and change paramtype
        sub := make([]string, 0, 2)
        p.Value = append(sub, p.Value.(string), value.(string))
        p.paramtype = 4
      } else if paramtype == 4 {
        // concatenate array of string
        p.Value = append(value.([]string), p.Value.(string))
      } else {
        return errors.New("The parameter cannot add an incompatible value to a string")
      }
    case 2:
      if paramtype == 2 {
        // transform the parameter into an array and change paramtype
        sub := make([]int, 0, 2)
        p.Value = append(sub, p.Value.(int), value.(int))
        p.paramtype = 5
      } else if paramtype == 5 {
        // concatenate array of int
        p.Value = append(value.([]int), p.Value.(int))
      } else {
        return errors.New("The parameter cannot add an incompatible value to an integer")
      }
    case 3:
      if paramtype == 3 {
        p.Value = value
      } else {
        return errors.New("The parameter cannot add an incompatible value to a boolean")
      }
    case 4:
      if paramtype == 1 {
        p.Value = append(p.Value.([]string), value.(string))
      } else if paramtype == 4 {
        // concatenate array of string
        p.Value = append(p.Value.([]string), value.([]string)...)
      } else {
        return errors.New("The parameter cannot add an incompatible value to an array of strings")
      }
    case 5:
      if paramtype == 2 {
        p.Value = append(p.Value.([]int), value.(int))
      } else if paramtype == 5 {
        // concatenate array of int
        p.Value = append(p.Value.([]int), value.([]int)...)
      } else {
        return errors.New("The parameter cannot add an incompatible value to an array of integers")
      }
    default:
      return errors.New("The parameter cannot add an incompatible value to a sub XConfig")
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
  return err
}


func (c *XConfig) setparam(line int, key string, typeparam int, value interface{}) {
  // check if key contains "+" (forced array) and . (subset of config)
  // and creates a Map[] if the value already exists (or just set it)
  //  keydata, merge := analyzeKey(key)
  //  mustmerge := false
  p := newParam()
  p.add(typeparam, value)
  (*c).Parameters[key] = *p
  c.Order = append(c.Order, key)
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
    
    if strvalue == "yes" || strvalue == "true" || strvalue == "on" {
      value = true
      typeparam = 3
    }
    if strvalue == "no" || strvalue == "none" || strvalue == "false" || strvalue == "off" {
      value = false
      typeparam = 3
    }
    
    if intvalue, err := strconv.Atoi(strvalue); err == nil {
      value = intvalue
      typeparam = 2
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

func (c *XConfig) Set(key string, value interface{}) error {
  // check if key contains "+" (forced array) and . (subset of config)
  // and just replace the value
/*

  mustmerge := false
  if key[len(key)-1] == '+' {
    mustmerge = true
    key = key[0:len(key)-2]
  }
  if key[len(key)-1] == '*' {
    key = key[0:len(key)-2]
  }

  if mustmerge {
    (*c)[key] = value
  } else {
    (*c)[key] = value
  }
  */
  return nil
}



func (c *XConfig) Add(key string, value interface{}) error {
  // check if key contains "+" (forced array) and . (subset of config)
  // and creates a Map[] if the value already exists (or just set it)
//  keydata, merge := analyzeKey(key)
//  mustmerge := false

  /*
  if val, ok := (*c)[key]; ok {
    switch reflect.TypeOf(val).Kind() {
      case reflect.Slice:
        (*c)[key] = append(val.([]string), value.(string))
      default:
        sub := make([]string, 0, 2)
        sub = append(sub, val.(string), value.(string))
        (*c)[key] = sub
    }
  } else {
    fmt.Println("Assign " + key)
    (*c)[key] = value
    fmt.Println(*c)
  }
  */
/*
  if mustmerge {
    (*c)[key] = value
  } else {
    (*c)[key] = value
  }
  (*c)[key] = value
*/




  return nil
}

func (c *XConfig) Get(key string) interface{} {
  // check if key contains "." (subset if config)
  if val, ok := (*c).Parameters[key]; ok {
    return val.Value
  }
  return nil
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
