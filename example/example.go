package main

import (
  "fmt"
  "github.com/webability-go/xconfig"
)

func main() {
  conf := &xconfig.XConfig{}
  conf.Load("./example.conf")
  
  fmt.Println("First file loaded:")
  fmt.Println(conf)
  
  conf.Merge("./mergeme.conf")

  fmt.Println("After second file merged:")
  fmt.Println(conf)
}