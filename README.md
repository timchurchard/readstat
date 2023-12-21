# readstat

Attempt to collect and show reading statistics from ereaders like Kobo.

Proof of concept! I've been able to read the Kobo database and produce stats! I'll start to tidy-up now.

## Usage

Use the `sync` command to read the Kobo database. Syncing updates a local json file. And use the `stats` command to show stats.

```shell
./readstat sync -d ./testfiles/20231219/libra2/KoboReader.sqlite -s tc_readstat.json

./readstat stats -s ./tc_readstat.json
In 2023 / Tim
Finished books                  : 20
Finish books (words)            : 1,208,242
Time spent reading books        : 9 days 15 hours 51 minutes 19 seconds
```
