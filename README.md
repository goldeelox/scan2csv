# scan2csv
Input delimited by new line is written to a CSV file. Each line contains two columns.
1. datetime
1. input text

Download [latest release](https://github.com/goldeelox/scan2csv/releases)

## Usage
```bash
Usage of scan2csv:
  -dated-file
    	Include date in output file name (e.g., scans_1970-01-01.csv) (default true)
```

## Example
```bash
# make it executable
$ chmod +x ./scan2csv

$ ./scan2csv
2026/01/18 14:23:24 INFO ready to scan
12345
2026/01/18 14:23:29 INFO  input=12345
asdf
2026/01/18 14:23:33 INFO  input=asdf

$ cat ./scans_2026-01-18.csv
2026-01-18 14:23:29,12345
2026-01-18 14:23:33,asdf
```
