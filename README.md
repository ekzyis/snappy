# snappy

<img src="https://upload.wikimedia.org/wikipedia/en/0/0c/Schnappi.png" width="128" height="128" />

The Go library that powers [@hn](https://stacker.news/hn), [@nitter](https://stacker.news/nitter)
and [@ctfbot_](https://stacker.news/ctfbot_) on [Stacker News](https://stacker.news/).

## How to use

As library:

```
$ go get github.com/ekzyis/snappy
```

```go
import sn "github.com/ekzyis/snappy"
```

As command:

```
$ go install github.com/ekzyis/snappy/cmd/snappy
$ snappy
  ___ ___  ___ ____  ___  __ __
 (_-</ _ \/ _ `/ _ \/ _ \/ // /
/___/_//_/\_,_/ .__/ .__/\_, /
             /_/  /_/   /___/

Usage:
  snappy <command> [options]

Commands:
  snappy query -author <username>      Query all items of a user.
  snappy query -territory <territory>  Query all items of a territory.
  snappy query -item <id>              Query a single item by id.

Options:
  -type string
      Items to query: all, posts, or comments (default: posts)
  -limit int
      how many items to fetch (default: 100)
```

`SN_API_KEY` must be set in your environment for authenticated API access.

## How to test

```
$ go test ./...
```

The tests run against an in-process mock GraphQL server. Every query the client sends is validated
against Stacker News' real schema (`client/testdata/schema.graphql`), and the mock replies with the
fixture for that operation from `client/testdata/fixtures/`. A request whose operation has no fixture
fails the test.

The fixtures are curated to represent the API shape with fixed, stable values; they are not verbatim
API responses, so tests assert structure rather than exact values.

Regenerate the schema after bumping the submodule:

```
$ make schema
```

Check the fixtures against the live API:

```
$ make fixtures
```

