package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
	"time"
)

const (
	pattern = `<!DOCTYPE html>
<html lang="en">
<meta charset="UTF-8">
<title>Scan summary</title>
<style>

    .poweredby {
		font-family: Verdana, Geneva, sans-serif;
		font-size: 10px;
		color: white;	
	}

    .title {
		border: 2px solid #1C6EA4;
		background: #80DCF5;
		font-family: Verdana, Geneva, sans-serif;
		font-size: 14px;
		color: #000000;	
		padding:5px;
		margin-bottom:5px;
	}

	.title-tbl {
		border:2px;
		padding:5px;
	}
	.title-tbl caption {
		text-align:left;
	}

	.title-tbl td {
		border:1px;
		padding:5px;
		background:#FFFFFF;
	}

    .result {
		border: 2px solid #1C6EA4;
		//background: #D0E4F5;
		font-family: Verdana, Geneva, sans-serif;
		font-size: 14px;
		color: #000000;
		padding:2px;
	}

    .summary {
		border: 1px solid #1C6EA4;
		background: #F5F5F5;
		font-family: Verdana, Geneva, sans-serif;
		font-size: 12px;
		color: #000000;
		padding:5px;
		margin:15px;
	}

	.scope {
		border-top: 2px dotted #AAAAAA;
		margin-bottom:15px;
		margin-top:15px;
		margin-left:15px;
	}

	.summary-title {
		background: #1C6EA4;
		font-family: Verdana, Geneva, sans-serif;
		font-size: 12px;
		color: #FFFFFF;
		padding:5px;
		margin:5px;
	}

	.tbl {
		width:100%;
		border:2px #A3A3A3;
		padding:5px;
		table-layout: fixed;
		width: 100%;  
	}
	.tbl caption {
		text-align:left;
	}
	.tbl th {
		text-align:left;
		border:1px;
		padding:5px;
		background:#E0E0E0;
	}
	.tbl td {
		text-align:left;
		border:1px;
		padding:5px;
		background:#FFFFFF;
	}

	.collapsible1 {
		background-color: #777;
		color: white;
		cursor: pointer;
		padding: 18px;
		width: 100%;
		border: none;
		text-align: left;
		outline: none;
		font-size: 15px;
	  }

	  .collapsible1 {
		background-color: #FFF;
		color: #000;
		cursor: pointer;
		padding: 0px;
		width: 100%;
		border: none;
		text-align: left;
		outline: none;
		font-size: 15px;
		background-image
	  }
	  
	  .active, .collapsible:hover {
		color: navy;
		font-weight: bold;
	  }
	  
	  .content {
		display: none;
		overflow: hidden;
		background-color: #f0f0f0;
		border: 1px dotted #AAAAAA;
		margin:10px;
		font-family: Courier New;
	  }

	}

</style>
<body>
	<div class="title" id="title">
	    <div class="poweredby">powered by gorex (https://github.com/tomdef/gorex)</div>
		<h2>Scan summary:</h2>		
		<table class="title-tbl">
		<caption>Parameters:</caption>
			<tbody>
			<tr>
				<td>Folder</td>
				<td>{{.Folder}}</td>
			</tr>
			<tr>
				<td>Filter</td>
				<td>{{.Filter}}</td>
			</tr>
			<tr>
			<td>Creation time</td>
			<td>{{.CreationTime}}</td>
		</tr>
			</tbody>
		</table>	

		<table class="title-tbl">		
		<caption>Summary:</caption>
			<tbody>
			<tr>
				<td>Scan file(s)</td>
				<td><b>{{.ScanFiles}}</b> </td>
			</tr>
			<tr>
				<td>Found in file(s)</td>
				<td><b>{{len .Summary}}</b> </td>
			</tr>
			{{range .Summary}}
			<tr>
				<td>File name</td>
				<td><b>{{.FileName}}</b></td>
				<td>Scope(s)</td>
				<td><b>{{len .Scopes}}</b></td>
				<td>All matches in file</td>
				<td><b>{{.AllMatches}}</b></td>
				<td><b><a href="#{{.FileName}}">Go to file details</a></b></td>
			</tr>
			{{end}}
			</tbody>
		</table>	

	</div>	
	<div class="result">
	{{range .Summary}}
		<div class="summary">
		<p class="summary-title" id="{{.FileName}}">File name [<b><a class="summary-title" href="file:///{{.FileName}}">{{.FileName}}</a></b>][<a href="#title" class="summary-title">Go to top</a>]</p>
		{{range .Scopes}}
		<div class="scope">
			<p>Scope name <b>{{.Name}}</b></p>
			{{if .Started}}
			<p>Scope line range: [<b>{{.Started}}</b>..<b>{{.Finished}}</b>]</p>
			<p type="button" class="collapsible">Scope content [show/hide]:</button>
			<div class="content">
				{{range $element := .ContentAsHTML}} 
{{$element}}<br/>
				{{end}}				
			</div>	
			{{end}}
			<table class="tbl">
				<caption>Matches:</caption>
				<thead>
					<tr>
						<th style="width:100px;">Line index</th>
						<th>Text</th>
					</tr>
				</thead>
				<tbody>
					{{range .Matches}}
					<tr>
						<td style="width:100px;">{{.Index}}</td>
						<td>{{.Line}}</td>
					</tr>
					{{end}}
				</tbody>
			</table>			
		</div>
		{{end}}	
		</div>
	{{end}}	
	</div>

	<script>
	var coll = document.getElementsByClassName("collapsible");
	var i;
	
	for (i = 0; i < coll.length; i++) {
	  coll[i].addEventListener("click", function() {
		this.classList.toggle("active");
		var content = this.nextElementSibling;
		if (content.style.display === "block") {
		  content.style.display = "none";
		} else {
		  content.style.display = "block";
		}
	  });
	}
	</script>
</body>
</html>	`
)

