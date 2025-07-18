# tg-bridge

It is a bridge for telegram internal API https://core.telegram.org/ (there is no way to get messages from telegram via bot api)

Communal organizations publish their events on telegram channels, messengers are source of truth for these events.

Communal outages messages could be parsed from these channels for the first time:

* @vodokanalpmrcom
* @eresofficial

The service gets messages from configured channels with configured interval and saves in external db as state last message id from parsed channel.

There could be configured more than one channel because of restrictions of telegram api.

After parsing, it saves original message with metadata to db and push it to queue for further processing.


# How to generate Telegram session

ℹ️ _Skip this step if session is already generated/provided._

⚠️ Be cautious and don't generate session too often to not get your service Telegram blocked.

Application is using cached Telegram session which is provided from environment variables.
Session is base64 encoded JSON which is generated during session generation.
In order to generate the session first time (or regenerate), use
```go
go run getsession/cmd/server/main.go
```

Find generated session in `generated-session` top-level folder.
Leave only fields mentioned below and remove the rest, then encode it in Base64 and provide in
`TELEGRAM_SESSION` environment variable in `.env` file.
```json
{
  "Version": 1,
  "Data": {
    "DC": DATA_CENTER_NUMBER_HERE,
    "Addr": "",
    "AuthKey": "AUTH_KEY_HERE",
    "AuthKeyID": "AUTH_KEY_ID_HERE",
    "Salt": SALT_HERE
  }
}
```


# How to run the application
Create `.env` file in repository root if missing.
Add required environment variables:
- `TELEGRAM_API_ID`=YOUR_API_ID_HERE
- `TELEGRAM_API_HASH`="YOUR_API_HASH_HERE"
- `PHONE`="YOUR_ACCOUNT_PHONE_NUMBER_IN_E.164_INTERNATIONAL_FORMAT"
- `TELEGRAM_SESSION`="YOUR_SESSION_JSON_ENCODED_INTO_BASE64_FORMAT"


```go
go run tgbridge/cmd/server/main.go
```
