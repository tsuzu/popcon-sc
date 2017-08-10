## Installation

### Prerequirements
- Docker
- wget or curl
- Vim(not necessarily)

### Docker Compose 
- cd path/to/any/directory
- Download docker-compose.yml from Github
    - $ wget "https://raw.githubusercontent.com/cs3238-tsuzu/popcon-sc/master/docker-compose.yml"
- Create .env file
    - $ curl > .env
    - PP_TOKEN=test_token
    - PP_DB_PASSWORD=test_password_for_db
    - Ctrl-C
- Start Docker containers
    - $ docker-compose up -d
- Get admin's password
    - $ docker-compose logs -f | grep Pass:
    - When you get it, stop the command by Ctrl-C
- Open localhost with browsers and signin the server

### Docker Swarm
- Deprecated
- Some bugs still happen, so not prepared yet.
