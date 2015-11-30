
# これは何？

Goron ( Go cron ) は、なにかと問題を起こすことが多い cron をgo で再実装したものです。
イメージとしては cron に以下の機能を実装したものです。

* コマンド実行時のエラー通知
* cronライクな記述で複数コマンドの逐次・並列実行
* 外形監視用WebAPI
* cronで問題になりがちなバッドノウハウの対応
  * %はそのまま記述してOK( \`date %Y-%m\` は cron ではNG )
  * 標準出力もログに残る


# 設定

設定ファイル配置場所。(起動時の-c [path]オプションで変更可能)

* /etc/goron.conf

```
[config]
# Web APIのポートを設定します。(デフォルトではコメントアウトしており無効です。セキュリティに注意して設定してください。)
webApi = localhost:6777

# ログファイルのパスを設定します。
log = /var/log/gorond/goron.log

# Cron実行時の出力ログファイルのパスを設定します。
cronLog = /var/log/gorond/cron.log

# apiサーバのアクセスログを記録するログファイルのパスを設定します。
apiLog = /var/log/gorond/api.log

# 通知は mail、fluentd、sns、slack、stdout の5種類から選べます
notifyType = { mail | fluentd | sns | slack | stdout }

# 通知のタイミングを onerror(ステータスコード!=0) と always(常に) の2種類から選べます。
notifyWhen = { onerror | always }


[mail]
# notify.type = mail の場合、以下の設定が有効になります。
# デフォルトのアラートメール通知先。
dest = alert@example.com

# デフォルトの送信者URL。
from = from@example.com

# デフォルトのSMTP送信先。
smtpHost = localhost:25

# SMTPユーザ・パスワード。
smtpUser = username
smtpPassword = password


[fluentd]
# notifytype = fluentd の場合、以下の設定が有効になります。

# デフォルトの通知先 fluentd のエンドポイント
url = http://endpoinst:8888/tag


[sns]
# notifytype = sns の場合、以下の設定が有効になります。

# sns topicのリージョンを指定します。
region = ap-northeast-1

# デフォルトの sns 通知先。
topicArn = "arn:aws:sns:ap-northeast-1:0000000000000:app/EXAMPLE/sample"

[slack]
# notifytype = slack の場合、以下の設定が有効になります。

# ポストするチャンネル。'#'なしで記述します。
Channel = xxxxx

# slackのwebhookurlを取得して貼り付けます。
WebhookUrl = http://xxxxxxxxxxxxxx/xxxx

# botのアイコンを変更する場合、アイコンのURLを指定します。
IconUrl = http://xxxxxxxxx/xxxxxxxx
```


# cron ライクな設定

設定ファイル配置場所。-d [path] で変更できます。

* /etc/goron.d/\*.conf

cronライクなフォーマットで定期処理ファイルを記述します。

```
# 基本は github.com/robfig/cron の仕様に依存します。

# 設定例
0 0 4 * * THU root command
0 0 4 * * * root command
0 4 * * * * user command
@daily      root command
```

# オプショナルな使い方

## ファイル内のみ設定を変更する。

/etc/goron.d/sample1.conf 内に上書きしたい設定を記述します。

    [config]
    notifyType = email

    [mail]
    dest = alert@example.com
    from = info@example.com
    smtpHost = localhost:25

    [job]
    0 0 4 * * * root command

## 複数のコマンドを逐次実行させる場合の記述

first_commandが正常に完了した場合、 second_command1とsecond_command2を同時に実行開始します。  
いずれも正常に終了した場合、third_command を実行します。

```
0 0 4 * * * root first_command
            - root second_command1
            - root second_command2
              - root third_command
```

# 外形監視用 Web API

各コマンド実行結果のステータスを取得する。

```
$ curl http://localhost:6777/statuses
{
  "app1.conf": {
    "root first_command": "waiting",
    "root second_command1": "running",
    "root second_command2": "running",
    "root third_command": "waiting"
  },
  "app2.conf": {
    "user command": "failed"
  }
}
```

confファイルの内容を返します。

```
$ curl http://localhost:6777/jobs
{
  "app1.conf": [
    "0 0 4 * * * root first_command",
    "            - root second_command1",
    "            - root second_command2",
    "            - root third_command"
  ],
  "app2.conf": [
    "0 4 * * * * user command"
  ]
}
```

# ライセンス

MITライセンスに準拠します。


# 免責

本ソフトウェアによって起こったいかなる事象についても制作者は責任を負いません。  
すべて自己責任にてご利用ください。


