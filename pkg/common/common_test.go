package common

import (
	"bytes"
	"testing"
	"time"
)

func Test_MatchConfig_ShouldBeValid(t *testing.T) {

	matches := []MatchConfig{
		{
			Name:       "match-1",
			Match:      ".*",
			Occurrence: "+",
		},
		{
			Name:       "match-2",
			Match:      ".*",
			Occurrence: "+",
		},
		{
			Name:       "match-3",
			Match:      "^\\d+",
			Occurrence: "!",
		},
		{
			Name:       "match-4",
			Match:      "^\\d+",
			Occurrence: "2",
		},
	}

	for _, m := range matches {
		if e := m.IsValid(); e != nil {
			t.Errorf("Match %v is not valid : %v", m.Name, e.Error())
		}
	}
}

func Test_MatchConfig_ShouldBeNotValid(t *testing.T) {

	matches := []MatchConfig{
		{
			Name:       "match-1",
			Match:      ".*",
			Occurrence: "a",
		},
		{
			Name:       "match-2",
			Match:      "^\\d+",
			Occurrence: "2-7",
		},
		{
			Name:       "match-3",
			Match:      "^\\d+",
			Occurrence: "!1",
		},
	}

	for _, m := range matches {
		if e := m.IsValid(); e == nil {
			t.Errorf("Match %v should be not valid", m.Name)
		}
	}
}

func Test_MatchConfig_IsMatchOccurrence(t *testing.T) {

	var lines []string = []string{"line1", "line2", "line3"}

	var matchConfigAll = MatchConfig{
		Name:           "match-all",
		Match:          ".*",
		IgnoreInResult: false,
		Occurrence:     "*",
	}

	var matchConfigAny = MatchConfig{
		Name:           "match-any",
		Match:          ".*",
		IgnoreInResult: false,
		Occurrence:     "+",
	}

	var matchConfigNone = MatchConfig{
		Name:           "match-none",
		Match:          ".*",
		IgnoreInResult: false,
		Occurrence:     "!",
	}

	var matchConfig3 = MatchConfig{
		Name:           "match-3",
		Match:          ".*",
		IgnoreInResult: false,
		Occurrence:     "3",
	}

	if v := matchConfigAll.IsMatchOccurrence(lines, 3); v == false {
		t.Errorf("Lines not match to [%v]", matchConfigAll.Name)
	}

	if v := matchConfigAll.IsMatchOccurrence(lines, 4); v == true {
		t.Errorf("Lines should not match to [%v]", matchConfigAll.Name)
	}

	if v := matchConfigAny.IsMatchOccurrence(lines, 100); v == false {
		t.Errorf("Lines not match to [%v]", matchConfigAny.Name)
	}

	if v := matchConfigAny.IsMatchOccurrence([]string{}, 100); v == true {
		t.Errorf("Lines should not match to [%v]", matchConfigAny.Name)
	}

	if v := matchConfigNone.IsMatchOccurrence([]string{}, 100); v == false {
		t.Errorf("Lines not match to [%v]", matchConfigNone.Name)
	}

	if v := matchConfigNone.IsMatchOccurrence(lines, 100); v == true {
		t.Errorf("Lines should not match to [%v]", matchConfigNone.Name)
	}

	if v := matchConfig3.IsMatchOccurrence(lines, 100); v == false {
		t.Errorf("Lines not match to [%v]", matchConfig3.Name)
	}

	if v := matchConfig3.IsMatchOccurrence([]string{"line1"}, 100); v == true {
		t.Errorf("Lines should not match to [%v]", matchConfig3.Name)
	}
}

