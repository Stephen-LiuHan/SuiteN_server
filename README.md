# SuiteN Server

SuiteN バックエンドAPIサーバー

## 技術スタック

- **言語**: Go 1.22
- **フレームワーク**: Echo v4
- **ORM**: sqlx
- **データベース**: PostgreSQL (Neon / Cloud SQL)
- **認証**: JWT (golang-jwt)
- **デプロイ**: Google Cloud Run

## ディレクトリ構造

```
suiten_server/
├── main.go                 # エントリーポイント
├── go.mod                  # Go modules
├── Dockerfile              # Cloud Run用マルチステージビルド
├── .env.example            # 環境変数サンプル
├── internal/
│   ├── db/
│   │   └── db.go           # DB接続ロジック
│   ├── middleware/
│   │   └── auth.go         # JWT認証ミドルウェア
│   ├── models/
│   │   ├── types.go        # 共通型定義
│   │   ├── user.go         # ユーザー関連モデル
│   │   ├── book.go         # 書籍モデル
│   │   ├── session.go      # セッションモデル
│   │   ├── capture.go      # キャプチャモデル
│   │   ├── ai_result.go    # AI結果モデル
│   │   ├── feedback.go     # フィードバックモデル
│   │   └── sync.go         # 同期リクエスト/レスポンス
│   └── handler/            # APIハンドラー（未実装）
└── migrations/             # DBマイグレーション
```

## セットアップ

### 1. 依存関係のインストール

```bash
go mod download
```

### 2. 環境変数の設定

```bash
cp .env.example .env
# .env を編集してDATABASE_URLなどを設定
```

### 3. マイグレーションの実行

```bash
# golang-migrate をインストール
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# マイグレーション実行
migrate -path ./migrations -database "$DATABASE_URL" up
```

### 4. ローカル実行

```bash
go run main.go
```

## Cloud Run へのデプロイ

### 1. Docker イメージのビルド

```bash
gcloud builds submit --tag gcr.io/PROJECT_ID/suiten-server
```

### 2. Cloud Run へデプロイ

```bash
gcloud run deploy suiten-server \
  --image gcr.io/PROJECT_ID/suiten-server \
  --platform managed \
  --region asia-northeast1 \
  --set-env-vars "DATABASE_URL=your-database-url" \
  --set-env-vars "JWT_SECRET=your-jwt-secret" \
  --allow-unauthenticated
```

## API エンドポイント

### 認証

| Method | Path | 説明 |
|--------|------|------|
| POST | /v1/auth/register | ユーザー登録 |
| POST | /v1/auth/login | ログイン |
| POST | /v1/auth/refresh | トークンリフレッシュ |
| POST | /v1/auth/logout | ログアウト |
| DELETE | /v1/auth/account | アカウント削除 |

### 同期

| Method | Path | 説明 |
|--------|------|------|
| POST | /v1/sync | データ同期 |
| GET | /v1/sync/status | 同期状態確認 |

### バックアップ

| Method | Path | 説明 |
|--------|------|------|
| POST | /v1/backup/export | データエクスポート |
| POST | /v1/backup/import | データインポート |

### リソース

| Method | Path | 説明 |
|--------|------|------|
| GET | /v1/books | 書籍一覧 |
| GET | /v1/books/:id | 書籍詳細 |
| GET | /v1/books/:book_id/sessions | セッション一覧 |
| GET | /v1/sessions/:session_id/captures | キャプチャ一覧 |
| GET | /v1/captures/:capture_id/ai-result | AI結果取得 |
| GET | /v1/profile | プロファイル取得 |

### アナリティクス

| Method | Path | 説明 |
|--------|------|------|
| GET | /v1/analytics/personal | 個人統計 |
| GET | /v1/analytics/global | 全体統計 |
| GET | /v1/analytics/trends | トレンド |

## 環境変数

| 変数名 | 必須 | 説明 |
|--------|------|------|
| DATABASE_URL | Yes | PostgreSQL接続URL |
| PORT | No | サーバーポート（デフォルト: 8080） |
| JWT_SECRET | Yes | JWT署名用シークレット |

## 開発

### ビルド

```bash
go build -o server ./main.go
```

### テスト

```bash
go test ./...
```
