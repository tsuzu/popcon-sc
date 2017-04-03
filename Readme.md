# popcon-sc

## What is popcon-sc?
- Programming OPen(source) CONtest Server - SCalable
- popcon-sc is an open-source contest management system for competitive programming.
- Mainly, made to be used in my club activities.
- Main: Pure Go
- Web: Bootstrap3
- I'll make this more confortable than the previous project, [popcon](https://github.com/cs3238-tsuzu/popcon) by using Docker.

## Features
- Scalable web server & judging system
- Easy to launch with Golang & Docker
- Support of multiple kinds of contests

## How to install
- install Docker
- $ git clone https://github.com/cs3238-tsuzu/popcon-sc.git
- $ cd popcon-sc && dockcer-compose build
- Prepare pp_data/ppweb/setting.json and pp_data/sendmail/config.json
- $ PP_DB_PASSWORD="password" PP_TOKEN="password" docker-compose up

## License
- Under the MIT License
- Copyright (c) 2017 Tsuzu
