# イントロクイズ

このリポジトリは、オンラインイントロクイズアプリケーションの最小スターターです。

## プロジェクト構成

- `frontend/` - Viteで構築されたReactアプリ。
- `backend/` - Gorillaを使用したGoのWebSocketサーバー。

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
go run main.go
```

サーバーは `http://localhost:8080` で待ち受けます。

### Docker での実行

それぞれのディレクトリで以下を実行してイメージをビルドし、コンテナを起動できます。

```bash
docker build -t intro-quiz-frontend ./frontend
docker run -p 80:80 intro-quiz-frontend

docker build -t intro-quiz-backend ./backend
docker run -p 8080:8080 intro-quiz-backend
```

### docker-compose を使った起動

`docker-compose.yml` が用意されているため、ルートディレクトリで次のコマンドを実行するだけでフロントエンドとバックエンドの両方を起動できます。

```bash
docker-compose up --build
```