func Test_MatchConfig_IsMatch(t *testing.T) {

	var lines []MatchLine = []MatchLine{
		{
			Index:      0,
			Line:       "line ABC",
			MatchNames: []string{},
		},
		{
			Index:      1,
			Line:       "line 500",
			MatchNames: []string{},
		},
		{
			Index:      2,
			Line:       "line 123",
			MatchNames: []string{},
		},
		{
			Index:      3,
			Line:       "/* ***",
			MatchNames: []string{},
		},
		{
			Index:      4,
			Line:       "/* 000",
			MatchNames: []string{},
		},
		{
			Index:      5,
			Line:       "line 999",
			MatchNames: []string{},
		},
		{
			Index:      6,
			Line:       "line 1000",
			MatchNames: []string{},
		},
	}

	var matchConfigAll = MatchConfig{
		Name:           "match-all",
		Match:          "^.*$",
		IgnoreInResult: false,
		Occurrence:     "*",
	}

	var matchConfigAny = MatchConfig{
		Name:           "match-any",
		Match:          "^\\w+\\W+\\d+$",
		IgnoreInResult: false,
		Occurrence:     "+",
	}

	var matchConfigNone = MatchConfig{
		Name:           "match-none",
		Match:          "^.*\\s+$",
		IgnoreInResult: false,
		Occurrence:     "!",
	}

	var matchConfig3 = MatchConfig{
		Name:           "match-3",
		Match:          "^\\w+\\W+\\d{3}$",
		IgnoreInResult: false,
		Occurrence:     "3",
	}

	if b := matchConfigAll.IsMatch(lines); len(b) == 0 {
		t.Errorf("Lines not match to [%v]", matchConfigAll.Name)
	}

	if b := matchConfigAny.IsMatch(lines); len(b) == 0 {
		t.Errorf("Lines not match to [%v]", matchConfigAny.Name)
	}

	if b := matchConfigNone.IsMatch(lines); len(b) != 0 {
		t.Errorf("Lines should not match to [%v]", matchConfigNone.Name)
	}
	if b := matchConfig3.IsMatch(lines); len(b) != 3 {
		t.Errorf("Lines not match to [%v]", matchConfig3.Name)
	}
}

func Test_ScanConfig_ShouldBeValid(t *testing.T) {

	s := ScanConfig{
		Folder: ".\\example",
		Filter: "*.txt",
		Scopes: []ScopeConfig{
			{
				Name:      "example-1",
				Begin:     "^\\W*BEGIN$",
				End:       "^\\W*END$",
				AutoClose: true,
				Matches: []MatchConfig{
					{
						Name:       "match-1",
						Match:      ".*",
						Occurrence: "1",
					},
					{
						Name:       "match-2",
						Match:      ".*",
						Occurrence: "+",
					},
				},
			},
		},
	}

	if e := s.IsValid(); e != nil {
		t.Errorf("Scope %v should be valid. %v", s, e)
	}
}

func Test_ScanConfig_ShouldBeNotValid(t *testing.T) {

	s := []ScanConfig{
		{
			Folder: "",
			Filter: "*.txt",
			Scopes: []ScopeConfig{
				{
					Name:      "example-1",
					Begin:     "^\\W*BEGIN$",
					End:       "^\\W*END$",
					AutoClose: true,
					Matches: []MatchConfig{
						{
							Name:       "match-1",
							Match:      ".*",
							Occurrence: "1",
						},
					},
				},
			},
		},
		{
			Folder: "name",
			Filter: "",
			Scopes: []ScopeConfig{
				{
					Name:      "example-1",
					Begin:     "^\\W*BEGIN$",
					End:       "^\\W*END$",
					AutoClose: true,
					Matches: []MatchConfig{
						{
							Name:       "match-1",
							Match:      ".*",
							Occurrence: "1",
						},
					},
				},
			},
		},
		{
			Folder: ".\\example",
			Filter: "*.txt",
			Scopes: []ScopeConfig{
				{
					Name:      "",
					Begin:     "^\\W*BEGIN$",
					End:       "^\\W*END$",
					AutoClose: true,
					Matches: []MatchConfig{
						{
							Name:       "match-1",
							Match:      ".*",
							Occurrence: "1",
						},
					},
				},
			},
		},
		{
			Folder: ".\\example",
			Filter: "*.txt",
			Scopes: []ScopeConfig{
				{
					Name:      "example-1",
					Begin:     "",
					End:       "^\\W*END$",
					AutoClose: true,
					Matches: []MatchConfig{
						{
							Name:       "match-1",
							Match:      ".*",
							Occurrence: "1",
						},
					},
				},
			},
		},
		{
			Folder: ".\\example",
			Filter: "*.txt",
			Scopes: []ScopeConfig{
				{
					Name:      "example-1",
					Begin:     "^\\W*BEGIN$",
					End:       "",
					AutoClose: true,
					Matches: []MatchConfig{
						{
							Name:       "match-1",
							Match:      ".*",
							Occurrence: "1",
						},
					},
				},
			},
		},
		{
			Folder: ".\\example",
			Filter: "",
			Scopes: []ScopeConfig{
				{
					Name:      "example-1",
					Begin:     "^\\W*BEGIN$",
					End:       "",
					AutoClose: true,
					Matches: []MatchConfig{
						{
							Name:       "match-1",
							Match:      ".*",
							Occurrence: "1",
						},
					},
				},
			},
		},
		{
			Folder: ".\\example",
			Filter: "*.txt",
			Scopes: []ScopeConfig{
				{
					Name:      "example-2",
					Begin:     "^\\W*BEGIN$",
					End:       "^\\W*END$",
					AutoClose: true,
					Matches: []MatchConfig{
						{
							Name:       "match-1",
							Match:      ".*",
							Occurrence: "abc",
						},
					},
				},
			},
		},
		{
			Folder: ".\\example",
			Filter: "*.txt",
		},
		{
			Folder: ".\\example",
			Filter: "*.txt",
			Scopes: []ScopeConfig{
				{
					Name:      "example-3",
					Begin:     "^\\W*BEGIN$",
					End:       "^\\W*END$",
					AutoClose: true,
				},
			},
		},
	}

	for _, v := range s {
		if e := v.IsValid(); e == nil {
			t.Errorf("Scope %v should be not valid", v)
		}
	}
}

