# gitlab-telegram-notification

<img width="654" alt="Screenshot 2023-12-03 at 02 18 14" src="https://github.com/herpiko/gitlab-telegram-notification/assets/2534060/c30ebfc6-d19b-417a-9ca9-4a8225e59630">

## Usage

1. Create `.env` file with this content

```
TELEGRAM_BOT_TOKEN=foobar
TELEGRAM_CHAT_ID=foobar
```

The last working bot I know that can generate group chat ID is `@SimpleID_Bot`.

2. Run

`go run main.go`

3. Setup the weebhook in Gitlab

Target it to `http://yourhost:8080`. SSL need to be turned off unless you setup the SSL for this service.
