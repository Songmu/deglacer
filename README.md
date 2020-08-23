deglacer
=======

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![GoDoc](https://godoc.org/github.com/MH4GF/notion-deglacer?status.svg)][godoc]

[actions]: https://github.com/MH4GF/notion-deglacer/actions?workflow=test
[license]: https://github.com/MH4GF/notion-deglacer/blob/master/LICENSE
[godoc]: https://godoc.org/github.com/MH4GF/notion-deglacer

NotionリンクをSlack上で展開してくれる君

## Description

notino-deglacerはNotionのリンクがSlackに投稿された際に、それを展開してくれるSlack Appです。以下の機能を備えています。

- Notionの記事タイトル展開
  
note: Notionの非公式なAPIクライアントを利用しています。Notion側が意図しない利用方法のため、リスクを承知の上利用してください。

## Installation

1. Notionのアクセストークン取得
2. Slack App作成
3. notion-deglacerのデプロイ
4. 2で作ったappに、3のURLを登録する
5. Slack Appのbotユーザーをチャンネルに招待する

### 1. Notionアクセストークン取得

[この記事](https://presstige.io/p/Using-Notion-API-Go-client-2567fcfa8f7a4ed4bdf6f6ec9298d34a)の Accessing non-public pages を参考にしながらアクセストークンを取得してください。

### 2. Slack Appの作成

1. https://api.slack.com/apps の Create New App からアプリ作成
2. 左メニュー OAuth & Permissions を開き、Scopesでlink:writeを追加
3. 左メニュー Event Subscriptions を開く
    - App unfurl domains を展開し、 Add Domain で、 `www.notion.so` を入力し、Save Changes
4. 左メニュー Install App を開き、 Install App to Workspace -> Allow
5. OAuth Access Token が表示されるのでメモ (`SLACK_TOKEN`)
6. Basic Information を開き App CredentialsのSigning Secretをメモ (`SLACK_SIGNING_SECRET`)

※後で戻ってくるので、Slack Appの管理画面は開いたままにしておく。

### 3. deglacerのデプロイ

deglacerはGoで書かれたWebアプリケーションなので、任意の場所で簡単に動かせますが、HerokuやGoogle App Engineを利用するのがより簡単でしょう。動作のためには以下の環境変数が必要です。

- `NOTION_TOKEN`: 手順1で取得したNotionのアクセストークン
- `SLACK_TOKEN`: 手順2-5で取得したSlack Appのトークン
- `SLACK_SIGNING_SECRET`: 手順2-6で取得したリクエスト署名検証secret

#### Herokuで動かす場合

以下のボタンからデプロイできます。

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

### 4. 2で作ったappに、3のURLを登録する

- 左メニュー Event Subscriptions を開き、 Request URL に 3でデプロイしたdeglacerのURLを入力
- Verified と表示されたら Enable Events を On にして Save Changes

### 5. Slack Appのbotをチャンネルに招待する

Bot名は、左メニューのApp Homeから確認してください。

これで準備完了です。

## See Also

notion-deglacerは[Songmu](https://github.com/Songmu)さんのリポジトリをフォークして作られています。    
[https://github.com/Songmu/deglacer](https://github.com/Songmu/deglacer)  
Webサーバーの処理、Slackへの送信処理の大半をそのまま利用させていただいています。この場をお借りして御礼申し上げます。ありがとうございました！

## Author

[miya](https://github.com/MH4GF)
