# snappy

<p align="center">
<img src="https://stacker.news/favicon.png" width="64" height="64" />
<img src="https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png" width="64" height="64" />
</p>

<p align="center">A Go client for the <a href="https://stacker.news" target="_blank">Stacker News</a> GraphQL API</p>

## How to use

```
$ go get github.com/ekzyis/snappy
```

`SN_API_KEY` must be set in your environment for authenticated API access.

## How to test

1. Run SN
2. Set `TEST_SN_BASE_URL` and `TEST_SN_API_KEY` in .env
3. Run `go test`
