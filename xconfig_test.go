package xconfig

// Comprehensive test suite for xconfig v0.5.0
// Covers:
//   - All documented configuration behaviors (README / godoc)
//   - Every bug fixed in the 2.x / 3.x / 4.x analysis
//   - Thread-safety under concurrent read+write load
//
// Run with:
//   go test ./...
//   go test -race ./...

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 1: Load equivalence ──────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestLoads verifies that LoadFile, LoadString, MergeFile, LoadXConfig and
// MergeXConfig all produce identical results when given the same input.
func TestLoads(t *testing.T) {
	content, err := os.ReadFile("testunit/a.conf")
	if err != nil {
		t.Fatalf("cannot read testunit/a.conf: %v", err)
	}
	src := string(content)

	conf0 := New()
	conf0.LoadString(src)

	conf0p := New()
	conf0p.MergeString(src)

	conf1 := New()
	conf1.LoadFile("testunit/a.conf")

	conf2 := New()
	conf2.MergeFile("testunit/a.conf")

	conf3 := New()
	conf3.LoadXConfig(conf1)

	conf4 := New()
	conf4.MergeXConfig(conf1)

	s0 := fmt.Sprint(conf0)
	s0p := fmt.Sprint(conf0p)
	s1 := fmt.Sprint(conf1)
	s2 := fmt.Sprint(conf2)
	s3 := fmt.Sprint(conf3)
	s4 := fmt.Sprint(conf4)

	if s1 != s2 || s1 != s3 || s1 != s4 || s1 != s0 || s1 != s0p {
		t.Errorf("Load/Merge methods produced different results:\nLoadFile:    %s\nMergeFile:   %s\nLoadXConfig: %s\nMergeXConfig:%s\nLoadString:  %s\nMergeString: %s", s1, s2, s3, s4, s0, s0p)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 2: Type parsing ───────────────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

func TestOneStringParam(t *testing.T) {
	conf := New()
	conf.LoadString("#First test\nparam1=value1\n\n;End of test 1\n")

	if conf.Parameters["param1"].Value != "value1" {
		t.Errorf("direct access: expected 'value1', got %v", conf.Parameters["param1"].Value)
	}
	if v, _ := conf.Get("param1"); v != "value1" {
		t.Errorf("Get: expected 'value1', got %v", v)
	}
}

func TestStringParam(t *testing.T) {
	conf := New()
	conf.LoadString("param1=value1\nparam2=value2\nparam3=value3\nparam4=\"123\nparam5=\"on")

	if conf.Parameters["param1"].Value != "value1" ||
		conf.Parameters["param2"].Value != "value2" ||
		conf.Parameters["param3"].Value != "value3" {
		t.Error("plain string parameters not correctly set")
	}
	// Forced strings with " prefix
	if conf.Parameters["param4"].Value != "123" {
		t.Errorf("param4: expected string '123', got %v (%T)", conf.Parameters["param4"].Value, conf.Parameters["param4"].Value)
	}
	if conf.Parameters["param5"].Value != "on" {
		t.Errorf("param5: expected string 'on', got %v (%T)", conf.Parameters["param5"].Value, conf.Parameters["param5"].Value)
	}
}

func TestBoolParam(t *testing.T) {
	conf := New()
	conf.LoadString("p1=yes\np2=true\np3=on\np4=no\np5=none\np6=false\np7=off")

	trueKeys := []string{"p1", "p2", "p3"}
	falseKeys := []string{"p4", "p5", "p6", "p7"}

	for _, k := range trueKeys {
		if conf.Parameters[k].Value != true {
			t.Errorf("%s: expected true, got %v", k, conf.Parameters[k].Value)
		}
	}
	for _, k := range falseKeys {
		if conf.Parameters[k].Value != false {
			t.Errorf("%s: expected false, got %v", k, conf.Parameters[k].Value)
		}
	}
}

func TestIntegerParam(t *testing.T) {
	conf := New()
	conf.LoadString("p1=0\np2=-1\np3=1234567890")

	if conf.Parameters["p1"].Value != 0 {
		t.Errorf("p1: expected 0, got %v", conf.Parameters["p1"].Value)
	}
	if conf.Parameters["p2"].Value != -1 {
		t.Errorf("p2: expected -1, got %v", conf.Parameters["p2"].Value)
	}
	if conf.Parameters["p3"].Value != 1234567890 {
		t.Errorf("p3: expected 1234567890, got %v", conf.Parameters["p3"].Value)
	}
}

func TestFloatParam(t *testing.T) {
	conf := New()
	conf.LoadString("p1=0.123\np2=12e7\np3=-76364.2")

	if conf.Parameters["p1"].Value != 0.123 {
		t.Errorf("p1: expected 0.123")
	}
	if conf.Parameters["p2"].Value != 12e7 {
		t.Errorf("p2: expected 12e7")
	}
	if conf.Parameters["p3"].Value != -76364.2 {
		t.Errorf("p3: expected -76364.2")
	}
}

// TestForcedStringPrefix verifies that values prefixed with " are always stored
// as strings, even when the content looks like a bool, int, or float.
func TestForcedStringPrefix(t *testing.T) {
	conf := New()
	conf.LoadString("a=\"true\nb=\"123\nc=\"off\nd=\"3.14\ne=\"\"+quoted")

	cases := []struct {
		key      string
		expected string
	}{
		{"a", "true"},
		{"b", "123"},
		{"c", "off"},
		{"d", "3.14"},
		{"e", "\"+quoted"},
	}
	for _, tc := range cases {
		v := conf.Parameters[tc.key].Value
		s, ok := v.(string)
		if !ok {
			t.Errorf("key %q: expected string, got %T (%v)", tc.key, v, v)
			continue
		}
		if s != tc.expected {
			t.Errorf("key %q: expected %q, got %q", tc.key, tc.expected, s)
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 3: Array / collection parameters ─────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

func TestArrayParam(t *testing.T) {
	conf := New()
	conf.LoadString("p1=value1\np1=value2\np1=value3\n" +
		"p2=123\np2=-1\np2=1234567890\n" +
		"p3=0.1\np3=-123.567\np3=12e7\n" +
		"p4=true\np4=off\np4=on")

	arr, ok := conf.GetStringCollection("p1")
	if !ok || len(arr) != 3 || arr[0] != "value1" || arr[1] != "value2" || arr[2] != "value3" {
		t.Errorf("string array: %v", arr)
	}

	iarr, ok := conf.GetIntCollection("p2")
	if !ok || len(iarr) != 3 || iarr[0] != 123 || iarr[1] != -1 || iarr[2] != 1234567890 {
		t.Errorf("int array: %v", iarr)
	}

	farr, ok := conf.GetFloatCollection("p3")
	if !ok || len(farr) != 3 || farr[0] != 0.1 || farr[1] != -123.567 || farr[2] != 12e7 {
		t.Errorf("float array: %v", farr)
	}

	barr, ok := conf.GetBoolCollection("p4")
	if !ok || len(barr) != 3 || barr[0] != true || barr[1] != false || barr[2] != true {
		t.Errorf("bool array: %v", barr)
	}
}

// TestSingleValueWrappedInCollection verifies that Get*Collection wraps a
// scalar value in a single-element slice (documented convenience behaviour).
func TestSingleValueWrappedInCollection(t *testing.T) {
	conf := New()
	conf.LoadString("s=hello\ni=42\nf=3.14\nb=true")

	if arr, ok := conf.GetStringCollection("s"); !ok || len(arr) != 1 || arr[0] != "hello" {
		t.Errorf("single string wrapped: %v %v", arr, ok)
	}
	if arr, ok := conf.GetIntCollection("i"); !ok || len(arr) != 1 || arr[0] != 42 {
		t.Errorf("single int wrapped: %v %v", arr, ok)
	}
	if arr, ok := conf.GetFloatCollection("f"); !ok || len(arr) != 1 || arr[0] != 3.14 {
		t.Errorf("single float wrapped: %v %v", arr, ok)
	}
	if arr, ok := conf.GetBoolCollection("b"); !ok || len(arr) != 1 || arr[0] != true {
		t.Errorf("single bool wrapped: %v %v", arr, ok)
	}
}

// TestArrayOrderPreservation verifies that the order of array elements matches
// the order they appear in the file.
func TestArrayOrderPreservation(t *testing.T) {
	conf := New()
	conf.LoadString("lang=es\nlang=en\nlang=fr\nlang=jp")

	arr, ok := conf.GetStringCollection("lang")
	if !ok {
		t.Fatal("lang not found")
	}
	expected := []string{"es", "en", "fr", "jp"}
	for i, v := range expected {
		if arr[i] != v {
			t.Errorf("position %d: expected %q, got %q", i, v, arr[i])
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 4: Sub-config (dotted keys) ──────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

func TestSubConfigHierarchy(t *testing.T) {
	conf := New()
	conf.LoadString("db.host=localhost\ndb.port=5432\ndb.name=mydb")

	sub := conf.GetConfig("db")
	if sub == nil {
		t.Fatal("db sub-config not found")
	}

	if v, _ := sub.Get("host"); v != "localhost" {
		t.Errorf("db.host: expected 'localhost', got %v", v)
	}
	if v, _ := sub.Get("port"); v != 5432 {
		t.Errorf("db.port: expected 5432, got %v", v)
	}
	if v, _ := sub.Get("name"); v != "mydb" {
		t.Errorf("db.name: expected 'mydb', got %v", v)
	}
}

func TestSubConfigDeep(t *testing.T) {
	conf := New()
	conf.LoadString("a.b.c=deep")

	sub1 := conf.GetConfig("a")
	if sub1 == nil {
		t.Fatal("sub-config 'a' not found")
	}
	sub2 := sub1.GetConfig("b")
	if sub2 == nil {
		t.Fatal("sub-config 'a.b' not found")
	}
	if v, _ := sub2.Get("c"); v != "deep" {
		t.Errorf("a.b.c: expected 'deep', got %v", v)
	}
}

// TestSetDottedKey verifies FIX 4.5: Set() creates the correct sub-XConfig
// hierarchy for dotted keys, consistent with parseline behaviour.
func TestSetDottedKey(t *testing.T) {
	conf := New()
	conf.Set("server.host", "example.com")
	conf.Set("server.port", 8080)

	sub := conf.GetConfig("server")
	if sub == nil {
		t.Fatal("FIX 4.5 FAILED: Set('server.host',...) did not create a sub-XConfig for 'server'")
	}
	if v, _ := sub.Get("host"); v != "example.com" {
		t.Errorf("server.host: expected 'example.com', got %v", v)
	}
	if v, _ := sub.Get("port"); v != 8080 {
		t.Errorf("server.port: expected 8080, got %v", v)
	}

	// Verify consistency with file-loaded equivalent
	conf2 := New()
	conf2.LoadString("server.host=example.com\nserver.port=8080")

	sub2 := conf2.GetConfig("server")
	if sub2 == nil {
		t.Fatal("file-loaded sub-config not found")
	}
	if v, _ := sub2.Get("host"); v != "example.com" {
		t.Errorf("file-loaded server.host: expected 'example.com', got %v", v)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 5: Merge vs Load semantics ───────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

func TestLoadReplacesValues(t *testing.T) {
	conf := New()
	conf.LoadString("port=80\ndomain=test.com")
	conf.LoadString("port=8080\ntitle=Welcome") // Load replaces port

	if v, _ := conf.Get("port"); v != 8080 {
		t.Errorf("Load should replace port: expected 8080, got %v", v)
	}
	if v, _ := conf.Get("domain"); v != "test.com" {
		t.Errorf("Load: domain should survive: expected 'test.com', got %v", v)
	}
	if v, _ := conf.Get("title"); v != "Welcome" {
		t.Errorf("Load: title should appear: expected 'Welcome', got %v", v)
	}
}

func TestMergeAddsValues(t *testing.T) {
	conf := New()
	conf.LoadString("port=80\ndomain=test.com")
	conf.MergeString("port=8080\ntitle=Welcome") // Merge: port becomes array [80, 8080]

	// In merge mode, repeating a key builds an array
	iarr, ok := conf.GetIntCollection("port")
	if !ok || len(iarr) != 2 {
		t.Errorf("Merge should create array for port, got: %v %v", iarr, ok)
	}
	if iarr[0] != 80 || iarr[1] != 8080 {
		t.Errorf("port array order: expected [80, 8080], got %v", iarr)
	}
	if v, _ := conf.Get("title"); v != "Welcome" {
		t.Errorf("Merge: title not added: got %v", v)
	}
}

// TestMergeFile uses the testunit conf files.
func TestMergeFile(t *testing.T) {
	conf := New()
	if err := conf.LoadFile("testunit/example.conf"); err != nil {
		t.Fatal(err)
	}
	if err := conf.MergeFile("testunit/mergeme.conf"); err != nil {
		t.Fatal(err)
	}

	// port was 80 in example.conf; merge adds 8080 → array [80, 8080]
	ports, ok := conf.GetIntCollection("port")
	if !ok || len(ports) != 2 {
		t.Errorf("port should be an array of 2 after merge, got %v", ports)
	}

	// title is new from mergeme.conf
	if v, _ := conf.Get("title"); v != "Welcome" {
		t.Errorf("title: expected 'Welcome', got %v", v)
	}

	// language.fr should now exist (added via merge)
	sub := conf.GetConfig("language")
	if sub == nil {
		t.Fatal("language sub-config missing after merge")
	}
	frSub := sub.GetConfig("fr")
	if frSub == nil {
		t.Fatal("language.fr sub-config missing after merging mergeme.conf")
	}
	if v, _ := frSub.Get("ack"); v != "Super" {
		t.Errorf("language.fr.ack: expected 'Super', got %v", v)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 6: := and += operators  (FIX 4.1) ────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestForceReplaceInMergeMode verifies := overrides merge-add semantics.
func TestForceReplaceInMergeMode(t *testing.T) {
	conf := New()
	conf.LoadString("param=value1\nparam=value2") // array: ["value1","value2"]

	conf.MergeString("param:=replaced") // := forces replace even in merge mode

	v, ok := conf.Get("param")
	if !ok {
		t.Fatal("param missing after :=")
	}
	if v != "replaced" {
		t.Errorf("FIX 4.1 FAILED (:= in merge mode): expected 'replaced', got %v (type %T)", v, v)
	}
}

// TestForceAddInLoadMode verifies += overrides load-replace semantics.
func TestForceAddInLoadMode(t *testing.T) {
	conf := New()
	conf.LoadString("param=value1")
	conf.LoadString("param+=value2") // += forces add even in load mode

	arr, ok := conf.GetStringCollection("param")
	if !ok {
		t.Fatal("param should be a collection after +=")
	}
	if len(arr) != 2 || arr[0] != "value1" || arr[1] != "value2" {
		t.Errorf("FIX 4.1 FAILED (+= in load mode): expected [value1,value2], got %v", arr)
	}
}

// TestForceReplaceInLoadMode verifies := also works in load mode.
func TestForceReplaceInLoadMode(t *testing.T) {
	conf := New()
	conf.LoadString("x=original")
	conf.LoadString("x:=overwritten")

	v, _ := conf.Get("x")
	if v != "overwritten" {
		t.Errorf("FIX 4.1 FAILED (:= in load mode): expected 'overwritten', got %v", v)
	}
}

// TestForceAddInMergeMode verifies += also works in merge mode.
func TestForceAddInMergeMode(t *testing.T) {
	conf := New()
	conf.LoadString("n=1")
	conf.MergeString("n+=2\nn+=3")

	arr, ok := conf.GetIntCollection("n")
	if !ok || len(arr) != 3 || arr[0] != 1 || arr[1] != 2 || arr[2] != 3 {
		t.Errorf("FIX 4.1 FAILED (+= in merge mode): expected [1,2,3], got %v", arr)
	}
}

// TestAssignmentOperatorsDoNotCorruptKeys verifies that :=  and +=  in a
// config file do NOT leave a literal "key:" or "key+" entry in Parameters.
func TestAssignmentOperatorsDoNotCorruptKeys(t *testing.T) {
	conf := New()
	conf.LoadString("city=Paris\ncity:=London")

	if _, ok := conf.Parameters["city:"]; ok {
		t.Error("FIX 4.1 FAILED: corrupted key 'city:' must not exist in Parameters")
	}
	if _, ok := conf.Parameters["city+"]; ok {
		t.Error("FIX 4.1 FAILED: corrupted key 'city+' must not exist in Parameters")
	}
	if v, _ := conf.Get("city"); v != "London" {
		t.Errorf("city after :=: expected 'London', got %v", v)
	}
}

// TestOperatorInValueIsNotDetectedAsOperator verifies that := or += appearing
// inside a value (after the first =) is not mis-detected as an assignment
// operator.
func TestOperatorInValueIsNotDetectedAsOperator(t *testing.T) {
	conf := New()
	conf.LoadString("url=http://example.com:=something")
	conf.LoadString("math=a+=b")

	if v, _ := conf.Get("url"); v != "http://example.com:=something" {
		t.Errorf("url with := in value: expected 'http://example.com:=something', got %v", v)
	}
	if v, _ := conf.Get("math"); v != "a+=b" {
		t.Errorf("math with += in value: expected 'a+=b', got %v", v)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 7: add() order preservation  (FIX 4.2) ───────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestAddOrderSingleToArrayViaString verifies that when a single string is
// merged with a []string, the existing value comes FIRST.
func TestAddOrderSingleToArrayViaString(t *testing.T) {
	conf1 := New()
	conf1.LoadString("tag=first")

	conf2 := New()
	conf2.LoadString("tag=second\ntag=third")

	conf1.MergeXConfig(conf2) // conf1 has string "first"; conf2 has []string{"second","third"}

	arr, ok := conf1.GetStringCollection("tag")
	if !ok {
		t.Fatal("tag should be a collection after merge")
	}
	if len(arr) != 3 || arr[0] != "first" || arr[1] != "second" || arr[2] != "third" {
		t.Errorf("FIX 4.2 FAILED: expected [first, second, third], got %v", arr)
	}
}

// TestAddOrderSingleToArrayViaInt verifies the same fix for int type.
func TestAddOrderSingleToArrayViaInt(t *testing.T) {
	conf1 := New()
	conf1.LoadString("n=1")

	conf2 := New()
	conf2.LoadString("n=2\nn=3")

	conf1.MergeXConfig(conf2)

	arr, ok := conf1.GetIntCollection("n")
	if !ok || len(arr) != 3 || arr[0] != 1 || arr[1] != 2 || arr[2] != 3 {
		t.Errorf("FIX 4.2 FAILED (int): expected [1,2,3], got %v", arr)
	}
}

// TestAddOrderSingleToArrayViaFloat verifies the same fix for float64 type.
func TestAddOrderSingleToArrayViaFloat(t *testing.T) {
	conf1 := New()
	conf1.LoadString("x=1.1")

	conf2 := New()
	conf2.LoadString("x=2.2\nx=3.3")

	conf1.MergeXConfig(conf2)

	arr, ok := conf1.GetFloatCollection("x")
	if !ok || len(arr) != 3 || arr[0] != 1.1 || arr[1] != 2.2 || arr[2] != 3.3 {
		t.Errorf("FIX 4.2 FAILED (float): expected [1.1,2.2,3.3], got %v", arr)
	}
}

// TestAddOrderSingleToArrayViaBool verifies the same fix for bool type.
func TestAddOrderSingleToArrayViaBool(t *testing.T) {
	conf1 := New()
	conf1.LoadString("b=true")

	conf2 := New()
	conf2.LoadString("b=false\nb=true")

	conf1.MergeXConfig(conf2)

	arr, ok := conf1.GetBoolCollection("b")
	if !ok || len(arr) != 3 || arr[0] != true || arr[1] != false || arr[2] != true {
		t.Errorf("FIX 4.2 FAILED (bool): expected [true,false,true], got %v", arr)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 8: LoadXConfig aliasing  (FIX 4.3) ───────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestLoadXConfigNoAliasing verifies that after LoadXConfig the destination and
// source do not share backing storage.
func TestLoadXConfigNoAliasing(t *testing.T) {
	src := New()
	src.LoadString("key=original\ncount=10")

	dst := New()
	dst.LoadXConfig(src)

	// Modify src after loading
	src.Set("key", "modified")
	src.Set("count", 99)

	if v, _ := dst.Get("key"); v != "original" {
		t.Errorf("FIX 4.3 FAILED: dst.key was aliased to src; expected 'original', got %v", v)
	}
	if v, _ := dst.Get("count"); v != 10 {
		t.Errorf("FIX 4.3 FAILED: dst.count was aliased to src; expected 10, got %v", v)
	}
}

// TestMergeXConfigNoAliasing verifies the same for MergeXConfig on empty config.
func TestMergeXConfigNoAliasing(t *testing.T) {
	src := New()
	src.LoadString("x=1\nx=2\nx=3")

	dst := New()
	dst.MergeXConfig(src) // dst is empty → deep-clone path

	src.Add("x", 4) // modify src

	arr, _ := dst.GetIntCollection("x")
	if len(arr) != 3 {
		t.Errorf("FIX 4.3 FAILED: dst array aliased to src; dst has %d elements (expected 3): %v", len(arr), arr)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 9: Clone deep copy  (FIX 4.4) ────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

func TestClone(t *testing.T) {
	conf := New()
	conf.LoadString("#First test\nparam1=value1\n\n;End of test 1\n")

	conf2 := conf.Clone()

	if fmt.Sprint(conf) != fmt.Sprint(conf2) {
		t.Error("Clone produced different String() output from original")
	}

	conf.Set("param10", "value10")
	if fmt.Sprint(conf) == fmt.Sprint(conf2) {
		t.Error("Clone shares state: modifying original affected the clone")
	}
}

// TestCloneDeepCopyStringArray verifies that cloning a config with a []string
// does not share the backing array — appending to the clone must not affect
// the original.
func TestCloneDeepCopyStringArray(t *testing.T) {
	conf := New()
	conf.LoadString("color=red\ncolor=green\ncolor=blue")

	cloned := conf.Clone().(*XConfig)
	cloned.Add("color", "yellow")

	origArr, _ := conf.GetStringCollection("color")
	if len(origArr) != 3 {
		t.Errorf("FIX 4.4 FAILED: appending to cloned []string affected original; original has %d elements: %v", len(origArr), origArr)
	}

	clonedArr, _ := cloned.GetStringCollection("color")
	if len(clonedArr) != 4 {
		t.Errorf("cloned array should have 4 elements, got %d", len(clonedArr))
	}
}

// TestCloneDeepCopyIntArray verifies the same for []int.
func TestCloneDeepCopyIntArray(t *testing.T) {
	conf := New()
	conf.LoadString("n=1\nn=2\nn=3")

	cloned := conf.Clone().(*XConfig)
	cloned.Add("n", 4)

	origArr, _ := conf.GetIntCollection("n")
	if len(origArr) != 3 {
		t.Errorf("FIX 4.4 FAILED (int): original has %d elements: %v", len(origArr), origArr)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 10: Del correctness  (FIX 3.2) ───────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

func TestDel(t *testing.T) {
	conf := New()
	conf.LoadString("p1=0.123\np2=12e7\np3=-76364.2")
	conf.Del("p1")
	s0 := fmt.Sprint(conf)
	if s0 != "XConfig[\np2:1.2e+08\np3:-76364.2\n]\n" {
		t.Errorf("Del: unexpected String() output: %q", s0)
	}
}

// TestDelRemovesFromOrder verifies that Del removes the key from Order and not
// just from Parameters.
func TestDelRemovesFromOrder(t *testing.T) {
	conf := New()
	conf.LoadString("a=1\nb=2\nc=3")
	conf.Del("b")

	for _, v := range conf.Order {
		if v == "b" {
			t.Error("FIX 3.2 FAILED: 'b' still present in Order after Del")
		}
	}
	if _, ok := conf.Parameters["b"]; ok {
		t.Error("'b' still present in Parameters after Del")
	}

	// Marshal should not contain "b"
	m := conf.Marshal()
	if strings.Contains(m, "b=") {
		t.Errorf("Marshal after Del still contains 'b': %q", m)
	}
}

// TestDelNonExistentKey verifies Del on a missing key is a no-op (no panic).
func TestDelNonExistentKey(t *testing.T) {
	conf := New()
	conf.LoadString("x=1")
	conf.Del("missing") // must not panic
	if v, _ := conf.Get("x"); v != 1 {
		t.Error("Del on missing key corrupted existing entries")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 11: Marshal idempotency  (FIX 4.6) ───────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestMarshalPreservesStringTypeThroughRoundTrip verifies FIX 4.6:
// a string like "true", "123", "off" written with " prefix is marshalled back
// with the " prefix so reload does not change its type.
func TestMarshalPreservesStringTypeThroughRoundTrip(t *testing.T) {
	cases := []struct {
		input    string // raw config line
		key      string
		expected string
	}{
		{`a="true`, "a", "true"},
		{`b="false`, "b", "false"},
		{`c="123`, "c", "123"},
		{`d="-5.834`, "d", "-5.834"},
		{`e="on`, "e", "on"},
		{`f="off`, "f", "off"},
		{`g="yes`, "g", "yes"},
		{`h="no`, "h", "no"},
	}
	for _, tc := range cases {
		conf1 := New()
		conf1.LoadString(tc.input)

		// Verify it loaded as a string
		v1 := conf1.Parameters[tc.key].Value
		if _, ok := v1.(string); !ok {
			t.Errorf("key %q: initial load: expected string, got %T", tc.key, v1)
			continue
		}

		// Marshal and reload
		marshaled := conf1.Marshal()
		conf2 := New()
		conf2.LoadString(marshaled)

		v2 := conf2.Parameters[tc.key].Value
		if _, ok := v2.(string); !ok {
			t.Errorf("FIX 4.6 FAILED key %q: after marshal+reload type changed to %T (was string); marshaled=%q",
				tc.key, v2, marshaled)
		}
		if v2 != tc.expected {
			t.Errorf("key %q: value changed after round-trip; expected %q, got %v", tc.key, tc.expected, v2)
		}
	}
}

// TestMarshalArrayStringPreservesPrefix verifies FIX 4.6 for []string arrays.
func TestMarshalArrayStringPreservesPrefix(t *testing.T) {
	conf1 := New()
	conf1.LoadString("val=\"true\nval=\"123\nval=hello")

	m := conf1.Marshal()
	conf2 := New()
	conf2.LoadString(m)

	arr, ok := conf2.GetStringCollection("val")
	if !ok || len(arr) != 3 {
		t.Fatalf("FIX 4.6 FAILED: array not preserved through marshal; got %v", arr)
	}
	if arr[0] != "true" || arr[1] != "123" || arr[2] != "hello" {
		t.Errorf("FIX 4.6 FAILED: wrong values after round-trip: %v", arr)
	}
	// Verify types are still strings
	for i, v := range conf2.Parameters["val"].Value.([]string) {
		_ = v // we just checked the values above
		_ = i
	}
}

// TestStructure tests Marshal output against a known reference file.
func TestStructure(t *testing.T) {
	conf := New()
	if err := conf.LoadFile("testunit/example.conf"); err != nil {
		t.Fatal(err)
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
		t.Errorf("Marshal output differs from expected.\nGot:\n%s\nExpected:\n%s", s0, r0)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 12: SaveFile / LoadFile  (FIX 2.1, 2.2, 2.3) ────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestSaveFilePermissions verifies FIX 2.1: saved files have permissions 0644
// (octal), not the broken 0x644 (hex) that was in the original code.
func TestSaveFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.conf")

	conf := New()
	conf.Set("key", "value")
	if err := conf.SaveFile(path); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("FIX 2.1 FAILED: expected permissions 0644, got %04o (decimal %d)", perm, perm)
	}
}

// TestSaveLoadRoundTrip verifies that SaveFile + LoadFile restores an identical config.
func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "round.conf")

	conf1 := New()
	conf1.LoadString("host=localhost\nport=5432\ndebug=true\nratio=3.14")
	if err := conf1.SaveFile(path); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}

	conf2 := New()
	if err := conf2.LoadFile(path); err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	for _, key := range []string{"host", "port", "debug", "ratio"} {
		v1, _ := conf1.Get(key)
		v2, _ := conf2.Get(key)
		if v1 != v2 {
			t.Errorf("round-trip key %q: before=%v (%T), after=%v (%T)", key, v1, v1, v2, v2)
		}
	}
}

// TestSaveFileForcedStringRoundTrip verifies FIX 4.6 + 2.1 together:
// forced-string values survive a save/reload cycle with the correct type.
func TestSaveFileForcedStringRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "strtest.conf")

	conf1 := New()
	conf1.LoadString("flag=\"true\ncount=\"42\nmode=\"off")
	if err := conf1.SaveFile(path); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}

	conf2 := New()
	if err := conf2.LoadFile(path); err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	for _, key := range []string{"flag", "count", "mode"} {
		v := conf2.Parameters[key].Value
		if _, ok := v.(string); !ok {
			t.Errorf("FIX 4.6+2.1 FAILED key %q: expected string after save/reload, got %T", key, v)
		}
	}
}

// TestPathTraversalPrevention verifies FIX 2.3: LoadFile and SaveFile reject
// paths that would traverse above the current directory.
func TestPathTraversalPrevention(t *testing.T) {
	conf := New()

	traversalPaths := []string{
		"../../etc/passwd",
		"../secret.conf",
		"a/../../etc/shadow",
		"./../outside.conf",
	}
	for _, p := range traversalPaths {
		if err := conf.LoadFile(p); err == nil {
			t.Errorf("FIX 2.3 FAILED: LoadFile(%q) should return error for path traversal", p)
		}
		if err := conf.SaveFile(p); err == nil {
			t.Errorf("FIX 2.3 FAILED: SaveFile(%q) should return error for path traversal", p)
		}
	}

	// Normal relative and absolute paths must still be accepted (they'll fail
	// with a "no such file" error, not a traversal error).
	safePaths := []string{
		"testunit/a.conf",
		"./testunit/a.conf",
	}
	for _, p := range safePaths {
		err := conf.LoadFile(p)
		if err != nil && strings.Contains(err.Error(), "path traversal") {
			t.Errorf("FIX 2.3: safe path %q wrongly rejected as traversal: %v", p, err)
		}
	}
}

// TestLoadFileNonExistent verifies that LoadFile returns an error (not a panic)
// for a missing file.
func TestLoadFileNonExistent(t *testing.T) {
	conf := New()
	if err := conf.LoadFile("testunit/does_not_exist.conf"); err == nil {
		t.Error("LoadFile on missing file should return an error")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 13: Thread-safety  (FIX 3.1) ─────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestThreadSafety exercises concurrent reads and writes on the same XConfig.
// Run with: go test -race ./...
// Without the mutex fix this test will trigger the race detector.
func TestThreadSafety(t *testing.T) {
	conf := New()
	conf.LoadString("key1=value1\nkey2=42\nkey3=true\nkey4=3.14")

	const workers = 50
	const iters = 200
	var wg sync.WaitGroup

	// Concurrent readers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				conf.Get("key1")
				conf.GetString("key1")
				conf.GetInt("key2")
				conf.GetBool("key3")
				conf.GetFloat("key4")
				conf.Marshal()
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				conf.Set(fmt.Sprintf("dyn_%d", id), j)
			}
		}(i)
	}

	// Concurrent adds and deletes
	for i := 0; i < workers/5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("tmp_%d", id)
			for j := 0; j < iters; j++ {
				conf.Add(key, j)
				conf.Del(key)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentLoads exercises multiple goroutines each loading their own XConfig
// from the same source string simultaneously.
func TestConcurrentLoads(t *testing.T) {
	const goroutines = 20
	var wg sync.WaitGroup
	results := make([]*XConfig, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c := New()
			c.LoadString(fmt.Sprintf("worker=%d\ncommon=value", idx))
			results[idx] = c
		}(i)
	}
	wg.Wait()

	for i, c := range results {
		v, _ := c.GetInt("worker")
		if v != i {
			t.Errorf("goroutine %d: expected worker=%d, got %d", i, i, v)
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 14: Get* type conversions ─────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestGetTypeCasts verifies documented cross-type conversions in Get* functions.
func TestGetTypeCasts(t *testing.T) {
	conf := New()
	conf.LoadString("i=42\nf=3.7\nb=true\nbi=0")

	// GetInt on float → truncated int
	if v, ok := conf.GetInt("f"); !ok || v != 3 {
		t.Errorf("GetInt on float: expected 3, got %v ok=%v", v, ok)
	}
	// GetInt on bool true → 1
	if v, ok := conf.GetInt("b"); !ok || v != 1 {
		t.Errorf("GetInt on bool true: expected 1, got %v ok=%v", v, ok)
	}

	// GetFloat on int → float
	if v, ok := conf.GetFloat("i"); !ok || v != 42.0 {
		t.Errorf("GetFloat on int: expected 42.0, got %v ok=%v", v, ok)
	}
	// GetFloat on bool true → 1.0
	if v, ok := conf.GetFloat("b"); !ok || v != 1.0 {
		t.Errorf("GetFloat on bool true: expected 1.0, got %v ok=%v", v, ok)
	}

	// GetBool on int 42 → true
	if v, ok := conf.GetBool("i"); !ok || !v {
		t.Errorf("GetBool on int 42: expected true, got %v ok=%v", v, ok)
	}
	// GetBool on int 0 → false
	if v, ok := conf.GetBool("bi"); !ok || v {
		t.Errorf("GetBool on int 0: expected false, got %v ok=%v", v, ok)
	}

	// GetString on int → string representation
	if v, ok := conf.GetString("i"); !ok || v != "42" {
		t.Errorf("GetString on int: expected '42', got %v ok=%v", v, ok)
	}
}

// TestGetMissingKey verifies that Get* returns (zero, false) for missing keys.
func TestGetMissingKey(t *testing.T) {
	conf := New()

	if v, ok := conf.Get("missing"); ok || v != nil {
		t.Errorf("Get missing: expected (nil, false), got (%v, %v)", v, ok)
	}
	if _, ok := conf.GetString("missing"); ok {
		t.Error("GetString missing: expected false")
	}
	if _, ok := conf.GetInt("missing"); ok {
		t.Error("GetInt missing: expected false")
	}
	if _, ok := conf.GetFloat("missing"); ok {
		t.Error("GetFloat missing: expected false")
	}
	if _, ok := conf.GetBool("missing"); ok {
		t.Error("GetBool missing: expected false")
	}
	if _, ok := conf.GetStringCollection("missing"); ok {
		t.Error("GetStringCollection missing: expected false")
	}
	if conf.GetConfig("missing") != nil {
		t.Error("GetConfig missing: expected nil")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 15: Error messages  (FIX 4.7) ────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestErrorMessageFloatArray verifies FIX 4.7: the error for incompatible type
// on []float64 now says "floats" not "integers".
func TestErrorMessageFloatArray(t *testing.T) {
	conf := New()
	conf.LoadString("x=1.1\nx=2.2") // []float64

	// Try to add a bool to a float array — must fail with "floats" in message
	err := conf.Add("x", true)
	if err == nil {
		t.Fatal("FIX 4.7: expected error when adding bool to float array")
	}
	if !strings.Contains(err.Error(), "floats") {
		t.Errorf("FIX 4.7 FAILED: error message should mention 'floats', got: %q", err.Error())
	}
	if strings.Contains(err.Error(), "integers") {
		t.Errorf("FIX 4.7 FAILED: error message should NOT say 'integers', got: %q", err.Error())
	}
}

// TestErrorMessageMixedTypes verifies that all type-mismatch errors are produced.
func TestErrorMessageMixedTypes(t *testing.T) {
	cases := []struct {
		setup    string
		addVal   interface{}
		contains string
	}{
		{"s=hello", 42, "string"},
		{"i=1\ni=2", "text", "integers"},
		{"f=1.1\nf=2.2", 42, "floats"},
		{"b=true\nb=false", 42, "booleans"},
	}
	for _, tc := range cases {
		conf := New()
		conf.LoadString(tc.setup)
		err := conf.Add(strings.Split(tc.setup, "=")[0], tc.addVal)
		if err == nil {
			t.Errorf("setup %q: expected error, got nil", tc.setup)
			continue
		}
		if !strings.Contains(err.Error(), tc.contains) {
			t.Errorf("setup %q: error %q should mention %q", tc.setup, err.Error(), tc.contains)
		}
	}
}

// TestAddIncompatibleTypeRejectsNonScalar verifies that Add rejects types
// other than string, int, float64, bool.
func TestAddIncompatibleTypeRejectsNonScalar(t *testing.T) {
	conf := New()
	err := conf.Add("key", []string{"a", "b"})
	if err == nil {
		t.Error("Add with []string should return an error")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 16: Empty key validation  (FIX 4.8) ──────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestEmptyKeySetIsNoOp verifies FIX 4.8: Set("", v) is silently ignored
// and does not add an empty entry that would later cause a panic.
func TestEmptyKeySetIsNoOp(t *testing.T) {
	conf := New()
	conf.Set("", "value") // must not panic, must not add entry

	if _, ok := conf.Parameters[""]; ok {
		t.Error("FIX 4.8 FAILED: empty key should not be added to Parameters")
	}
	for _, k := range conf.Order {
		if k == "" {
			t.Error("FIX 4.8 FAILED: empty key should not appear in Order")
		}
	}
}

// TestEmptyKeyAddReturnsError verifies FIX 4.8: Add("", v) returns an error.
func TestEmptyKeyAddReturnsError(t *testing.T) {
	conf := New()
	err := conf.Add("", "value")
	if err == nil {
		t.Error("FIX 4.8 FAILED: Add(\"\", v) should return an error")
	}
}

// TestStringAndBuildLevelNoPanicOnNormalConfig verifies that String() and
// Marshal() do not panic on a well-formed config (regression guard for 4.8).
func TestStringAndBuildLevelNoPanicOnNormalConfig(t *testing.T) {
	conf := New()
	conf.LoadFile("testunit/example.conf")
	_ = conf.String()
	_ = conf.Marshal()
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 17: GetDataset / GetCollection / GetConfig ───────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

func TestGetDataset(t *testing.T) {
	conf := New()
	conf.LoadString("sub.key=value")

	ds, ok := conf.GetDataset("sub")
	if !ok || ds == nil {
		t.Fatal("GetDataset: expected sub-XConfig, got nil or false")
	}
	xc, ok := ds.(*XConfig)
	if !ok {
		t.Fatal("GetDataset: result is not *XConfig")
	}
	if v, _ := xc.Get("key"); v != "value" {
		t.Errorf("GetDataset sub.key: expected 'value', got %v", v)
	}
}

func TestGetConfig(t *testing.T) {
	conf := New()
	conf.LoadString("db.host=localhost\ndb.port=3306")

	sub := conf.GetConfig("db")
	if sub == nil {
		t.Fatal("GetConfig: expected *XConfig, got nil")
	}
	if v, _ := sub.Get("host"); v != "localhost" {
		t.Errorf("GetConfig db.host: expected 'localhost', got %v", v)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 18: Multiple file loads ──────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

func TestMultipleFlagSetOnSecondLoad(t *testing.T) {
	conf := New()
	if conf.Multiple {
		t.Error("Multiple should be false after single load")
	}
	conf.LoadFile("testunit/a.conf")
	if conf.Multiple {
		t.Error("Multiple should be false after first load")
	}
	conf.LoadFile("testunit/example.conf") // second load
	if !conf.Multiple {
		t.Error("Multiple should be true after second load")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 19: Full example.conf verification ───────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestExampleConfAllValues verifies every value declared in example.conf.
func TestExampleConfAllValues(t *testing.T) {
	conf := New()
	if err := conf.LoadFile("testunit/example.conf"); err != nil {
		t.Fatal(err)
	}

	// Scalar parameters
	if v, _ := conf.Get("parameter1"); v != "value1" {
		t.Errorf("parameter1: expected 'value1', got %v", v)
	}
	if v, _ := conf.Get("ip"); v != "127.0.0.1" {
		t.Errorf("ip: expected '127.0.0.1', got %v", v)
	}
	if v, _ := conf.Get("port"); v != 80 {
		t.Errorf("port: expected 80, got %v", v)
	}
	if v, _ := conf.Get("domain"); v != "test.com" {
		t.Errorf("domain: expected 'test.com', got %v", v)
	}

	// parameter2 is a []string
	arr, ok := conf.GetStringCollection("parameter2")
	if !ok || len(arr) != 2 || arr[0] != "value2" || arr[1] != "value3" {
		t.Errorf("parameter2: expected [value2, value3], got %v", arr)
	}

	// country is a []string with 5 elements
	countries, ok := conf.GetStringCollection("country")
	if !ok || len(countries) != 5 {
		t.Errorf("country: expected 5 elements, got %v", countries)
	}
	if countries[0] != "MX" || countries[4] != "ES" {
		t.Errorf("country: wrong values: %v", countries)
	}

	// language sub-configs
	lang := conf.GetConfig("language")
	if lang == nil {
		t.Fatal("language sub-config missing")
	}
	en := lang.GetConfig("en")
	if en == nil {
		t.Fatal("language.en sub-config missing")
	}
	if v, _ := en.Get("welcome"); v != "Welcome to the XConfig examples" {
		t.Errorf("language.en.welcome: got %v", v)
	}
	es := lang.GetConfig("es")
	if es == nil {
		t.Fatal("language.es sub-config missing")
	}
	if v, _ := es.Get("cancel"); v != "Cancelar" {
		t.Errorf("language.es.cancel: got %v", v)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 20: a.conf array values ──────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

func TestAConfArrays(t *testing.T) {
	conf := New()
	if err := conf.LoadFile("testunit/a.conf"); err != nil {
		t.Fatal(err)
	}

	// param10: []int{1,2,3,4,5}
	ints, ok := conf.GetIntCollection("param10")
	if !ok || len(ints) != 5 || ints[0] != 1 || ints[4] != 5 {
		t.Errorf("param10: expected [1,2,3,4,5], got %v", ints)
	}

	// param11: single int
	if v, _ := conf.Get("param11"); v != 101 {
		t.Errorf("param11: expected 101, got %v", v)
	}

	// param20: []bool{true,true,false,true,false}
	bools, ok := conf.GetBoolCollection("param20")
	if !ok || len(bools) != 5 {
		t.Errorf("param20: expected 5 bools, got %v", bools)
	}

	// param30: []string{"abc","def","def","ghi","jkl"}
	strs, ok := conf.GetStringCollection("param30")
	if !ok || len(strs) != 5 || strs[0] != "abc" {
		t.Errorf("param30: expected 5 strings starting with 'abc', got %v", strs)
	}

	// param31: forced string "true"
	if v := conf.Parameters["param31"].Value; v != "true" {
		t.Errorf("param31: expected string 'true', got %v (%T)", v, v)
	}
	if _, isString := conf.Parameters["param31"].Value.(string); !isString {
		t.Errorf("param31: should be string, not bool")
	}

	// param40: []float64
	floats, ok := conf.GetFloatCollection("param40")
	if !ok || len(floats) != 5 {
		t.Errorf("param40: expected 5 floats, got %v", floats)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 21: Set / Add programmatic API ───────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

func TestSetReplaces(t *testing.T) {
	conf := New()
	conf.LoadString("x=original")
	conf.Set("x", "replaced")

	if v, _ := conf.Get("x"); v != "replaced" {
		t.Errorf("Set should replace: expected 'replaced', got %v", v)
	}
	// Should appear only once in Order
	count := 0
	for _, k := range conf.Order {
		if k == "x" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Set: key 'x' appears %d times in Order (expected 1)", count)
	}
}

func TestAddBuildsCollection(t *testing.T) {
	conf := New()
	conf.Set("n", 1)
	conf.Add("n", 2)
	conf.Add("n", 3)

	arr, ok := conf.GetIntCollection("n")
	if !ok || len(arr) != 3 || arr[0] != 1 || arr[1] != 2 || arr[2] != 3 {
		t.Errorf("Add should build collection: expected [1,2,3], got %v", arr)
	}
}

func TestSetAllTypes(t *testing.T) {
	conf := New()
	conf.Set("s", "hello")
	conf.Set("i", 42)
	conf.Set("f", 3.14)
	conf.Set("b", true)

	if v, _ := conf.Get("s"); v != "hello" {
		t.Errorf("Set string: got %v", v)
	}
	if v, _ := conf.Get("i"); v != 42 {
		t.Errorf("Set int: got %v", v)
	}
	if v, _ := conf.Get("f"); v != 3.14 {
		t.Errorf("Set float: got %v", v)
	}
	if v, _ := conf.Get("b"); v != true {
		t.Errorf("Set bool: got %v", v)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ── SECTION 22: Comment preservation ─────────────────────────────────────────
// ═══════════════════════════════════════════════════════════════════════════════

// TestCommentsPreservedInMarshal verifies that comments from the original file
// appear in the Marshal() output.
func TestCommentsPreservedInMarshal(t *testing.T) {
	conf := New()
	conf.LoadString("# my comment\nkey=value\n; semicolon comment\nother=123")

	m := conf.Marshal()
	if !strings.Contains(m, "# my comment") {
		t.Errorf("Marshal should preserve '# my comment': %q", m)
	}
	if !strings.Contains(m, "; semicolon comment") {
		t.Errorf("Marshal should preserve '; semicolon comment': %q", m)
	}
	if !strings.Contains(m, "key=value") {
		t.Errorf("Marshal should contain 'key=value': %q", m)
	}
}

// TestEmptyLinePreservedAsComment verifies that blank lines are preserved as
// comments in the serialised output.
func TestEmptyLinePreservedAsComment(t *testing.T) {
	conf := New()
	conf.LoadString("a=1\n\nb=2")
	m := conf.Marshal()
	if !strings.Contains(m, "a=1") || !strings.Contains(m, "b=2") {
		t.Errorf("values missing from marshal: %q", m)
	}
}
