TeamSpirit 打刻
===============

![screen](https://ja.ngs.io/images/2018-02-14-ts-dakoku/screen.gif)

Slack のコマンドで TeamSpirit の打刻をします

[![Docker Automated build](https://img.shields.io/docker/automated/atsnngs/ts-dakoku.svg?maxAge=2592000)](https://hub.docker.com/r/atsnngs/ts-dakoku/)
[![CircleCI](https://circleci.com/gh/ngs/ts-dakoku.svg?style=svg&circle-token=9c154b7114e81b3ed97b85121e98c7ee5a9ad23c)](https://circleci.com/gh/ngs/ts-dakoku)
[![Coverage Status](https://coveralls.io/repos/github/ngs/ts-dakoku/badge.svg?branch=master)](https://coveralls.io/github/ngs/ts-dakoku?branch=master)

導入手順
-------

詳細は [ブログ記事](https://ja.ngs.io/2018/02/14/ts-dakoku/) を参照してください。

### Heroku

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

上のボタンをクリック、もしくは、以下のコマンドを実行

```sh
mkdir -p ~/.go/src/github.com/ngs
cd ~/.go/src/github.com/ngs

git clone git@github.com:ngs/ts-dakoku.git
cd ts-dakoku

heroku create
heroku addons:create heroku-redis:hobby-dev

heroku config:set \
  SALESFORCE_CLIENT_ID=${SALESFORCE_CLIENT_ID} \
  SALESFORCE_CLIENT_SECRET=${SALESFORCE_CLIENT_SECRET} \
  SLACK_VERIFICATION_TOKEN=${SLACK_VERIFICATION_TOKEN} \
  TEAMSPIRIT_HOST=${TEAMSPIRIT_HOST}

git push heroku master
```

### Docker

```sh
docker pull redis
docker pull atsnngs/ts-dakoku

docker run --name ts-dakoku-redis -d redis
docker run --name ts-dakoku -p 8000:8000 -d --rm \
  --link ts-dakoku-redis:redis \
  -e SALESFORCE_CLIENT_ID=${SALESFORCE_CLIENT_ID} \
  -e SALESFORCE_CLIENT_SECRET=${SALESFORCE_CLIENT_SECRET} \
  -e SLACK_VERIFICATION_TOKEN=${SLACK_VERIFICATION_TOKEN} \
  -e TEAMSPIRIT_HOST=${TEAMSPIRIT_HOST} \
  -e REDIS_URL="redis://redis:6379" \
  atsnngs/ts-dakoku
```

環境変数
--------

| Name                       | Description                                  |
| :------------------------- | :------------------------------------------  |
| `SALESFORCE_CLIENT_ID`     | 接続アプリケーションのコンシューマ鍵         |
| `SALESFORCE_CLIENT_SECRET` | 接続アプリケーションのコンシューマ秘密鍵     |
| `SLACK_VERIFICATION_TOKEN` | Slack アプリケーション の Verification Token |
| `TEAMSPIRIT_HOST`          | TeamSpirit のホスト名                        |

Author
======

[Atushi Nagase]

License
=======

Copyright &copy; 2018 [Atushi Nagase]. All rights reserved.

[Atushi Nagase]: https://ngs.io/
