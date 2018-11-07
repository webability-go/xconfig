package xconfig

import (
  "fmt"
  "bufio"
  "os"
  "strings"
)

type XConfig map[string]interface{}

func Load(filename string) (*XConfig, error) {
  config := make(XConfig)

  if len(filename) == 0 {
    return &config, nil
  }
  file, err := os.Open(filename)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    line := scanner.Text()
    if len(line) == 0 || line[0] == '#' { continue }
    posequal := strings.Index(line, "=")
    if posequal >= 0 {
      key := strings.TrimSpace(line[:posequal])
      if len(key) > 0 {
        var value interface{}
        if len(line) > posequal {
          value = strings.TrimSpace(line[posequal+1:])
          
          if (value == "yes" || value == "true") {
            value = true
          }
          if (value == "no" || value == "false") {
            value = false
          }
        }
        config.Set(key, value)
      }
    }
  }

  if err := scanner.Err(); err != nil {
    return nil, err
  }

  fmt.Println("CONFIG CREADO DESDE XCONFIG:")
  fmt.Println(config)
  fmt.Printf("%p\n", &config)
  fmt.Printf("%p\n", config["mainpage"])
  
  return &config, nil
}

func (c *XConfig) Set(key string, value interface{}) {
  (*c)[key] = value
}

func (c *XConfig) Get(key string) interface{} {
  return (*c)[key]
}

