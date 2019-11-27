# koron/sedproxy

[![GoDoc](https://godoc.org/github.com/koron/sedproxy?status.svg)](https://godoc.org/github.com/koron/sedproxy)
[![CircleCI](https://img.shields.io/circleci/project/github/koron/sedproxy/master.svg)](https://circleci.com/gh/koron/sedproxy/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/koron/sedproxy)](https://goreportcard.com/report/github.com/koron/sedproxy)

sedproxy is a HTTP reverse proxy, which rewrite HTML with regular expressions.

![](./sample.png)

## Getting started

Go 1.13.3 or above is recommended.

To install or update sedproxy:

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

* `-accesslog` - Output access log to STDERR (default disabled)
* `-addr` - Address and port which reverse proxy to listen.
* `-messages` - JSON file for substitutions.
* `-target` - Target HTTP server to rewriting.

    `SEDPROXY_TARGET` environment variable can be used, instead of this.

### Message JSON format

sedproxyはメッセージファイルで「置き換え設定」を定義します。

メッセージファイルは1つの配列、すなわち複数の「ベージ置き換え」でなっています。

```json
[
  // 置き換え設定
  {
    // ページ書き換え1
  },
  {
    // ページ書き換え2
  },
  // ...
]
```

1つの「ページ置き換え」はJSON Objectで `mediaTypes`, `path` そして `items` の3
つのプロパティで構成されます。 

`mediaType` プロパティは文字列の配列で、`Content-Type` のMIMEタイプにマッチする
ものが配列にある場合にこの「ページ置き換え」を適用します。省略した場合は
`text/html` にだけ適用します。

`path` プロパティは文字列で正規表現を指定します。省略はできません。この正規表現
にマッチするパスのコンテンツが、この「ページ置き換え」の対象となります。

`mediaType` と `path` の両方を指定した場合は両方共にマッチするページが対象とな
ります。

`items` プロパティは後に説明する「置き換えアイテム」の配列です。

「ページ置き換え」の設定例:

```json
{
  // 全部のパス(かつ media type が text/html) に適用する置き換え
  "path": "^.*",
  "items": [
    // ...
  ]
},

{
  // JavaScript (拡張子が .js かつ media type がいずれかに合致)に
  // 適用する置き換え
  "path": "\\.js$",
  "mediaType": [
    "application/javascript",
    "text/javascript",
    "application/x-javascript"
  ],
  "items": [
    // ...
  ]
}
```

「置き換えアイテム」は `src`, `srcRx` そして `repl` の3つの文字列プロパティから
構成されます。そのうち `repl` は必須です。 `src` と `srcRx` はどちらか片方だけ
指定してください。両方指定した場合は `src` を優先します。

`src` の文字列は完全一致に用います。ページ内の一致した箇所をすべて `repl` の文
字列に置き換えます。

`srcRx` の文字列は正規表現として用います。ページ内でこの正規表現に一致した箇所
をすべて `repl` の文字列に置き換えます。 `repl` 内の `$1` は `srcRx` でキャプ
チャした文字列に展開するので、部分的に可変な文字列を置き換えるのに利用できま
す。この展開について詳しくは <https://golang.org/pkg/regexp/#Regexp.Expand> を
参照してください。

「置き換えアイテム」の設定例:

```json
[
  // 文字通りの置き換え
  {
    "src":  "software",
    "repl": "ソフトウェア"
  },

  // 正規表現とキャプチャを利用した置き換え
  // 一例として `Total cost: $987` を `総コスト: USD 987` に置き換える。
  {
    "srcRx": "Total cost: \$(\d\+)",
    "repl":  "総コスト: USD $1",
  },
]
```

メッセージファイルで利用する正規表現の振る舞いや表記は、すべてGo言語の正規表現
エンジンに従います。詳しくは <https://golang.org/pkg/regexp/syntax/> を参照して
ください。

#### Format definition

The schema of message file is (in [JSON schema][jsonschema] in YAML format):

```yaml
type: array
items:

  title: Substitution Group
  description: |
    A substitution group is a pair of `mediaTypes`, `path` and `items`.

    A substitution group will be evaluated when both `mediaTypes` and `path`
    conditions are passed.

    Actual substitutions are in `items`.
  type: object
  properties:
    mediaTypes:
      type: array
      items:
        type: string
      description: |
        Array of media types which this substitution group is applied to.
        Media type is core part of `Content-Type`.
        For example, the media type of `text/html; charset=utf8` is `text/html`.
        When one of the media types was matched with `Content-Type`,
        this substitution group will be evaluated.
        When this is omited, only `text/html` will be used as default.
        So you should put `text/javascript`, if you want to apply a
        substitution group to JavaScript.
    path:
      type: string
      description: |
        A regexp pattern to match with "path" of HTTP request.

        See <https://golang.org/pkg/regexp/syntax/> for the syntax.
    items:
      type: array
      items:

        title: A substitution pair.
        description: |
          A pair of regexp pattern and replacement text.
        type: object
        properties:
          srcRx:
            type: string
            description: |
              A regexp pattern to match with "body" of HTTP response.
              See <https://golang.org/pkg/regexp/syntax/> for the syntax.
          src:
            type: string
            description: |
              A literal string to replace.
          repl:
            type: string
            description: |
              Replacement text.  Captured texts can be refered by `$1` or so.
              See <https://golang.org/pkg/regexp/#Regexp.Expand> for the
              expansion.
```

See <./messages.json> for example.
