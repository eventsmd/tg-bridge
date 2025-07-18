# tg-bridge

It is a bridge for telegram internal API https://core.telegram.org/ (there is no way to get messages from telegram via bot api)

Communal organizations publish their events on telegram channels, messengers are source of truth for these events.

Communal outages messages could be parsed from these channels for the first time:

* @vodokanalpmrcom
* @eresofficial

The service gets messages from configured channels with configured interval and saves in external db as state last message id from parsed channel.

There could be configured more than one channel because of restrictions of telegram api.

After parsing, it saves original message with metadata to db and push it to queue for further processing.

# How to run the application
```go
go run tgbridge/cmd/server/main.go
```

# How to generate Telegram session
Application is using cached Telegram session which is provided from environment variables.
Session is base64 encoded JSON which is generated during session generation.
In order to generate the session first time (or regenerate), use 
```go
go run getsession/cmd/server/main.go
```
