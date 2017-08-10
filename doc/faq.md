# FAQ

## for any platforms
### Redis initialization error in ppweb
- Have you written .env file?
- You must write .env file in the same directory as docker-compose.yml
- Like the following
```
PP_TOKEN=test_token
PP_DB_PASSWORD=test_password_for_db
```

## for Windows
### Installation failed by character encoding of Python
- In cmd or powershell, execute "chcp 936" before docker-compose up -d

### Launching failed by "permission denied" for traefik
- Some server applications or the kernel are using :80
- There are two approaches
    - Stop the applications
    - Edit docker-compose.yml and change the port from 80:80 to (any port):80 in traefik
