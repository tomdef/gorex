# README #

Simple tool to find text "scopes" in files. Scope is a block of text with begin and end. For example:
```
<begin>
text
<end>
```
or
```
BEGIN OF SCOPE
// comment
SOME TEXT
// comment
END OF SCOPE
```

## Commands: ##

### scan ###

Execute scan command with json file. Json file is a recipe to find scope(s).

#### scan flags ####

| Flag | Description |
| --- | --- |
|  ``-h``, ``--help`` | Help for scan |
|  ``-i``, ``--input string`` | Input file path (*.json) with scan commands |
|  ``-d``, ``--outputdata string`` | Output raw data in json format |
|  ``-o``, ``--outputhtml string`` | Output html report (preffered) |
|  ``-s``, ``--show`` | Show result after scan (only for html report) |
|  ``-t``, ``--trace`` | Set trace mode |

Json file:

```
{
	"folder": ".\\example",
	"filter": "F1.txt",
	"scopes": [
		{
			"name": "example-find-any-command-in-scope",
			"startQuery": "^\\W*BEGIN\\W*$",
			"finishQuery": "^\\W*(END)\\W*$",
			"startQueryCloseScope": true,
			"searchQuery": [
				"^\\s*COMMAND\\=.*$"
			],
			"searchQueryMode": 1
		}
	]
}
```
| Field | Description |
| --- | --- |
|  ``folder`` | folder to scan |
|  ``filter`` | files filter |
|  ``scopes`` | List of scopes |
|  ``scopes\name`` | Name of the scope |
|  ``scopes\startQuery`` | Regular expression to find start of the scope |
|  ``scopes\finishQuery`` | Regular expression to find end of the scope |
|  ``scopes\startQueryCloseScope`` | ``true`` means if line match to startQuery then currend find scope is closed and new is open |
|  ``scopes\searchQuery`` | List of queries to find in scope (between start and finish lines) |
|  ``scopes\searchQueryMode`` | Mode of search queries : ``1`` - all queries should be exists in scope, ``2`` - any query should be exists, ``3`` - all queries should be exists in strict order |



#### Usage ####

* Scan file(s) and generate html report:
```
.\gorex.exe scan --input .\example.json --outputhtml .\example.html
```

* Scan file(s), generate and open html report:
```
.\gorex.exe scan --input .\example.json --outputhtml .\example.html --show
```

### gen ###

Generate example input file for scan command

### Usage ###

* Generate example input file (example.json):
```
.\gorex.exe gen
```


