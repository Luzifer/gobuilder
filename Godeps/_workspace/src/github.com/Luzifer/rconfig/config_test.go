package rconfig

import (
	"os"
	"testing"
)

func TestGeneralMechanics(t *testing.T) {
	cfg := struct {
		Test        string `default:"foo" env:"shell" flag:"shell" description:"Test"`
		Test2       string `default:"blub" env:"testvar" flag:"testvar,t" description:"Test"`
		DefaultFlag string `default:"goo"`
		SadFlag     string
	}{}

	parse(&cfg, []string{
		"--shell=test23",
		"-t", "bla",
	})

	if cfg.Test != "test23" {
		t.Errorf("Test should be 'test23', is '%s'", cfg.Test)
	}

	if cfg.Test2 != "bla" {
		t.Errorf("Test2 should be 'bla', is '%s'", cfg.Test2)
	}

	if cfg.SadFlag != "" {
		t.Errorf("SadFlag should be '', is '%s'", cfg.SadFlag)
	}

	if cfg.DefaultFlag != "goo" {
		t.Errorf("DefaultFlag should be 'goo', is '%s'", cfg.DefaultFlag)
	}

	parse(&cfg, []string{})

	if cfg.Test != "foo" {
		t.Errorf("Test should be 'foo', is '%s'", cfg.Test)
	}

	os.Setenv("shell", "test546")
	parse(&cfg, []string{})

	if cfg.Test != "test546" {
		t.Errorf("Test should be 'test546', is '%s'", cfg.Test)
	}
}

func TestBool(t *testing.T) {
	cfg := struct {
		Test1 bool `default:"true"`
		Test2 bool `default:"false" flag:"test2"`
		Test3 bool `default:"true" flag:"test3,t"`
		Test4 bool `flag:"test4"`
	}{}

	parse(&cfg, []string{
		"--test2",
		"-t",
	})

	if !cfg.Test1 {
		t.Errorf("Test1 should be 'true', is '%+v'", cfg.Test1)
	}
	if !cfg.Test2 {
		t.Errorf("Test1 should be 'true', is '%+v'", cfg.Test2)
	}
	if !cfg.Test3 {
		t.Errorf("Test1 should be 'true', is '%+v'", cfg.Test3)
	}
	if cfg.Test4 {
		t.Errorf("Test1 should be 'false', is '%+v'", cfg.Test3)
	}
}

func TestInt(t *testing.T) {
	cfg := struct {
		Test    int   `flag:"int"`
		TestP   int   `flag:"intp,i"`
		Test8   int8  `flag:"int8"`
		Test8P  int8  `flag:"int8p,8"`
		Test32  int32 `flag:"int32"`
		Test32P int32 `flag:"int32p,3"`
		Test64  int64 `flag:"int64"`
		Test64P int64 `flag:"int64p,6"`
		TestDef int8  `default:"66"`
	}{}

	parse(&cfg, []string{
		"--int=1", "-i", "2",
		"--int8=3", "-8", "4",
		"--int32=5", "-3", "6",
		"--int64=7", "-6", "8",
	})

	if cfg.Test != 1 || cfg.TestP != 2 || cfg.Test8 != 3 || cfg.Test8P != 4 || cfg.Test32 != 5 || cfg.Test32P != 6 || cfg.Test64 != 7 || cfg.Test64P != 8 {
		t.Errorf("One of the int tests failed.")
	}

	if cfg.TestDef != 66 {
		t.Errorf("TestDef should be '66', is '%d'", cfg.TestDef)
	}
}

func TestUint(t *testing.T) {
	cfg := struct {
		Test    uint   `flag:"int"`
		TestP   uint   `flag:"intp,i"`
		Test8   uint8  `flag:"int8"`
		Test8P  uint8  `flag:"int8p,8"`
		Test16  uint16 `flag:"int16"`
		Test16P uint16 `flag:"int16p,1"`
		Test32  uint32 `flag:"int32"`
		Test32P uint32 `flag:"int32p,3"`
		Test64  uint64 `flag:"int64"`
		Test64P uint64 `flag:"int64p,6"`
		TestDef uint8  `default:"66"`
	}{}

	parse(&cfg, []string{
		"--int=1", "-i", "2",
		"--int8=3", "-8", "4",
		"--int32=5", "-3", "6",
		"--int64=7", "-6", "8",
		"--int16=9", "-1", "10",
	})

	if cfg.Test != 1 || cfg.TestP != 2 || cfg.Test8 != 3 || cfg.Test8P != 4 || cfg.Test32 != 5 || cfg.Test32P != 6 || cfg.Test64 != 7 || cfg.Test64P != 8 || cfg.Test16 != 9 || cfg.Test16P != 10 {
		t.Errorf("One of the uint tests failed.")
	}

	if cfg.TestDef != 66 {
		t.Errorf("TestDef should be '66', is '%d'", cfg.TestDef)
	}
}

