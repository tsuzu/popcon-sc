# popcon-sc

## What is popcon-sc?
- Programming OPen(source) CONtest Server - SCalable
- popcon-scはオープンソースな競技プログラミング用コンテストサーバです。
- 主に部活内でコンテストを開催するために製作しています。
- Pure Go、WebはBootstrap3
- 前プロジェクト[popcon](https://github.com/cs3238-tsuzu/popcon)よりもDockerベースで扱いやすいシステムを目指しています。

## Features
- Scalable web server & judging system
- Easy to launch with Golang & Docker
- Support of multiple kinds of contests

## How to install
- install Docker
- $ git clone https://github.com/cs3238-tsuzu/popcon-sc.git
- $ cd popcon-sc && dockcer-compose build
- $ PP_MYSQL_PASSWORD="password" PP_TOKEN="password" docker-compose up

## License
- Under the MIT License
- Copyright (c) 2017 Tsuzu
