# CLAUDE.md — AI Assistant Guide for gorond

## Communication Language

**日本語で会話すること。** ユーザーへの返答・説明・質問はすべて日本語で行う。コード・コメント・コミットメッセージは英語でも可。

---

## コードマップ（インデックス）

| パス | 役割 |
|------|------|
| `main.go` | エントリポイント・CLI フラグ |
| `goron/goron.go` | cron スケジューラ・ジョブ実行・シグナル処理・自動リロード |
| `goron/notify.go` | 通知チャンネルへのディスパッチ |
| `config/config.go` | Config 構造体・`LoadConfig()` |
| `config/config_parser.go` | INI 形式パーサ |
| `config/job_parser.go` | ジョブファイルパーサ（逐次・並列構文） |
| `logging/logger.go` | レベル付きロガー |
| `webapi/webapi.go` | REST API サーバ（`/statuses`, `/jobs`） |
| `fswatch/fswatch.go` | 設定ファイル変更監視 |
| `notify/*.go` | 各通知チャンネル実装（email/slack/sns/fluentd/stdout） |
| `util/pid.go` | PID ファイル管理 |
| `Makefile` | ビルド・テスト・RPM 等のタスク定義 |
| `README.md` | ユーザ向けドキュメント（日本語） |
| `.travis.yml` | CI パイプライン定義 |
| `config/config_test.d/` | config パッケージのテスト用フィクスチャ |
| `webapi/config_test.d/` | webapi パッケージのテスト用フィクスチャ |
| `rpmbuild/` | RPM パッケージング |

---

## 非自明なナレッジ

### テストのモック方法

`SystemCommand`（`goron/goron.go`）は **インターフェースではなく `var` の関数変数** として宣言されている。テストではこの変数を直接差し替えてシェル実行なしにジョブ結果をシミュレートする。新しく外部コマンドを呼ぶ処理を追加する場合も同じパターンにすること。

```go
// テスト内でのモック例
goron.SystemCommand = func(command string, args ...string) ([]byte, int, error) {
    return []byte("ok"), 0, nil
}
```

### テストフィクスチャの置き場所

テスト用の設定ファイルは `*_test.d/` ディレクトリに置く（例: `config/config_test.d/`）。テストはグローバル状態を持たず、フィクスチャを直接ロードして使う。

### ジョブの逐次・並列実行の仕様

インデントと先頭 `-` で親子関係を表現する。**親が成功した場合のみ子が実行される。** 同じ深さの子は並列実行、さらに深い子は全並列ジョブの完了後に実行される。

```
0 0 4 * * *  root  first_command
             - root  second_command1    # first 成功後に並列起動
             - root  second_command2    # first 成功後に並列起動
               - root  third_command   # second 両方成功後に実行
```

### `%` はエスケープ不要

標準 cron では `%` を `\%` とする必要があるが、gorond では不要。

```
0 0 * * * *  root  date +%Y-%m-%d   # そのまま書いてよい
```

### ユーザ切り替えは常に `su -` 経由

ジョブは必ず `su - <user> -c <command>` 経由で実行される（`goron/goron.go: executionCommand()`）。セキュリティ上の理由でこの仕組みをバイパスしてはならない。

### ロガーの使い方

`log` パッケージを直接使わず、`logging.Logger` ラッパーを使う。`goron/` パッケージ内では `logger`（デーモンログ）と `cronlogger`（cron 実行ログ）の2本を使い分ける。

### コメントの言語

既存のソースコメントは日本語で書かれている。新規コードは英語でも可。

### 設定ホットリロードの挙動

`.conf` ファイルの変更を `fswatch` が検知すると、**cron を一度 Stop → 設定再パース → 新しい cron で Start** という順序で行われる（`goron.StartAutoReload()`）。リロード中に実行中のジョブがあっても Stop は即時。

### ジョブステータスの状態遷移

`WAITING` → `RUNNING` → `WAITING`（正常終了）または `FAILED`（非ゼロ終了）。`FAILED` は次回実行トリガーまでその状態を保持する。
