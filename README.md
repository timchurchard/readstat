# readstat

![Build Status](https://github.com/timchurchard/readstat/workflows/Test/badge.svg)
![Coverage](https://img.shields.io/badge/Coverage-67.3%25-yellow)
[![License](https://img.shields.io/github/license/timchurchard/readstat)](/LICENSE)
[![Release](https://img.shields.io/github/release/timchurchard/readstat.svg)](https://github.com/timchurchard/readstat/releases/latest)
[![GitHub Releases Stats of readstat](https://img.shields.io/github/downloads/timchurchard/readstat/total.svg?logo=github)](https://somsubhra.github.io/github-release-stats/?username=timchurchard&repository=readstat)

Attempt to collect and show reading statistics from e-reader devicess, like [Kobo](https://uk.kobobooks.com/collections/ereaders).  This is a proof of concept using my two devices (Kobo Clara 2E and Libra 2 with database version 174).

## Usage

Use the `sync` command to read the Kobo database. Syncing updates a local json file. And use the `stats` command to make stats.

```shell
./readstat sync -d ./testfiles/20231219/libra2/KoboReader.sqlite -s tc_readstat.json

./readstat stats -s ./tc_readstat.json
In 2023 / Tim
Finished books                  : 20
Finish books (words)            : 1,208,242
Time spent reading books        : 9 days 15 hours 51 minutes 19 seconds
```
