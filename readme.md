# easy vault api server

## docker build
```
docker rmi -f easy_vault:internal
docker build -f ./docker/Dockerfile.internal -t easy_vault:internal .
```

## local env
```
docker stop easy_vault
docker rm easy_vault
docker run -p 8081:8081 -p 9443:9443 --name easy_vault easy_vault:internal 
docker run -p 3311:3306 --name mysql8 -e MYSQL_ROOT_PASSWORD=123456 -d mysql:8.1.0 --default-authentication-plugin=mysql_native_password --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
docker run --name local-redis-7.2 -p 6379:6379 -d redis:7.2
	
# 如果要連線本機 .env裡的ip要換成host.docker.internal
# http://localhost:8081/test/swagger/index.html
```

## docker compose
```
docker-compose -f docker/docker-compose.yml up -d
```

## 備註

1. stringer
  ```
  go get golang.org/x/tools/cmd/stringer
  ```

2. swaggo
  ```
  vim .bash_profile

  export GOPATH=$HOME/go
  export PATH=$GOPATH/bin:$PATH
  export GOBIN=$GOPATH/bin

  renew bash_profile
  source .bash_profile

  go install github.com/swaggo/swag/cmd/swag@latest
  swag init --parseDependency
  ```

3. ngrok
  ```
  brew install ngrok/ngrok/ngrok
  ngrok config add-authtoken 2i26gVbSZel7HkdRgiWfjZnzu3g_3aSgzXxPepPXhuAxztA2G
  ngrok http http://localhost:9487

  docker run -it -e NGROK_AUTHTOKEN=2i26gVbSZel7HkdRgiWfjZnzu3g_3aSgzXxPepPXhuAxztA2G ngrok/ngrok http 9487
  ```

4. gocron
  ```shell
  docker run --name gocron -p 5920:5920 -d ouqg/gocron
  ```
  https://127.0.0.1:5920
  ```shell
  # 取得本機ip
  ipconfig getifaddr en0 
  ```