func TestFloat(t *testing.T) {
	cfg := struct {
		Test32  float32 `flag:"float32"`
		Test32P float32 `flag:"float32p,3"`
		Test64  float64 `flag:"float64"`
		Test64P float64 `flag:"float64p,6"`
		TestDef float32 `default:"66.256"`
	}{}

	parse(&cfg, []string{
		"--float32=5.5", "-3", "6.6",
		"--float64=7.7", "-6", "8.8",
	})

	if cfg.Test32 != 5.5 || cfg.Test32P != 6.6 || cfg.Test64 != 7.7 || cfg.Test64P != 8.8 {
		t.Errorf("One of the int tests failed.")
	}

	if cfg.TestDef != 66.256 {
		t.Errorf("TestDef should be '66.256', is '%.3f'", cfg.TestDef)
	}
}

func TestSubStruct(t *testing.T) {
	cfg := struct {
		Test string `default:"blubb"`
		Sub  struct {
			Test string `default:"Hallo"`
		}
	}{}

	if err := parse(&cfg, []string{}); err != nil {
		t.Errorf("Test errored: %s", err)
	}

	if cfg.Test != "blubb" {
		t.Errorf("Test should be 'blubb', is '%s'", cfg.Test)
	}

	if cfg.Sub.Test != "Hallo" {
		t.Errorf("Sub.Test should be 'Hallo', is '%s'", cfg.Sub.Test)
	}
}

func TestSlice(t *testing.T) {
	cfg := struct {
		Int     []int    `default:"1,2,3" flag:"int"`
		String  []string `default:"a,b,c" flag:"string"`
		IntP    []int    `default:"1,2,3" flag:"intp,i"`
		StringP []string `default:"a,b,c" flag:"stringp,s"`
	}{}

	if err := parse(&cfg, []string{
		"--int=4,5", "-s", "hallo,welt",
	}); err != nil {
		t.Errorf("Test errored: %s", err)
	}

	if len(cfg.Int) != 2 || cfg.Int[0] != 4 || cfg.Int[1] != 5 {
		t.Errorf("Int should be '4,5', is '%+v'", cfg.Int)
	}

	if len(cfg.String) != 3 || cfg.String[0] != "a" || cfg.String[1] != "b" {
		t.Errorf("String should be 'a,b,c', is '%+v'", cfg.String)
	}

	if len(cfg.StringP) != 2 || cfg.StringP[0] != "hallo" || cfg.StringP[1] != "welt" {
		t.Errorf("StringP should be 'hallo,welt', is '%+v'", cfg.StringP)
	}
}

func TestErrors(t *testing.T) {
	if err := parse(&struct {
		A int `default:"a"`
	}{}, []string{}); err == nil {
		t.Errorf("Test should have errored")
	}

	if err := parse(&struct {
		A float32 `default:"a"`
	}{}, []string{}); err == nil {
		t.Errorf("Test should have errored")
	}

	if err := parse(&struct {
		A uint `default:"a"`
	}{}, []string{}); err == nil {
		t.Errorf("Test should have errored")
	}

	if err := parse(&struct {
		B struct {
			A uint `default:"a"`
		}
	}{}, []string{}); err == nil {
		t.Errorf("Test should have errored")
	}

	if err := parse(&struct {
		A []int `default:"a,bn"`
	}{}, []string{}); err == nil {
		t.Errorf("Test should have errored")
	}
}

func TestOSArgs(t *testing.T) {
	os.Args = []string{"--a=bar"}

	cfg := struct {
		A string `default:"a" flag:"a"`
	}{}

	Parse(&cfg)

	if cfg.A != "bar" {
		t.Errorf("A should be 'bar', is '%s'", cfg.A)
	}
}

func TestNonPointer(t *testing.T) {
	cfg := struct {
		A string `default:"a"`
	}{}

	if err := parse(cfg, []string{}); err == nil {
		t.Errorf("Test should have errored")
	}
}

func TestOtherType(t *testing.T) {
	cfg := "test"

	if err := parse(&cfg, []string{}); err == nil {
		t.Errorf("Test should have errored")
	}
}
