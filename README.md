# go-image-analyzer

Hugging Face Inference APIを活用したGoバックエンドの画像物体検出サービス。

画像から物体を検出し、その結果をPostgreSQLに永続化します。

## 技術スタック

| 種別 | 技術 |
|------|------|
| 言語 | Go 1.27.3（CI検証環境） |
| HTTPフレームワーク | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io) |
| データベース | PostgreSQL 16 |
| 物体検出モデル | [facebook/detr-resnet-50](https://huggingface.co/facebook/detr-resnet-50) |
| コンテナ | Docker / Docker Compose |
| ホットリロード | [Air](https://github.com/air-verse/air) |
| テスト | Go testing / sqlmock |
| 静的解析 | go vet / golangci-lint |
| セキュリティチェック | govulncheck |
| CI/CD | GitHub Actions |

## ディレクトリ構成

```text
go-image-analyzer/
├── .github/
│   └── workflows/
│       └── ci.yml                 # GitHub Actions CI
├── db/
│   ├── postgres.go                # DB接続（GORM）
│   └── init.sql                   # テーブル初期化SQL
├── handler/
│   ├── analyze.go                 # HTTPハンドラー
│   └── analyze_test.go            # Handlerのユニットテスト
├── model/
│   ├── response.go                # データモデル・GORM構造体
│   └── response_test.go           # Modelのユニットテスト
├── repository/
│   ├── analysis.go                # DB操作
│   ├── analysis_test.go           # Repositoryのユニットテスト
│   └── integration_test.go        # PostgreSQL統合テスト
├── service/
│   ├── huggingface.go             # Hugging Face APIクライアント
│   └── huggingface_test.go        # Serviceのユニットテスト
├── docker-compose.yml             # PostgreSQLコンテナ定義
├── main.go                        # エントリーポイント
├── go.mod                         # Goモジュール定義
├── go.sum                         # 依存パッケージ情報
├── .env                           # 環境変数（gitignore済み）
└── .air.toml                      # Airホットリロード設定
```

## セットアップ

### 1. APIキーの取得

[Hugging Face Settings](https://huggingface.co/settings/tokens) からトークンを取得します。

付与する権限は `Make calls to Inference Providers` のみです。

### 2. 環境変数の設定

プロジェクトのルートに `.env` ファイルを作成します。

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

画像を送信して物体検出を実行します。

検出結果はPostgreSQLに保存されます。

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

**レスポンス：200 OK**

```json
{
  "objects": [
    {
      "label": "cat",
      "score": 0.988,
      "box": {
        "xmin": 0,
        "ymin": 2,
        "xmax": 639,
        "ymax": 476
      }
    }
  ]
}
```

**レスポンス：422 Unprocessable Entity（物体未検出）**

```json
{
  "error": "no objects detected: the image could not be analyzed or contains no recognizable objects"
}
```

物体が検出されなかった場合はエラーを返却し、DBへの保存は行いません。

## テスト

Handler・Service・Repository・Modelを対象にユニットテストを実装しています。

また、PostgreSQLを使用したRepositoryの統合テストも実装しています。

### テスト構成

| 対象 | テスト内容 | テスト方法 |
|------|------------|------------|
| Handler | 正常系、リクエスト不正、外部APIエラー、DB保存エラーなど | モック |
| Service | API正常応答、HTTPエラー、JSON解析エラー、通信エラーなど | HTTPクライアントのモック |
| Repository | DB保存の正常系・異常系 | sqlmock |
| Repository | PostgreSQLへの保存・取得、JSONB、ID、作成日時の確認 | PostgreSQL統合テスト |
| Model | JSONBのValue / Scan処理 | ユニットテスト |

### テスト実行

PostgreSQLが起動していることを確認したうえで、統合テスト用DBを指定して実行します。

```bash
DB_NAME=image_analyzer_test go test -race -cover ./...
```

※ 統合テストでは `DB_NAME=image_analyzer_test` を使用します。DB接続情報やテスト用スキーマも実行環境に合わせて設定してください。

### 静的解析

```bash
go vet ./...
```

### ビルド確認

```bash
go build ./...
```

### フォーマット確認

```bash
gofmt -l .
```

出力がなければ、Goファイルのフォーマットに問題はありません。

### セキュリティチェック

[govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) を使用して、Goの依存パッケージやコードから到達可能な脆弱性を確認します。

```bash
govulncheck ./...
```

## GitHub Actions

GitHub Actionsを利用して、PushおよびPull Request時に自動チェックを実行します。

### 実行タイミング

- 任意のブランチへのPush
- `main` ブランチを対象としたPull Request

### CIチェック内容

| ジョブ | チェック内容 |
|--------|--------------|
| Test and Build | PostgreSQL起動、DB初期化、gofmt、テスト、go vet、ビルド |
| Lint | golangci-lintによる静的解析 |
| Security | govulncheckによる脆弱性チェック |

テスト・ビルドジョブではPostgreSQL 16をサービスコンテナとして起動し、統合テストを含めて実行します。

GoのバージョンはCIとローカル環境で統一しています。

## DB確認

PostgreSQLに保存された検出結果は、以下のコマンドで確認できます。

```bash
docker exec -it go-image-analyzer-db psql -U postgres -d image_analyzer \
  -c "SELECT id, filename, created_at, objects FROM analysis_results;"
```
