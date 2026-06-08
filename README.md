# go-image-analyzer

Hugging Face Inference APIを活用したGoバックエンドの画像物体検出サービス。
検出結果はPostgreSQLに永続化。

## 技術スタック

| 種別 | 技術 |
|------|------|
| 言語 | Go 1.22+ |
| HTTPフレームワーク | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io) |
| データベース | PostgreSQL 16 |
| 物体検出モデル | [facebook/detr-resnet-50](https://huggingface.co/facebook/detr-resnet-50) |
| コンテナ | Docker / Docker Compose |
| ホットリロード | [Air](https://github.com/air-verse/air) |

## ディレクトリ構成

```
go-image-analyzer/
├── db/
│   ├── postgres.go          # DB接続（GORM）
│   └── init.sql             # テーブル初期化SQL
├── handler/
│   └── analyze.go           # HTTPハンドラー
├── model/
│   └── response.go          # データモデル・GORM構造体
├── repository/
│   └── analysis.go          # DB操作
├── service/
│   └── huggingface.go       # Hugging Face APIクライアント
├── docker-compose.yml       # PostgreSQLコンテナ定義
├── main.go                  # エントリーポイント
├── .env                     # 環境変数（gitignore済み）
└── .air.toml                # Airホットリロード設定
```

## セットアップ

### 1. APIキーの取得

[Hugging Face Settings](https://huggingface.co/settings/tokens) からトークンを取得（無料・カード登録不要）。
付与する権限は `Make calls to Inference Providers` のみ。

### 2. 環境変数の設定

`.env` ファイルを作成し、APIキーを設定。

```env
HF_API_KEY=hf_xxxxxxxxxxxxxxxxxx
PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=image_analyzer
```

### 3. PostgreSQLの起動

```bash
docker compose up -d
```

### 4. 依存パッケージのインストール

```bash
go mod tidy
```

### 5. サーバー起動

通常起動：
```bash
go run main.go
```

ホットリロード（Air）：
```bash
air
```

## API

### POST /api/analyze

画像を送信して物体検出を実行。検出結果はPostgreSQLに保存。

**リクエスト（multipart/form-data）**

```bash
curl -X POST http://localhost:8080/api/analyze \
  -F "image=@/path/to/image.jpg"
```

**リクエスト（raw bytes）**

```bash
curl -X POST http://localhost:8080/api/analyze \
  -H "Content-Type: image/jpeg" \
  --data-binary @/path/to/image.jpg
```

**レスポンス 200 OK**

```json
{
  "objects": [
    {
      "label": "cat",
      "score": 0.988,
      "box": { "xmin": 0, "ymin": 2, "xmax": 639, "ymax": 476 }
    }
  ]
}
```

**レスポンス 422 Unprocessable Entity**（物体未検出）

```json
{
  "error": "no objects detected: the image could not be analyzed or contains no recognizable objects"
}
```

> 物体未検出の場合はエラーを返却。DBへの保存は非実施。

## DB確認

```bash
docker exec -it go-image-analyzer-db psql -U postgres -d image_analyzer \
  -c "SELECT id, filename, created_at, objects FROM analysis_results;"
```