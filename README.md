# savepeek

I wanted to check my Stardew save from my phone without uploading the save to somebody else's website.

SavePeek reads the latest local save and serves a small read-only page for it.

```text
$ savepeek -lan
farm:  Cranberry
save:  Fall 18, Year 2
web:   http://192.168.1.42:8273
```

That's basically the whole idea.

## what it shows

- current season/day/year
- money
- player + farm name
- skill levels
- play time
- achievement count
- closest relationships

It rereads the save when the page refreshes, so there is no database or sync step.

## run it

You need Go 1.23+ for now:

```sh
git clone https://github.com/MrSecretMan/save-peek
cd save-peek
go run ./cmd/savepeek
```

By default it only listens on `127.0.0.1`.

To open it from another device on the same network:

```sh
go run ./cmd/savepeek -lan
```

If your save folder is somewhere weird:

```sh
go run ./cmd/savepeek -save-dir /path/to/StardewValley/Saves
```

## privacy

SavePeek never writes to a save file. It does not have analytics, accounts, a database, or any outside API calls.

`-lan` is opt-in because exposing local data to the whole LAN should not be the default.

## status

This is intentionally small. v0.1 is Stardew-only and focuses on getting one thing right before adding more games.
