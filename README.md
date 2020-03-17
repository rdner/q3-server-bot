# Quake 3 Server Bot

The purpose of this bot is to make it easier for Quake 3 players to know what's happening on the server.

## Features

This bot sends messages to the given group/channel on Telegram when:

* the Quake 3 server becomes available or unavailable
* players join or leave the Quake 3 server

There is a throttling interval of 3 seconds in which the messages are grouped together, so the bot does not spam the chat much.

Here is an example of the messages from the bot:

![bot messages](img/bot-messages.png)

## Compatibility

This bot and its parsers are tested with [Challenge ProMode Arena 1.52](https://www.playmorepromode.com/). There is absolutely no warranty it would work with other mods/versions.

## Usage

```
Usage of ./bin/q3serverbot:
  -debug
        Use this flag if you'd like to see detailed logging output
  -q3-password string
        The rcon password of the server. You can use 'Q3_PASSWORD' environment variable instead
  -q3-server-addr string
        Address of the quake server including port. You can use 'Q3_SERVER_ADDR' environment variable instead
  -telegram-chat-id string
        Unique identifier for the target chat or username of the target channel (in the format @channelusername). You can use 'TELEGRAM_CHAT_ID' environment variable instead.
  -telegram-token string
        Bot token that you get when you create a Telegram bot. You can use 'TELEGRAM_TOKEN' environment variable instead.
```

## Deployment

* `make` builds the binary in `./bin/q3serverbot`
* `make systemd` prompts some questions interactively and generates a systemd unit file (the service process will run from the current user and from `./bin/q3serverbot` path).
* `make clean` removes the binary and the systemd unit file.

MIT License.

Denis Rechkunov (mail@pragmader.me)
