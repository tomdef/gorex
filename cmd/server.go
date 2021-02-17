package cmd

import (
	"fmt"
	"gorex/pkg/utils"
	"html/template"
	"net/http"

	"github.com/spf13/cobra"
)

const (
	fPort   = "port"
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
			background: #D0E4F5;
			font-family: Verdana, Geneva, sans-serif;
			font-size: 14px;
			color: #000000;
			padding:5px;
		}
	
		.summary {
			border: 2px solid #1C6EA4;
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
		<div>
	</body>
	</html>	`
)

var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Run web server with gui",

		RunE: func(cmd *cobra.Command, args []string) error {

			logger := utils.CreateLogger("scan", trace)
			p, err := cmd.Flags().GetInt16(fPort)

			if err != nil {
				return err
			}

			http.HandleFunc("/scan", scanRequest)

			url := fmt.Sprintf(":%d", p)

			http.ListenAndServe(url, nil)

			return nil
		},
	}
)

// -----------------------------------------------------------------------------
// functions
// -----------------------------------------------------------------------------

func scanRequest(w http.ResponseWriter, req *http.Request) {
	t, err := template.New("template").Parse(pattern)
	if err != nil {
		return
	}
	t.Execute(w, nil)
}

func init() {

	serverCmd.Flags().Int16P(fPort, "p", 8080, "Web server port")
	rootCmd.AddCommand(serverCmd)
}
