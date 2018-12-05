package main

import (
  "fmt"
  "github.com/webability-go/xconfig"
)

func main() {
  conf := &xconfig.XConfig{}
  conf.LoadFile("./example.conf")
  
  fmt.Println("First file loaded:")
  fmt.Println(conf)
  fmt.Println(conf.Get("language"))
  fmt.Println(conf.Get("language").(*xconfig.XConfig).Get("en"))
  
  conf.MergeFile("./mergeme.conf")

  fmt.Println("After second file merged:")
  fmt.Println(conf)
}