// -----------------------------------------------------------------------------
// types
// -----------------------------------------------------------------------------

// SearchQueryOperator describe types of operator ...
type SearchQueryOperator int

const (
	//SearchQueryOperatorAll is AND operator for queries
	SearchQueryOperatorAll SearchQueryOperator = iota
	//SearchQueryOperatorAny is OR operator for queries
	SearchQueryOperatorAny
	//SearchQueryOperatorStrictOrder ...
	SearchQueryOperatorStrictOrder
)

// Scan config structs :

// ScopeConfig provides configuration of scan
type ScopeConfig struct {
	Name                 string              `json:"name" xml:"name,attr"`
	StartQuery           string              `json:"startQuery" xml:"startQuery"`
	FinishQuery          string              `json:"finishQuery" xml:"finishQuery"`
	StartQueryCloseScope bool                `json:"startQueryCloseScope" xml:"startQueryCloseScope"`
	SearchQuery          []string            `json:"searchQuery" xml:"searchQuery"`
	SearchQueryMode      SearchQueryOperator `json:"searchQueryMode" xml:"searchQueryMode"`
}

// ScanConfig provides scan configuration
type ScanConfig struct {
	Folder string        `json:"folder" xml:"folder,attr"`
	Filter string        `json:"filter" xml:"filder,attr"`
	Scopes []ScopeConfig `json:"scopes" xml:"scopes"`
}

// Scan summary structs :

// FileScopeSummary provides...
type FileScopeSummary struct {
	FileName   string         `json:"fileName" xml:"fileName,attr"`
	Scopes     []ScopeSummary `json:"scopes" xml:"scopes"`
	AllMatches int
}

// MatchLine provides...
type MatchLine struct {
	Line  string `json:"line" xml:"line,attr"`
	Index int    `json:"index" xml:"index,attr"`
}

// ScopeSummary provides...
type ScopeSummary struct {
	Name          string   `json:"name" xml:"name,attr"`
	FileName      string   `json:"fileName" xml:"fileName,attr"`
	Started       int      `json:"started" xml:"started,attr"`
	Finished      int      `json:"finished" xml:"finished,attr"`
	Content       []string `json:"content" xml:"content"`
	ContentAsHTML []string
	Matches       []MatchLine `json:"matches" xml:"matches"`
}

// ScanSummary provides...
type ScanSummary struct {
	Folder       string
	Filter       string
	CreationTime time.Time
	Summary      []FileScopeSummary
	ScanFiles    int
}

// ScopeSummaryWithConfig provides...
type ScopeSummaryWithConfig struct {
	ScopeSummary ScopeSummary
	ScopeConfig  ScopeConfig
}

// ReadScopeConfiguration read ScanConfig from file
func ReadScopeConfiguration(configPath string) (ScanConfig, error) {

	jsonFile, err := os.Open(configPath)
	if err != nil {
		return ScanConfig{}, err
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		return ScanConfig{}, err
	}

	var scanConfig ScanConfig

	if json.Unmarshal(byteValue, &scanConfig) != nil {
		return ScanConfig{}, err
	}
	return scanConfig, nil
}

// -----------------------------------------------------------------------------
// extensions
// -----------------------------------------------------------------------------

// IsValid check if ScanConfig contains every required fields
func (cfg ScanConfig) IsValid() error {

	if cfg.Folder == "" {
		return errors.New("Empty folder")
	}

	if cfg.Filter == "" {
		return errors.New("Empty filter")
	}

	if len(cfg.Scopes) == 0 {
		return errors.New("Empty scopes")
	}

	for i, v := range cfg.Scopes {
		if v.Name == "" {
			return fmt.Errorf("Empty name of scope [%v]", i)
		}
		if len(v.SearchQuery) == 0 {
			return fmt.Errorf("Empty search queries of scope [%v]", i)
		}

		for j, q := range v.SearchQuery {
			if q == "" {
				return fmt.Errorf("Empty #%v search query of scope [%v]", j, i)
			}
		}

		if (v.StartQuery != "") && (v.FinishQuery == "") {
			return fmt.Errorf("Empty finish query of scope [%v]", i)
		}
		if (v.StartQuery == "") && (v.FinishQuery != "") {
			return fmt.Errorf("Empty start query of scope [%v]", i)
		}
	}

	return nil
}

// LogToFile writes summary to json file
func (s ScanSummary) LogToFile(p string) error {
	file, _ := json.MarshalIndent(s, "", " ")

	return ioutil.WriteFile(p, file, 0644)
}

// LogToHTML generate html log file
func (s ScanSummary) LogToHTML(p string) error {
	if s.Summary == nil {
		return errors.New("scan summary does not contains any summaries")
	}

	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	t, err := template.New("template").Parse(pattern)
	if err != nil {
		return err
	}
	err = t.Execute(f, s)
	if err != nil {
		return err
	}
	return nil
}
