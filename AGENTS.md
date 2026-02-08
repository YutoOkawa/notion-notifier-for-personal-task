# AGENTS.md

## プロジェクト概要
Notion タスクの締切通知を Discord に送信するアプリ。Go + DDD アーキテクチャ。

## ディレクトリ構造
```
cmd/server/main.go          # エントリポイント
internal/
  domain/task/              # Task エンティティ、Repository インターフェース
  domain/notification/      # Notifier インターフェース
  application/              # NotificationService（ユースケース）
  infrastructure/notion/    # Notion API クライアント
  infrastructure/discord/   # Discord Webhook クライアント
  scheduler/                # cron スケジューラー
  config/                   # 設定読み込み
```

## 開発コマンド
```bash
go test ./...                              # テスト
go build -o bin/notion-notifier cmd/server/main.go  # ビルド
./bin/notion-notifier -config config.yaml  # ローカル実行
docker compose up -d                       # Docker 実行
```

## 環境変数
- `NOTION_API_TOKEN` - Notion API トークン
- `NOTION_DATABASE_ID` - 対象データベース ID
- `DISCORD_WEBHOOK_URL` - Discord Webhook URL
- `RUN_ON_STARTUP` - 起動時に即実行（true/false）

## 設計方針
- DDD レイヤー分離を維持（domain は外部依存なし）
- インターフェースで依存性逆転
- public メソッドに単体テスト

## Notion データ構造（Personal Tasks）
| プロパティ | 型 | 説明 |
|-----------|------|------|
| `Task name` | title | タスク名 |
| `Due` | date | 締切日（YYYY-MM-DD または RFC3339） |
| `Status` | status | `Not Started` / `In Progress` / `Done` / `Archived` |

※ `Not Started` または `In Progress` のタスクのみ通知対象
