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
