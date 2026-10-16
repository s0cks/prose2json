# prose2json

A CLI tool for converting prose segments to JSON blobs using [Ollama](https://ollama.com/).

## Building

```sh
go build \
  -o prose2json \
  .
```

## Usage

```text
Convert prose segments to structured JSON

Usage:
  prose2json [prose segment] [flags]

Flags:
  -h, --help            help for prose2json
      --host string     The ollama host endpoint (default "http://localhost:11434")
  -m, --model string    The ollama model to use (default "nuextract:3.8b")
  -o, --output string   The output file to write the JSON to
  -s, --schema string   The schema to convert to
```

Example:

```sh
prose2json -s example/quote.json -o quote.result.json "$(cat example/quote.txt)"
```

## Credits

N/A

## License

See [LICENSE](/LICENSE)
