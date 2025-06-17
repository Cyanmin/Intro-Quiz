import { useEffect, useRef, useState } from "react";
import { WS_URL } from "../services/websocket";

export default function WebSocketClient({ roomId, user }) {
  const wsRef = useRef(null);
  const [messages, setMessages] = useState([]);

  useEffect(() => {
    if (!roomId || !user) return;
    const ws = new WebSocket(`${WS_URL}?roomId=${roomId}`);
    wsRef.current = ws;

    ws.onopen = () => {
      ws.send(JSON.stringify({ type: "join", user }));
    };

    ws.onmessage = (event) => {
      setMessages((prev) => [...prev, event.data]);
    };

    return () => {
      if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
        ws.close(1000, "normal closure");
      }
    };
  }, [roomId, user]);

  const sendReady = () => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: "ready", user }));
    }
  };

  return (
    <div>
      <button onClick={sendReady}>Send Ready</button>
      <ul>
        {messages.map((msg, i) => (
          <li key={i}>{msg}</li>
        ))}
      </ul>
    </div>
  );
}
