package cmd

import (
	"testing"

	"github.com/dlclark/regexp2"
)

func Test_checkIfBeginScope_Valid(t *testing.T) {

	var rxMatch = "^\\s*BEGIN\\s*$"
	var rx = regexp2.MustCompile(rxMatch, regexp2.Singleline)

	var lines []string = []string{
		"BEGIN",
		"BEGIN  ",
		"  BEGIN",
		"  BEGIN  ",
	}

	for _, v := range lines {
		if ok := checkIfBeginScope(v, rx, false); ok == false {
			t.Errorf("Begin scope [%v] should be found in [%v]", rxMatch, v)
		}
		if ok := checkIfBeginScope(v, rx, true); ok == true {
			t.Errorf("Begin scope [%v] should be found in [%v]", rxMatch, v)
		}
	}
}

func Test_checkIfEndScope_Valid(t *testing.T) {

	var rxMatch = "^\\s*END\\s*$"
	var rx = regexp2.MustCompile(rxMatch, regexp2.Singleline)

	var lines []string = []string{
		"END",
		"END  ",
		"  END",
		"  END  ",
	}

	for _, v := range lines {
		if ok := checkIfEndScope(v, rx, true); ok == false {
			t.Errorf("End scope [%v] should be found in [%v]", rxMatch, v)
		}
		if ok := checkIfEndScope(v, rx, false); ok == true {
			t.Errorf("End scope [%v] should be found in [%v]", rxMatch, v)
		}
	}
}

func Test_fileNameHash_Valid(t *testing.T) {

	txt := "golang"
	if v := fileNameHash(txt); v != "21cc2" {
		t.Errorf("fileNameHash [%v] is not valid for [%v]", v, txt)
	}

}
