import { useEffect, useRef } from "react";

export default function useWebSocket(url) {
  const socketRef = useRef(null);

  const connect = (roomId, onMessage, onOpen) => {
    const socket = new WebSocket(`${url}?roomId=${roomId}`);
    socketRef.current = socket;
    socket.onmessage = onMessage;
    if (onOpen) socket.onopen = onOpen;
  };

  useEffect(() => {
    return () => {
      socketRef.current?.close();
    };
  }, []);

  const send = (data) => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send(data);
    }
  };

  return { connect, send };
}
