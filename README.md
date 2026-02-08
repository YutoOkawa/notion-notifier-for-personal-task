# Notion Due Date Notifier

Notion の Personal Tasks データベースから締切が近いタスクを取得し、Discord 経由で通知するアプリケーションです。

## 機能

- 締切日の N 日前からタスクを通知
- 毎日正午に自動チェック
- Discord Webhook による通知

## セットアップ

### 1. 環境変数の設定

```bash
export NOTION_API_TOKEN="your_notion_integration_token"
export NOTION_DATABASE_ID="your_notion_database_id"
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
```

### 2. 設定ファイルの編集

`config.yaml` を編集して設定をカスタマイズ:

```yaml
notification:
  days_before: 3              # 締切の何日前から通知するか
  check_schedule: "0 12 * * *"  # cron形式 (毎日12時)
```

### 3. 実行

```bash
go run cmd/server/main.go -config config.yaml
```

### 4. Docker で実行

#### docker-compose（推奨）

```bash
# ビルド
docker compose build

# 実行
docker compose up -d

# ログ確認
docker compose logs -f
```

#### docker run

```bash
# ビルド
docker build -t notion-notifier .

# 実行（config.yaml をボリュームマウント）
docker run \
  -v $(pwd)/config.yaml:/etc/config/notion-notifier/config.yaml:ro \
  -e NOTION_API_TOKEN=xxx \
  -e NOTION_DATABASE_ID=xxx \
  -e DISCORD_WEBHOOK_URL=xxx \
  notion-notifier
```

## 開発

```bash
# テスト実行
go test ./...

# ビルド
go build -o bin/notion-notifier cmd/server/main.go

# ローカル実行
./bin/notion-notifier -config config.yaml
```

## ライセンス

MIT
