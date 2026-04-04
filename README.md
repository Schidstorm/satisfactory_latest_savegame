# Satisfactory Latest Savegame Server

A simple HTTP server that serves the latest Satisfactory savegame file from a designated directory.

Used by [SC-InteractiveMap](https://github.com/AnthorNet/SC-InteractiveMap) to fetch the current game save for visualization and analysis.

## Installation

```bash
go install github.com/Schidstorm/satisfactory_latest_savegame@latest
```

## Usage

```bash
satisfactory_latest_savegame \
  --listen.address ":8080" \
  --savegame.dir "/path/to/savegames"
```

### Flags

- `--listen.address` — HTTP listen address (default: `:8080`)
- `--savegame.dir` — Directory containing Satisfactory savegames (default: `/home/steam/.config/Epic/FactoryGame/Saved/SaveGames/server/`)

## Endpoints

### GET /latest
Downloads the latest `.sav` file from the configured directory.

**Response Headers:**
- `Content-Type: application/octet-stream`
- `Cache-Control: no-cache`
- `ETag` — File hash for conditional requests
- `Access-Control-Allow-Origin: https://satisfactory-calculator.com`

**Status Codes:**
- `200` — Latest savegame found and streamed
- `304` — Savegame unchanged (if `If-None-Match` or `If-Modified-Since` matches)
- `404` — No savegames found in the directory
- `500` — Server error

### OPTIONS /latest
Returns CORS headers without streaming a file body.

**Status Code:**
- `204` — No Content

## License

See [LICENSE](LICENSE) file.
