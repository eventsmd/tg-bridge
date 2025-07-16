# tg-bridge

It is a bridge for telegram internal API https://core.telegram.org/ (there is no way to get messages from telegram via bot api)

Communal organizations publish their events on telegram channels, messangers are source of truth for this events.

Communal outages messages could be parsed from these channels for the first time:

* @vodokanalpmrcom
* @eresofficial

The service gets messages from configured channels with configured interval and saves in external db as state last message id from parsed channel.

There could be configured more than one channel because of restrictions of telegram api.

After parsing it saves original message with metadata to db and push it to queue for further processing.
