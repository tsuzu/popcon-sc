# popcon-sc
[![Build Status](http://img.shields.io/travis/cs3238-tsuzu/popcon-sc/swarm.svg?style=flat-square)](https://travis-ci.org/cs3238-tsuzu/popcon-sc)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg?style=flat-square)](./LICENSE)

## What is popcon-sc?
- Programming OPen(source) CONtest Server - SCalable
- popcon-sc is an open-source contest management system for competitive programming.
- Mainly, made to be used in my club activities.
- Main: Pure Go
- Web: Bootstrap3
- I'll make this more useful than the previous project, [popcon](https://github.com/cs3238-tsuzu/popcon) by using Docker.

## Features
- Scalable web server & judging system
- Easy to launch with Golang & Docker
- Support of multiple kinds of contests

## How to install
- Requirements: Docker

### For docker-compose
- install Docker
- $ git clone https://github.com/cs3238-tsuzu/popcon-sc.git
- $ cd popcon-sc && dockcer-compose build
- Prepare pp_data/ppweb/setting.json and pp_data/sendmail/config.json
- $ PP_DB_PASSWORD="your password" PP_TOKEN="your password" docker-compose up

## License
- Under the MIT License
- Copyright (c) 2017 Tsuzu