func Test_WriteScopeConfiguration_Valid(t *testing.T) {
	s := ScanConfig{
		Folder: ".\\example",
		Filter: "*.txt",
		Scopes: []ScopeConfig{
			{
				Name:      "example-1",
				Begin:     "^\\W*BEGIN$",
				End:       "^\\W*END$",
				AutoClose: true,
				Matches: []MatchConfig{
					{
						Name:       "match-1",
						Match:      ".*",
						Occurrence: "1",
					},
				},
			},
		},
	}

	if b, e := s.WriteScanConfiguration(); b == nil || e != nil {
		t.Errorf("Scan write error: %v", e)
	}
}

func Test_WriteAndReadScopeConfiguration_Valid(t *testing.T) {
	s := ScanConfig{
		Folder: ".\\example",
		Filter: "*.txt",
		Scopes: []ScopeConfig{
			{
				Name:      "example-1",
				Begin:     "^\\W*BEGIN$",
				End:       "^\\W*END$",
				AutoClose: true,
				Matches: []MatchConfig{
					{
						Name:       "match-1",
						Match:      ".*",
						Occurrence: "1",
					},
				},
			},
		},
	}

	var b1 []byte
	var s2 ScanConfig
	var e error

	if b1, e = s.WriteScanConfiguration(); b1 == nil || e != nil {
		t.Errorf("Scan write error: %v", e)
	}

	if s2, e = ReadScanConfiguration(b1); s2.IsValid() != nil || e != nil {
		t.Errorf("Scan read error: %v", e)
	}

	if b2, e := s2.WriteScanConfiguration(); !bytes.Equal(b1, b2) || e != nil {
		t.Errorf("Scan in not the same after write & read")
	}
}

func Test_ReadScopeConfiguration_NotValid_Content(t *testing.T) {

	b := []byte{'A', 'B', 'C', 'D', 'E'}

	if s, e := ReadScanConfiguration(b); s.IsValid() == nil || e == nil {
		t.Errorf("Scan read should be ends with error")
	}
}

func Test_ReadScopeConfiguration_ContentNotValid(t *testing.T) {
	s := ScanConfig{
		Folder: "",
		Filter: "*.txt",
		Scopes: []ScopeConfig{
			{
				Name:      "example-1",
				Begin:     "^\\W*BEGIN$",
				End:       "^\\W*END$",
				AutoClose: true,
				Matches: []MatchConfig{
					{
						Name:       "match-1",
						Match:      ".*",
						Occurrence: "1",
					},
				},
			},
		},
	}

	var b []byte
	var e error

	if b, e = s.WriteScanConfiguration(); b != nil || e == nil {
		t.Errorf("Scan should be invalid: %v", s)
	}

	if len(b) > 0 {
		b[0] = 0
	}

	if s, e := ReadScanConfiguration(b); s.IsValid() == nil || e == nil {
		t.Errorf("Scan read should be ends with error: %v", e)
	}
}

func Test_WriteAsHTML_Valid(t *testing.T) {

	s := ScanSummary{
		Folder:       "folder",
		Filter:       "*",
		CreationTime: time.Date(2020, 01, 01, 01, 01, 01, 01, time.UTC),
		DurationTime: time.Duration(1),
		Summary: []FileScopeSummary{
			{
				FileName:   "file1.txt",
				AllMatches: 1,
				Scopes: []ScopeSummary{
					{
						Name:          "name",
						FileName:      "file1.txt",
						Started:       1,
						Finished:      2,
						Content:       []string{"content"},
						ContentAsHTML: []string{"content-html"},
					},
				},
			},
		},
	}

	var b bytes.Buffer
	if e := s.WriteAsHTML(&b); e != nil {
		t.Errorf("ScanSummary cannot write as html.%v", e)
	}
	if len(b.Bytes()) == 0 {
		t.Errorf("ScanSummary write to html produced empty content")
	}
}
