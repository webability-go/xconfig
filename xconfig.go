package xconfig

import (
  "bufio"
  "os"
  "strings"
  "strconv"
)

type XConfig map[string]interface{}

func (c *XConfig) Set(key string, value interface{}) {
  // check if key contains "+" (forced array) and . (subset of config)
  // and just replace the value
  (*c)[key] = value
}

func (c *XConfig) Add(key string, value interface{}) {
  // check if key contains "+" (forced array) and . (subset of config)
  // and creates a Map[] if the value already exists (or just set it)
  (*c)[key] = value
}

func (c *XConfig) Get(key string) interface{} {
  // check if key contains "." (subset if config)

  return (*c)[key]
}

func (c *XConfig) Load(filename string) error {
  return c.loadandparse(filename, false)
}

func (c *XConfig) Merge(filename string) error {
  return c.loadandparse(filename, true)
}

func (c *XConfig) LoadString(data string) error {
  return c.parsestring(data, false)
}

func (c *XConfig) MergeString(data string) error {
  return c.parsestring(data, true)
}

func (c *XConfig) LoadXConfig(data XConfig) error {
  return c.parsemap(data, false)
}

func (c *XConfig) MergeXConfig(data XConfig) error {
  return c.parsemap(data, true)
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
    err := c.parseline(line, merge)
    if (err != nil) {
      return err
    }
  }

  if err := scanner.Err(); err != nil {
    return err
  }
  return nil
}

func (c *XConfig) parsestring(data string, merge bool) error {
  // No filename: we let the config object as is
  if len(data) == 0 {
    return nil
  }
  scanner := bufio.NewScanner(strings.NewReader(data))
  for scanner.Scan() {
    line := scanner.Text()
    err := c.parseline(line, merge)
    if (err != nil) {
      return err
    }
  }

  if err := scanner.Err(); err != nil {
    return err
  }
  return nil
}

func (c *XConfig) parsemap(data XConfig, merge bool) error {
  for p, v := range data {
    c.Set(p, v)
  }
  return nil
}

func (c *XConfig) parseline(line string, merge bool) error {
  // we ignore empty and comments lines
  if len(line) == 0 || line[0] == '#' || line[0] == ';' { return nil }

  // if there is no =, this is not a key=val line
  posequal := strings.Index(line, "=")
  if posequal < 0 { return nil }
  
  // we separate the key. if there is no key, we ignore the line
  key := strings.TrimSpace(line[:posequal])
  if len(key) == 0 { return nil }
  
  // we capture the value if it exists. If not, the key entry is initialized with a nil value
  var value interface{}
  if len(line) > posequal {
    strvalue := strings.TrimSpace(line[posequal+1:])
    value = strvalue
    
    if strvalue == "yes" || strvalue == "true" || strvalue == "on" {
      value = true
    }
    if strvalue == "no" || strvalue == "none" || strvalue == "false" || strvalue == "off" {
      value = false
    }
    
    if intvalue, err := strconv.Atoi(strvalue); err == nil {
      value = intvalue
    }
  }
  c.Set(key, value)
  return nil
}

