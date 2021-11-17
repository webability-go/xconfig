package xconfig

import (
	"fmt"
	"io/ioutil"
	"testing"
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
	//	fmt.Println(conf)

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
	//	fmt.Println(conf)

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

	//	fmt.Println(conf)

	if (*conf).Parameters["param1"].Value != true || (*conf).Parameters["param2"].Value != true || (*conf).Parameters["param3"].Value != true || (*conf).Parameters["param4"].Value != false || (*conf).Parameters["param5"].Value != false || (*conf).Parameters["param6"].Value != false || (*conf).Parameters["param7"].Value != false {
		t.Errorf("The boolean parameters are not correctly set")
	}
}

func TestIntegerParam(t *testing.T) {
	// Test 4:
	conf := New()
	conf.LoadString("param1=0\nparam2=-1\nparam3=1234567890")

	//	fmt.Println(conf)

	if (*conf).Parameters["param1"].Value != 0 || (*conf).Parameters["param2"].Value != -1 || (*conf).Parameters["param3"].Value != 1234567890 {
		t.Errorf("The integer parameters are not correctly set")
	}
}

func TestFloatParam(t *testing.T) {
	// Test 4:
	conf := New()
	conf.LoadString("param1=0.123\nparam2=12e7\nparam3=-76364.2")

	//	fmt.Println(conf)

	if (*conf).Parameters["param1"].Value != 0.123 || (*conf).Parameters["param2"].Value != 12e7 || (*conf).Parameters["param3"].Value != -76364.2 {
		t.Errorf("The float parameters are not correctly set")
	}
}

func TestArrayParam(t *testing.T) {
	// Test 5:
	conf := New()
	conf.LoadString("param1=value1\nparam1=value2\nparam1=value3\nparam2=123\nparam2=-1\nparam2=1234567890\nparam3=0.1\nparam3=-123.567\nparam3=12e7\nparam4=true\nparam4=off\nparam4=on")

	//	fmt.Println(conf)

	arr, ext := conf.GetStringCollection("param1")
	if !ext {
		t.Errorf("The array parameter is not correctly set")
		return
	}
	if arr[0] != "value1" || arr[1] != "value2" || arr[2] != "value3" {
		t.Errorf("The array parameter is not correctly set")
	}
}

func TestClone(t *testing.T) {
	// Test 1: assign a simple parameter string with some comments
	conf := New()
	conf.LoadString("#First test\nparam1=value1\n\n;End of test 1\n")

	conf2 := conf.Clone()

	if fmt.Sprint(conf) != fmt.Sprint(conf2) {
		t.Errorf("Error cloning the xconfig")
	}

	// print what we got
	//	fmt.Println("ANTES DE CLONE", conf)
	//	fmt.Println("OBJETO CLONED", conf2)

	conf.Set("param10", "value10")
	if fmt.Sprint(conf) == fmt.Sprint(conf2) {
		t.Errorf("Error cloning the xconfig")
	}

}

func TestStructure(t *testing.T) {
	conf := New()
	err := conf.LoadFile("testunit/example.conf")
	if err != nil {
		t.Error(err)
		return
	}

	s0 := conf.Marshal()
	r0 := `# this file is named myconfig.conf, used in following examples
# the # denotes a comment.
; is also a comment
parameter1=value1
parameter2=value2
parameter2=value3

# global config:
ip=127.0.0.1
port=80
domain=test.com

# Some list of values, they will result into an array
country=MX
country=US
country=FR
country=JP
country=ES

# some subsets
language.en.welcome=Welcome to the XConfig examples
language.en.ack=OK
language.en.cancel=Cancel
language.es.welcome=Bienvenido a los ejemplos de XConfig
language.es.ack=Perfecto
language.es.cancel=Cancelar
# spanish
`

	if s0 != r0 {
		t.Errorf("Error marshelling file, considering pushing the comments into an array of values")
	}
}

func TestDel(t *testing.T) {
	conf := New()
	conf.LoadString("param1=0.123\nparam2=12e7\nparam3=-76364.2")
	conf.Del("param1")
	s0 := fmt.Sprint(conf)
	if s0 != "XConfig[\nparam2:1.2e+08\nparam3:-76364.2\n]\n" {
		t.Errorf("The parameter has not been correctly deleted")
	}
}
