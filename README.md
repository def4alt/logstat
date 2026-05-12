# logstat

A command-line tool that parses web server logs and gives summary statistics. Written in Go for a systems programming class project. Tested against the [NASA Kennedy Space Center web server logs](https://ita.ee.lbl.gov/html/contrib/NASA-HTTP.html) from July 1995.

## What it does

You give it a log file (or pipe one in), and it tells you:

- **Total entries**
- **Total bytes served**
- **Unique hosts**
- **Status code breakdown**
- **HTTP method counts**
- **Top hosts**
- **Top paths**
- **Response size percentiles**

It parses the [Common Log Format](https://en.wikipedia.org/wiki/Common_Log_Format):

```
199.72.81.55 - - [01/Jul/1995:00:00:01 -0400] "GET /history/apollo/ HTTP/1.0" 200 6245
```

## Installation

You need Go installed. Then:

```bash
go install github.com/def4alt/logstat/cmd/logstat@latest
```

Or clone and build:

```bash
git clone https://github.com/def4alt/logstat.git
cd logstat
go build -o logstat ./cmd/logstat
```

## Usage

### Read from a file

```bash
logstat --file ~/datasets/NASA_access_log_Jul95
```

### Pipe it in

```bash
cat ~/datasets/NASA_access_log_Jul95 | logstat
```

### Control output

```bash
# Show top 20 instead of default 10
logstat --file access.log --top 20

# Get JSON
logstat --file access.log --json

```

### Strict mode

By default, logstat skips malformed lines and keeps going. If you want it to
fail fast:

```bash
logstat --file access.log --strict
```


## Sample output (NASA July 1995 logs)

Running against the full 1.89 million requests:

```
Malformed lines skipped: 18
Total Entries:        1891697
Total Bytes:          38695973491
Unique Hosts:         81982

Status Code Counts:
200:  1701534
304:  132627
302:  46573
404:  10833
403:  54
500:  62
501:  14

Method Counts:
GET:   1887634
HEAD:  3952
POST:  111

Top Hosts:
piweba3y.prodigy.com:    17572
piweba4y.prodigy.com:    11591
piweba1y.prodigy.com:    9868
alyssa.prodigy.com:      7852
siltb10.orl.mmc.com:     7573

Top Paths:
/images/NASA-logosmall.gif:    111388
/images/KSC-logosmall.gif:     89639
/images/MOSAIC-logosmall.gif:  60468
/images/USA-logosmall.gif:     60014
/images/WORLD-logosmall.gif:   59489

Bytes Percentiles:
P50:  3635
P90:  46573
P95:  78588
P99:  283389
```

A few things jump out from this data:

- **The web was mostly images.** Five of the top six paths are GIF logos. The actual HTML pages are buried further down.
- **Prodigy users were obsessed with space.** The top four hosts are all Prodigy (`prodigy.com`) proxy/cache boxes. They account for over 46,000 requests on their own. Either Prodigy had a lot of space fans or they just didn't cache.
- **Almost everything was GET.** 188,7634 GETs vs 3,952 HEADs vs 111 POSTs. The 90s web was a read-only medium.
- **P50 response was 3.6 KB** but P99 was 283 KB.
- **Only 18 malformed lines** out of 1.89 million. The NASA logs are clean.

### JSON output

Same data, machine-readable:

```bash
logstat --file ~/datasets/NASA_access_log_Jul95 --json --top 3
```

```json
{
  "TotalEntries": 1891697,
  "SkippedEntries": 18,
  "TotalBytes": 38695973491,
  "UniqueHosts": 81982,
  "StatusCodeCounts": {
    "200": 1701534,
    "302": 46573,
    "304": 132627,
    "403": 54,
    "404": 10833,
    "500": 62,
    "501": 14
  },
  "MethodCounts": {
    "GET": 1887634,
    "HEAD": 3952,
    "POST": 111
  },
  "TopKHosts": [
    { "Key": "piweba3y.prodigy.com", "Value": 17572 },
    { "Key": "piweba4y.prodigy.com", "Value": 11591 },
    { "Key": "piweba1y.prodigy.com", "Value": 9868 }
  ],
  "TopKPaths": [
    { "Key": "/images/NASA-logosmall.gif", "Value": 111388 },
    { "Key": "/images/KSC-logosmall.gif", "Value": 89639 },
    { "Key": "/images/MOSAIC-logosmall.gif", "Value": 60468 }
  ],
  "P50Bytes": 3635,
  "P90Bytes": 46573,
  "P95Bytes": 78588,
  "P99Bytes": 283389
}
```


## How it works under the hood

1. **Parser** (`internal/parser/`): reads the file line by line, splits each one into fields while keeping track of quoted strings and brackets. It validates that status codes are three digits in the 1xx–5xx range and that each line has all 7 required fields.
2. **Stats** (`internal/stats/`): takes the parsed entries, counts everything, sorts the top K hits, and computes byte percentiles. Percentiles are calculated by sorting all byte values and picking the value at the appropriate index (no interpolation — it's a proper sample percentile).
3. **Output** (`internal/output/`): formats the summary either as a nicely aligned table (using Go's `tabwriter`) or as pretty-printed JSON.
4. **CLI** (`cmd/logstat/main.go`): `flag`-based argument parsing, wires everything together.

The whole thing is a single Go module with no external dependencies.

## Things I'd change if I had more time

- Support for gzipped log files would be nice (just transparent decompression)
- Configurable field delimiter for non-standard formats
- Maybe a `--since` / `--until` filter by timestamp
- Regex-based parsing so people could define their own log format
- Real-time mode with `inotify` or similar
- Breakdown by hour or day to see traffic patterns over time
