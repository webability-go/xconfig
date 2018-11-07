package xconfig

import (
  "bufio"
  "os"
  "strings"
  "strconv"
)

type XConfig map[string]interface{}

func (c *XConfig) Load(filename string) error {
  return c.loadandparse(filename, false)
}

func (c *XConfig) Merge(filename string) error {
  return c.loadandparse(filename, true)
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

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    line := scanner.Text()
    // we ignore empty and comments lines
    if len(line) == 0 || line[0] == '#' || line[0] == ';' { continue }

    // if there is no =, this is not a key=val line
    posequal := strings.Index(line, "=")
    if posequal < 0 { continue }
    
    // we separate the key. if there is no key, we ignore the line
    key := strings.TrimSpace(line[:posequal])
    if len(key) == 0 { continue }
    
    // we capture the value if it exists. If not, the key entry is initialized with a nil value
    var value interface{}
    if len(line) > posequal {
      strvalue := strings.TrimSpace(line[posequal+1:])
      value = strvalue
      
      if strvalue == "yes" || strvalue == "true" {
        value = true
      }
      if strvalue == "no" || strvalue == "false" {
        value = false
      }
      
      if intvalue, err := strconv.Atoi(strvalue); err == nil {
        value = intvalue
      }
    }
    c.Set(key, value)
  }

  if err := scanner.Err(); err != nil {
    return err
  }
  return nil
}


func (c *XConfig) Set(key string, value interface{}) {
  // check if key contains "+" (forced array) and . (subset of config)

  (*c)[key] = value
}

func (c *XConfig) Get(key string) interface{} {
  // check if key contains "." (subset if config)

  return (*c)[key]
}

