deglacer
=======

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![GoDoc](https://godoc.org/github.com/Songmu/deglacer?status.svg)][godoc]

[actions]: https://github.com/Songmu/deglacer/actions?workflow=test
[license]: https://github.com/Songmu/deglacer/blob/master/LICENSE
[godoc]: https://godoc.org/github.com/Songmu/deglacer

KibelaリンクをSlack上で展開してくれる君

## Description

deglacerはkibelaのリンクがSlackに投稿された際に、それを展開してくれるSlack Appです。以下の機能を備えています。

- kibelaの記事の展開
- kibelaのコメントの展開

deglacerは木べらを用いる調理手法である、déglacer(デグラッセ)から命名しました。

## Installation

1. Kibelaアクセストークン取得
2. Slack App作成
3. deglacerのデプロイ
4. 2で作ったappに、3のURLを登録する
5. Slack Appのbotユーザーをチャンネルに招待する

### 1. Kibelaアクセストークン取得

以下のURLからアクセストークンを作成します(`KIBELA_TOKEN`)。権限はreadのみで大丈夫です。

<https://my.kibe.la/settings/access_tokens>

### 2. Slack Appの作成

1. https://api.slack.com/apps の Create New App からアプリ作成
2. 左メニュー OAuth & Permissions を開き、Scopesでlink:writeを追加
3. 左メニュー Event Subscriptions を開く
    - App unfurl domains を展開し、 Add Domain で、 {`KIBELA_TEAM`}.kibe.la を入力し、Save Changes
4. 左メニュー Install App を開き、 Install App to Workspace -> Allow
5. OAuth Access Token が表示されるのでメモ (`SLACK_TOKEN`)
6. Basic Information を開き App CredentialsのVerification Tokenをメモ (`SLACK_VERIFICATION_TOKEN`)

※後で戻ってくるので、Slack Appの管理画面は開いたままにしておく。

### 3. deglacerのデプロイ

deglacerはGoで書かれたWebアプリケーションなので、任意の場所で簡単に動かせますが、HerokuやGoogle App Engineを利用するのがより簡単でしょう。動作のためには以下の環境変数が必要です。

- `KIBELA_TEAM`: Kibelaのチーム名
- `KIBELA_TOKEN`: 手順1で取得したKibelaのアクセストークン
- `SLACK_TOKEN`: 手順2-5で取得したSlack Appのトークン
- `SLACK_VERIFICATION_TOKEN`: 手順2-6で取得したリクエスト検証トークン

#### Herokuで動かす場合

以下のボタンからデプロイできます。

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

#### Google App Engineで動かす場合

1. 事前にgcloudコマンドのインストールやGCP上のアプリの作成をおこなう
    - ref. https://cloud.google.com/appengine/docs/standard/go/quickstart#before-you-begin
2. 当リポジトリをcloneする
3. `secret.yaml` に設定項目を記述する(`secret.yaml.example`を参考に)
4. `gcloud app deploy` でデプロイ

### 4. 2で作ったappに、3のURLを登録する

- 左メニュー Event Subscriptions を開き、 Request URL に 3でデプロイしたdeglacerのURLを入力
- Verified と表示されたら Enable Events を On にして Save Changes

### 5. Slack Appのbotをチャンネルに招待する

Bot名は、左メニューのApp Homeから確認してください。

これで準備完了です。

## See Also

deglacerはotofune/slack-unfurl-kibelaのRuby実装を参考に移植しました。設定にあたっては、higebuさんのエントリが非常に参考になりました。

- https://github.com/higebu/slack-app-unfurl-kibela
    - https://www.higebu.com/blog/2019/12/04/slack-app-unfurl-kibela/
- https://github.com/otofune/slack-unfurl-kibela

## Author

[Songmu](https://github.com/Songmu)
