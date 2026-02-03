# AhLong

A fully automatic expense tracker that detects DBS PayLah!/PayNow transactions from Gmail, prompts you to categorize them via Telegram, and logs them to a Notion database.

## How It Works

1. Gmail receives a DBS PayLah!/PayNow transaction email
2. Google Pub/Sub pushes a notification to the backend on Cloud Run
3. Backend fetches the email via Gmail API and extracts transaction details (amount, recipient)
4. Sends a Telegram message with category buttons (Food, Personal, Transportation)
5. You tap a button, and the expense is logged to your Notion database

## Architecture

```
┌─────────┐    ┌──────────┐    ┌──────────────────────────┐
│  Gmail  │───>│ Pub/Sub  │──> │   Cloud Run (Go Backend) │
│  (DBS   │    │          │    │                          │
│  email  │    │          │    │  /handlePubSub           │
└─────────┘    └──────────┘    │    │ Gmail History API   │
                               │    │ Gmail Messages API  │
                               │    │ Extract transaction │
                               │    v                     │
                               │  /telegramCallback  <──┐ │
                               │    │ Parse callback    │ │
                               │    │ Create Notion     │ │
                               │    │ Send confirmation │ │
                               └────┼───────────────────┼─┘
                                    │                   │
                           sends    │                   │ user taps
                           prompt   │                   │ button
                                    v                   │
                               ┌──────────────────┐     │
                               │   Telegram Bot   │─────┘
                               │                  │
                               │  [Food]          │
                               │  [Personal]      │
                               │  [Transportation]│
                               └────────┬─────────┘
                                        │
                                        │ logs expense
                                        v
                               ┌──────────────────┐
                               │    Notion DB     │
                               │                  │
                               │  Name | Amount   │
                               │  Category | Date │
                               └──────────────────┘
```

## Project Structure

```
main.go       — HTTP server, routing, health check
app.go        — App struct, NewApp() initialization
gmail.go      — Pub/Sub handler, Gmail history fetching, email processing
telegram.go   — Telegram message sending, callback handling, message deletion
notion.go     — Notion API payload creation and request handling
utils.go      — Shared structs and helper functions
```

## Setup

### Prerequisites

- Go 1.24+
- Google Cloud account with Gmail API and Pub/Sub enabled
- Telegram Bot (create via [@BotFather](https://t.me/BotFather))
- Notion integration with a database

### Environment Variables

Create a `.env` file in the project root:

```
export GMAIL_CLIENT_ID=your_client_id
export GMAIL_CLIENT_SECRET=your_client_secret
export GMAIL_REFRESH_TOKEN=your_refresh_token
export TELEGRAM_BOT_TOKEN=your_bot_token
export TELEGRAM_CHAT_ID=your_chat_id
export NOTION_API_KEY=your_notion_api_key
```

### Run Locally

```bash
source .env && go run .
```

### Deploy to Cloud Run

```bash
gcloud run deploy ahlong --source . --region asia-southeast1 --allow-unauthenticated
```

### Set Cloud Run Environment Variables

```bash
gcloud run services update ahlong --region asia-southeast1 --update-env-vars KEY=value,KEY2=value2
```

### View Logs

```bash
gcloud beta run services logs tail ahlong --region asia-southeast1
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/handlePubSub` | Receives Pub/Sub push notifications |
| POST | `/telegramCallback` | Receives Telegram inline keyboard callbacks |
