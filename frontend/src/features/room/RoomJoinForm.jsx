import { useState } from "react";

export default function RoomJoinForm({ onJoin }) {
  const [roomId, setRoomId] = useState("");

  const submit = (e) => {
    e.preventDefault();
    onJoin(roomId);
  };

  return (
    <form onSubmit={submit}>
      <input
        placeholder="ルームIDを入力してください"
        value={roomId}
        onChange={(e) => setRoomId(e.target.value)}
      />
      <button type="submit">参加</button>
    </form>
  );
}
