package main

import (
  "fmt"
  "testing"
  "github.com/webability-go/xconfig"
)

func TestOneStringParam(t *testing.T) {
  // Test 1: assign a simple parameter string with some comments
  conf := xconfig.New()
  conf.LoadString("#First test\nparam1=value1\n\n;End of test 1\n")
  
  // print what we got
  fmt.Println(conf)

  // direct access
  if (*conf).Parameters["param1"].Value != "value1" {
    t.Errorf("The parameter param1 is not correctly set")
  }
  
  // Get
  if v := conf.Get("param1"); v != "value1" {
    t.Errorf("The parameter param1 is not correctly passed")
  }
}

func TestTwoStringParam(t *testing.T) {
  // Test 2: assign 3 different parameters string
  conf := xconfig.New()
  conf.LoadString("param1=value1\nparam2=value2\nparam3=value3")

  // print what we got
  fmt.Println(conf)

  // direct access
  if (*conf).Parameters["param1"].Value != "value1" || (*conf).Parameters["param2"].Value != "value2"  || (*conf).Parameters["param3"].Value != "value3" {
    t.Errorf("The parameters are not correctly set")
  }
  
  // Get
  v1 := conf.Get("param1")
  v2 := conf.Get("param2")
  v3 := conf.Get("param3")
  if v1 != "value1" || v2 != "value2" || v3 != "value3" {
    t.Errorf("The parameters are not correctly passed")
  }
}

func TestBoolParam(t *testing.T) {
  // Test 3: assign a simple bool
  conf := xconfig.New()
  conf.LoadString("param1=yes\nparam2=true\nparam3=on\nparam4=no\nparam5=none\nparam6=false\nparam7=off")
  
  fmt.Println(conf)
  
  if (*conf).Parameters["param1"].Value != true || (*conf).Parameters["param2"].Value != true || (*conf).Parameters["param3"].Value != true || (*conf).Parameters["param4"].Value != false || (*conf).Parameters["param5"].Value != false || (*conf).Parameters["param6"].Value != false || (*conf).Parameters["param7"].Value != false {
    t.Errorf("The boolean parameters are not correctly set")
  }
}

func TestIntegerParam(t *testing.T) {
  // Test 4: 
  conf := xconfig.New()
  conf.LoadString("param1=0\nparam2=1\nparam3=1234567890")

  fmt.Println(conf)

  if (*conf).Parameters["param1"].Value != 0 || (*conf).Parameters["param2"].Value != 1 || (*conf).Parameters["param3"].Value != 1234567890 {
    t.Errorf("The integer parameters are not correctly set")
  }
}

func TestArrayParam(t *testing.T) {
  // Test 5: 
  conf := xconfig.New()
  conf.LoadString("param1=value1\nparam1=value2\nparam1=value3")

  fmt.Println(conf)
  
//  arr := (*conf).Parameters["param1"].Value

//  if arr.([]string)[0] != "value1" || arr.([]string)[1] != "value2" || arr.([]string)[2] != "value3" {
//    t.Errorf("The array parameter is not correctly set")
//  }
}



