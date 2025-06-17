import { useState, useEffect, useRef } from "react";
import RoomJoinForm from "../features/room/RoomJoinForm";
import { useRoomStore } from "../stores/roomStore";
import useWebSocket from "../hooks/useWebSocket";
import { WS_URL } from "../services/websocket";

export default function RoomPage() {
  const addMessage = useRoomStore((state) => state.addMessage);
  const clearMessages = useRoomStore((state) => state.clearMessages);
  const setQuestionActive = useRoomStore((state) => state.setQuestionActive);
  const setWinner = useRoomStore((state) => state.setWinner);
  const questionActive = useRoomStore((state) => state.questionActive);
  const winner = useRoomStore((state) => state.winner);
  const messages = useRoomStore((state) => state.messages);
  const [joined, setJoined] = useState(false);
  const [name, setName] = useState("");
  const [roomId, setRoomId] = useState("");
  const [timeLeft, setTimeLeft] = useState(0);
  const timerRef = useRef(null);
  const { connect, send } = useWebSocket(WS_URL);

  const handleJoin = (rid, userName) => {
    clearMessages();
    setName(userName);
    setRoomId(rid);
    connect(rid, (event) => {
      const data = JSON.parse(event.data);
      if (data.type === "start") {
        setQuestionActive(true);
        setWinner(null);
        setTimeLeft(10);
        if (timerRef.current) clearInterval(timerRef.current);
        timerRef.current = setInterval(() => {
          setTimeLeft((t) => (t > 0 ? t - 1 : 0));
        }, 1000);
      } else if (data.type === "buzz_result") {
        setWinner(data.user);
        setQuestionActive(false);
        clearInterval(timerRef.current);
      } else if (data.type === "timeout") {
        setWinner(null);
        setQuestionActive(false);
        clearInterval(timerRef.current);
      }
      addMessage(event.data);
    });
    setJoined(true);
  };

  const sendStart = () => {
    send(JSON.stringify({ type: "start" }));
  };

  const sendBuzz = () => {
    send(JSON.stringify({ type: "buzz", user: name }));
  };

  useEffect(() => {
    if (!questionActive && timerRef.current) {
      clearInterval(timerRef.current);
    }
  }, [questionActive]);

  return (
    <div>
      {joined ? (
        <div>
          <button onClick={sendStart}>問題開始</button>
          {questionActive && (
            <div>
              <p>制限時間: {timeLeft}秒</p>
              <button onClick={sendBuzz}>解答ボタン</button>
            </div>
          )}
          {winner && <p>{winner}さんが解答権を獲得しました</p>}
          <ul>
            {messages.map((msg, i) => (
              <li key={i}>{msg}</li>
            ))}
          </ul>
        </div>
      ) : (
        <RoomJoinForm onJoin={handleJoin} />
      )}
    </div>
  );
}
