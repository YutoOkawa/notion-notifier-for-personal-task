# AGENTS.md

## プロジェクト概要
Notion タスクの締切通知を Discord に送信するアプリ。Go + DDD アーキテクチャ。
k3d (Kubernetes) クラスターへのデプロイに対応。

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
k8s/                        # Kubernetes マニフェスト (Kustomize)
  base/                     # 基本リソース (Deployment, ConfigMap)
  kustomization.yaml        # Kustomize 設定
scripts/                    # ユーティリティスクリプト
  deploy_k8s.sh             # デプロイスクリプト
  bump_version.sh           # バージョン更新スクリプト
Taskfile.yaml               # タスクランナー設定
VERSION                     # 現在のバージョン
```

## 開発コマンド
```bash
go test ./...                              # テスト
go build -o bin/notion-notifier cmd/server/main.go  # ローカルビルド
./bin/notion-notifier -config config.yaml  # ローカル実行
```

## デプロイメント (Kubernetes / k3d)
本プロジェクトは **SSH Side-loading** 方式を採用しており、開発機でビルドしたイメージを SSH 経由でリモートの k3d クラスターへ直接転送します。

### 前提条件
- `task` (go-task)
- `docker`
- `kubectl`
- リモートホストへの SSH 接続 (`ssh mac-mini.local`)
- リモートホスト上の `k3d` クラスター (`mac-mini-cluster`)

### コマンド
```bash
task deploy      # ビルド -> 転送 -> 適用 の一連フローを実行
task build       # Docker イメージのビルド
task transfer    # イメージを SSH 経由で転送・インポート
task apply       # Kubernetes マニフェストの適用
```

### バージョン管理
セマンティックバージョニング (`X.Y.Z`) を採用。

```bash
task bump:patch  # パッチバージョン更新 (例: 1.0.0 -> 1.0.1)
task bump:minor  # マイナーバージョン更新 (例: 1.0.0 -> 1.1.0)
task bump:major  # メジャーバージョン更新 (例: 1.0.0 -> 2.0.0)
```

## 環境変数 (Secrets)
`.env` ファイルの内容は Kustomize の `secretGenerator` により Kubernetes Secret としてデプロイされます。
- `NOTION_API_TOKEN`
- `NOTION_DATABASE_ID`
- `DISCORD_WEBHOOK_URL`
- `RUN_ON_STARTUP`

## 設計方針
- **DDD**: レイヤー分離を維持（domain は外部依存なし）
- **Configuration**: 環境変数は ConfigMap/Secret に分離
- **Deployment**:
  - `imagePullPolicy: IfNotPresent` を使用し、ローカルインポートされたイメージを優先
  - リソース制限 (Requests: 32Mi, Limits: 128Mi) により省メモリ運用

## Notion データ構造（Personal Tasks）
| プロパティ | 型 | 説明 |
|-----------|------|------|
| `Task name` | title | タスク名 |
| `Due` | date | 締切日（YYYY-MM-DD または RFC3339） |
| `Status` | status | `Not Started` / `In Progress` / `Done` / `Archived` |

※ `Not Started` または `In Progress` のタスクのみ通知対象
