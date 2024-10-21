# mattermost-bot

## Build
```
go build -ldflags="-X 'main.mmToken=MATTERMOST_BOT_ACCESS_TOKEN'" -o "${BUILD_DIR}/" ./...
```

## Running
* Follow the instructions for setting up and running Mattermost with Docker.
  - https://docs.mattermost.com/install/install-docker.html#deploy-mattermost-on-docker-for-production-use
* `./build/mattermost-bot`
