import { useState } from "react";

export default function RoomJoinForm({ onJoin }) {
  const [roomId, setRoomId] = useState("");
  const [name, setName] = useState("");
  const [playlistId, setPlaylistId] = useState("");

  const submit = (e) => {
    e.preventDefault();
    onJoin(roomId, name, playlistId);
  };

  return (
    <form onSubmit={submit}>
      <input
        placeholder="ルームIDを入力してください"
        value={roomId}
        onChange={(e) => setRoomId(e.target.value)}
      />
      <input
        placeholder="名前を入力してください"
        value={name}
        onChange={(e) => setName(e.target.value)}
      />
      <input
        placeholder="プレイリストIDを入力してください"
        value={playlistId}
        onChange={(e) => setPlaylistId(e.target.value)}
      />
      <button type="submit">参加</button>
    </form>
  );
}
