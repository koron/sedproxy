# koron/sedproxy

[![GoDoc](https://godoc.org/github.com/koron/sedproxy?status.svg)](https://godoc.org/github.com/koron/sedproxy)
[![CircleCI](https://img.shields.io/circleci/project/github/koron/sedproxy/master.svg)](https://circleci.com/gh/koron/sedproxy/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/koron/sedproxy)](https://goreportcard.com/report/github.com/koron/sedproxy)

sedproxy is a HTTP reverse proxy, which rewrite HTML with regular expressions.

## Getting started

Go 1.13.1 or above required.

Install or updated sedproxy:

```console
$ go get -u -i github.com/koron/sedproxy
```

Start reverse proxy.

```console
$ sedproxy -target http://127.0.0.1:4000/ -messages messages.json
```

Let's access http://127.0.0.1:8000/


## Usage

```
sedproxy [OPTIONS]
```

### Options

* `-addr` - Address and port which reverse proxy to listen.
* `-target` - Target HTTP server to rewriting.
* `-messages` - JSON file for substitutions.
* `-structuredmessages` JSON file for substitutions with path limitation.

### Message JSON format

Example:

```json
[
  {
    "src": "店主([a-zA-Z]+)",
    "rep": "$1 the owner"
  },
  {
    "src": "ホーム",
    "rep": "Home"
  },
  {
    "src": "ニュース",
    "rep": "News"
  },
  {
    "src": "ブログ",
    "rep": "Blog"
  },
]
```

![](./sample.png)

### Structured Message JSON format

```json
{
  "": [
    {
      "src": "message#1 wich applied to globally",
      "rep": "xxx"
    },
    {
      "src": "message#2 wich applied to globally",
      "rep": "xxx"
    },
  ],
  "/foo" [
    {
      "src": "message#3 wich applied to only /foo.* (prefix match)",
      "rep": "xxx"
    },
    {
      "src": "message#4 wich applied to only /foo.* (prefix match)",
      "rep": "xxx"
    },
  ]
}
```
