# 詳細設計書

## 概要
本プロジェクトはオンラインイントロクイズを実現するための最小構成サンプルです。フロントエンドはReact + Vite、バックエンドはGo (Gin) で実装されており、WebSocketを用いてリアルタイム通信を行います。

## ディレクトリ構成
READMEに記載の通り、主要ディレクトリは以下の通りです。
- `features/` などReactコンポーネント群
- `backend/` Goで実装されたサーバー
- `hooks/` `services/` `stores/` などのユーティリティ類

バックエンドの詳細は `cmd/intro-quiz/`、`internal/handler/`、`internal/service/` などに分割されています【F:README.md†L5-L24】。

## バックエンド設計
### エンドポイント
`main.go` でHTTPルーティングを設定しています。
- `/ws` WebSocketエンドポイント
- `/api/hello` テスト用の簡易API
- `/api/youtube/test` 固定プレイリストの先頭動画タイトル取得
- `/api/youtube/embeddable/:videoId` 動画埋め込み可能判定
これらは `cmd/intro-quiz/main.go` で定義されています【F:backend/cmd/intro-quiz/main.go†L19-L28】。

### WebSocket処理
`WSHandler` で接続をアップグレードし `RoomService` を通じてメッセージを処理します【F:backend/internal/handler/ws.go†L20-L43】。`RoomManager` はルームごとの状態管理を行い、ユーザー登録・準備状態・解答権などを管理します。

`RoomService.ProcessMessage` では `join` `playlist` `ready` `start` `buzz` `answer_text` といった種類のメッセージを解析し、各種ブロードキャストや状態遷移を実施します【F:backend/internal/service/room.go†L374-L442】。

YouTube の動画取得や埋め込み可否判定は `YouTubeService` にまとめられています【F:backend/internal/service/youtube.go†L1-L165】。APIキーは環境変数 `YOUTUBE_API_KEY` から読み込みます【F:backend/internal/service/youtube.go†L146-L148】。

環境変数の読み込みは `LoadEnv` で行われ、`TIME_LIMIT` を設定可能です【F:backend/internal/config/env.go†L11-L23】。

## フロントエンド設計
React で構成され、`RoomPage.jsx` が主要画面となります。ここではWebSocket接続、YouTube再生、解答権管理等を行います。制限時間は `.env` から `VITE_TIME_LIMIT` として取得します【F:frontend/src/pages/RoomPage.jsx†L1-L9】。

WebSocket通信は `useWebSocket` フックで抽象化されています【F:frontend/src/hooks/useWebSocket.js†L1-L25】。ユーザーの準備状態や押した順序等は Zustand ストア `roomStore.js` で管理します【F:frontend/src/stores/roomStore.js†L1-L15】。

`RoomPage.jsx` 内では `join` `buzz` `playlist` `ready` `answer_text` などのメッセージを送信し、サーバーからの `start` `buzz_result` `ready_state` `video` などを受け取って画面を更新します【F:frontend/src/pages/RoomPage.jsx†L70-L110】。

YouTube の再生は `YouTubePlayer` コンポーネントで行いますが、プレイヤーは非表示で音声のみ再生する設計です【F:frontend/src/components/YouTubePlayer.jsx†L1-L31】。

## 動作シーケンス概要
1. 参加者は `RoomJoinForm` からルームIDと名前を入力し `join` メッセージを送信。
2. ルーム内で全員が `ready` を送ると、サーバー側でタイマーが開始され `start` が送信されます。
3. YouTube 音源再生中にユーザーが `buzz` を送ると、最速ユーザーが決定され `buzz_result` がブロードキャストされます。
4. 解答者は `answer_text` を送信し、正誤判定結果 `answer_result` が共有されます。正解後またはタイムアウト後に次の動画へ進みます。

## 環境設定
`.env.example` を各ディレクトリで `.env` にコピーして利用します。`VITE_TIME_LIMIT` と `TIME_LIMIT` を同一値にするとクイズの制限時間を変更できます【F:README.md†L54-L59】。

## 依存関係・実行方法
必要なツールは Node.js 18 以上と Go 1.20 以上です【F:README.md†L30-L33】。詳細な起動手順は README に記載されています【F:README.md†L35-L82】。
