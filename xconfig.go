package xconfig

import (
  "bufio"
  "os"
  "strings"
)

type XConfig map[string]interface{}

func Load(filename string) (*XConfig, error) {
  config := &XConfig{}

  if len(filename) == 0 {
    return config, nil
  }
  file, err := os.Open(filename)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    line := scanner.Text()
    if line[0] == '#' { continue }
    posequal := strings.Index(line, "=")
    if posequal >= 0 {
      key := strings.TrimSpace(line[:posequal])
      if len(key) > 0 {
        value := ""
        if len(line) > posequal {
          value = strings.TrimSpace(line[posequal+1:])
        }
        (*config)[key] = value
      }
    }
  }

  if err := scanner.Err(); err != nil {
    return nil, err
  }

  return config, nil
}
