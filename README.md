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
  -detect-removable-disks
    	Scans for attached removable disks then exits
  -output-dir string
    	Directory to write CSV files to (default ".")
  -uuid string
    	UUID of removable disk used to backup scan files (Attach removable disk and run with -detect-removable-disks to detect and list UUIDs)
```

## Example
```bash
# make it executable
$ chmod +x ./scan2csv

$ ./scan2csv
2026/01/20 22:28:42 INFO ready to scan
2026/01/20 22:28:44 INFO writing to file path=./scans_2026-01-20.csv input="asdf"
2026/01/20 22:28:45 INFO writing to file path=./scans_2026-01-20.csv input="1234"

$ cat scans_2026-01-20.csv
2026-01-20 22:28:44,asdf
2026-01-20 22:28:45,1234
```
