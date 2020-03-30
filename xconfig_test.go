package xconfig

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/webability-go/xcore/v2"
)

func TestLoads(t *testing.T) {
	// Test 1: assign a simple parameter string with some comments
	conf0 := New()
	content, err := ioutil.ReadFile("testunit/a.conf")
	if err != nil {
		t.Errorf("Error loading a.conf")
		return
	}
	conf0.LoadString(string(content))

	conf0p := New()
	conf0p.MergeString(string(content))

	conf1 := New()
	conf1.LoadFile("testunit/a.conf")

	conf2 := New()
	conf2.MergeFile("testunit/a.conf") // load is same as merge on first time

	conf3 := New()
	conf3.LoadXConfig(conf1) // load is same as merge on first time

	conf4 := New()
	conf4.MergeXConfig(conf1) // load is same as merge on first time

	// print what we got
	s0 := fmt.Sprint(conf0)
	s0p := fmt.Sprint(conf0p)
	s1 := fmt.Sprint(conf1)
	s2 := fmt.Sprint(conf2)
	s3 := fmt.Sprint(conf3)
	s4 := fmt.Sprint(conf4)
	/*
		fmt.Println(s0)
		fmt.Println(s0p)
		fmt.Println(s1)
		fmt.Println(s2)
		fmt.Println(s3)
		fmt.Println(s4)
	*/

	if s1 != s2 || s1 != s3 || s1 != s4 || s1 != s0 || s1 != s0p {
		t.Errorf("error loading and merging natural files")
	}
}

func TestOneStringParam(t *testing.T) {
	// Test 1: assign a simple parameter string with some comments
	conf := New()
	conf.LoadString("#First test\nparam1=value1\n\n;End of test 1\n")

	// print what we got
	fmt.Println(conf)

	// direct access
	if (*conf).Parameters["param1"].Value != "value1" {
		t.Errorf("The parameter param1 is not correctly set")
	}

	// Get
	if v, _ := conf.Get("param1"); v != "value1" {
		t.Errorf("The parameter param1 is not correctly passed")
	}
}

func TestStringParam(t *testing.T) {
	// Test 2: assign 3 different parameters string
	conf := New()
	conf.LoadString("param1=value1\nparam2=value2\nparam3=value3\nparam4=\"123\nparam5=\"on")

	// print what we got
	fmt.Println(conf)

	// direct access
	if (*conf).Parameters["param1"].Value != "value1" || (*conf).Parameters["param2"].Value != "value2" || (*conf).Parameters["param3"].Value != "value3" {
		t.Errorf("The parameters are not correctly set")
	}
	if (*conf).Parameters["param4"].Value != "123" || (*conf).Parameters["param5"].Value != "on" {
		t.Errorf("The parameters are not correctly set")
	}

	// Get
	v1, _ := conf.Get("param1")
	v2, _ := conf.Get("param2")
	v3, _ := conf.Get("param3")
	v4, _ := conf.Get("param4")
	v5, _ := conf.Get("param5")
	if v1 != "value1" || v2 != "value2" || v3 != "value3" || v4 != "123" || v5 != "on" {
		t.Errorf("The parameters are not correctly passed")
	}
}

func TestBoolParam(t *testing.T) {
	// Test 3: assign a simple bool
	conf := New()
	conf.LoadString("param1=yes\nparam2=true\nparam3=on\nparam4=no\nparam5=none\nparam6=false\nparam7=off")

	fmt.Println(conf)

	if (*conf).Parameters["param1"].Value != true || (*conf).Parameters["param2"].Value != true || (*conf).Parameters["param3"].Value != true || (*conf).Parameters["param4"].Value != false || (*conf).Parameters["param5"].Value != false || (*conf).Parameters["param6"].Value != false || (*conf).Parameters["param7"].Value != false {
		t.Errorf("The boolean parameters are not correctly set")
	}
}

func TestIntegerParam(t *testing.T) {
	// Test 4:
	conf := New()
	conf.LoadString("param1=0\nparam2=-1\nparam3=1234567890")

	fmt.Println(conf)

	if (*conf).Parameters["param1"].Value != 0 || (*conf).Parameters["param2"].Value != -1 || (*conf).Parameters["param3"].Value != 1234567890 {
		t.Errorf("The integer parameters are not correctly set")
	}
}

func TestFloatParam(t *testing.T) {
	// Test 4:
	conf := New()
	conf.LoadString("param1=0.123\nparam2=12e7\nparam3=-76364.2")

	fmt.Println(conf)

	if (*conf).Parameters["param1"].Value != 0.123 || (*conf).Parameters["param2"].Value != 12e7 || (*conf).Parameters["param3"].Value != -76364.2 {
		t.Errorf("The float parameters are not correctly set")
	}
}

func TestArrayParam(t *testing.T) {
	// Test 5:
	conf := New()
	conf.LoadString("param1=value1\nparam1=value2\nparam1=value3\nparam2=123\nparam2=-1\nparam2=1234567890\nparam3=0.1\nparam3=-123.567\nparam3=12e7\nparam4=true\nparam4=off\nparam4=on")

	fmt.Println(conf)

	//  arr := (*conf).Parameters["param1"].Value

	//  if arr.([]string)[0] != "value1" || arr.([]string)[1] != "value2" || arr.([]string)[2] != "value3" {
	//    t.Errorf("The array parameter is not correctly set")
	//  }
}

/* Test injection of a config into a template */
func TestTemplate(t *testing.T) {

	tmpl, _ := xcore.NewXTemplateFromString(`
Some data:
{{param1}}
{{param2}}
{{param3>data1}}
{{param3>data2}}
{{param4}}
{{param5}}
`)

	conf := New()
	conf.LoadString("param1=value1\nparam2=value2\nparam3.data1=value3-data1\nparam3.data2=value3-data2\nparam4=\"123\nparam5=\"on")

	fmt.Println(conf)

	result := tmpl.Execute(conf)
	fmt.Println("Result: ", result)

}

func TestClone(t *testing.T) {
	// Test 1: assign a simple parameter string with some comments
	conf := New()
	conf.LoadString("#First test\nparam1=value1\n\n;End of test 1\n")

	conf2 := conf.Clone()
	conf.Set("param10", "value10")

	// print what we got
	fmt.Println("ANTES DE CLONE", conf)
	fmt.Println("OBJETO CLONED", conf2)

}
