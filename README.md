# イントロクイズ

このリポジトリは、オンラインイントロクイズアプリケーションの最小スターターです。

## プロジェクト構成

- `features/` - Feature specific components such as room or quiz modules.
- `components/` - Reusable UI components like buttons and inputs.
- `pages/` - Components used as routing pages.
- `hooks/` - Reusable React hooks.
- `services/` - API and WebSocket communication utilities.
- `stores/` - Global state management stores using Zustand.
- `utils/` - Generic utility functions.
- `routes/` - React Router configuration.
- `backend/` - GinとGorilla WebSocketで実装されたGoサーバー。

### backend ディレクトリ構成

- `cmd/intro-quiz/` - サーバーのエントリポイント。
- `internal/handler/` - HTTP や WebSocket のハンドラー。
- `internal/service/` - ビジネスロジック。
- `internal/model/` - ドメインモデルの定義。
- `pkg/ws/` - WebSocket 接続の管理ヘルパー。
- `config/` - 設定読み込み用（将来利用）。

各ディレクトリには、コンテナイメージをビルドするためのDockerfileが含まれています。

## 実行方法

### 必要なツール

- Node.js 18 以上
- Go 1.20 以上

### フロントエンド

```bash
cd frontend
npm install
npm run dev
```

ブラウザで `http://localhost:5173` を開きます。

### バックエンド

```bash
cd backend
go run ./cmd/intro-quiz
```

サーバーは `http://localhost:8080` で待ち受けます。

バックエンドを起動する前に `backend/.env.example` を `backend/.env` にコピーし、
`YOUTUBE_API_KEY` や `YOUTUBE_PLAYLIST_ID` などの値を適切に設定してください。
`docker-compose.yml` もこのファイルを利用します。

### Docker での実行

それぞれのディレクトリで以下を実行してイメージをビルドし、コンテナを起動できます。

```bash
docker build -t intro-quiz-frontend ./frontend
docker run -p 80:80 intro-quiz-frontend

docker build -t intro-quiz-backend ./backend
docker run -p 8080:8080 intro-quiz-backend
```

バックエンドイメージは静的リンクされたバイナリを生成するため、追加のライブラリを必要としません。
これは GLIBC のバージョン違いによる起動失敗を防ぐためです。

### docker-compose を使った起動

`docker-compose.yml` が用意されているため、ルートディレクトリで次のコマンドを実行するだけでフロントエンドとバックエンドの両方を起動できます。

```bash
docker-compose up --build
```
