import { useState, useEffect, useRef } from "react";
import RoomJoinForm from "../features/room/RoomJoinForm";
import { useRoomStore } from "../stores/roomStore";
import useWebSocket from "../hooks/useWebSocket";
import { WS_URL } from "../services/websocket";
import { API_URL } from "../services/api";
import YouTubePlayer from "../components/YouTubePlayer";

const TIME_LIMIT = parseInt(import.meta.env.VITE_TIME_LIMIT) || 10;

export default function RoomPage() {
  const addMessage = useRoomStore((state) => state.addMessage);
  const clearMessages = useRoomStore((state) => state.clearMessages);
  const setQuestionActive = useRoomStore((state) => state.setQuestionActive);
  const setWinner = useRoomStore((state) => state.setWinner);
  const questionActive = useRoomStore((state) => state.questionActive);
  const winner = useRoomStore((state) => state.winner);
  const messages = useRoomStore((state) => state.messages);
  const readyStates = useRoomStore((state) => state.readyStates);
  const setReadyStates = useRoomStore((state) => state.setReadyStates);
  const [joined, setJoined] = useState(false);
  const [name, setName] = useState("");
  const [roomId, setRoomId] = useState("");
  const [playlistId, setPlaylistId] = useState("");
  const [videoId, setVideoId] = useState("");
  const [timeLeft, setTimeLeft] = useState(0);
  const [playing, setPlaying] = useState(false);
  const [playerReady, setPlayerReady] = useState(false);
  const [pauseInfo, setPauseInfo] = useState("");
  const timerRef = useRef(null);
  const { connect, send } = useWebSocket(WS_URL);

  useEffect(() => {
    setPlayerReady(false);
  }, [videoId]);

  const handleJoin = (rid, userName, pid) => {
    clearMessages();
    setName(userName);
    setRoomId(rid);
    setPlaylistId(pid);
    setPlayerReady(false);
    connect(
      rid,
      (event) => {
        const data = JSON.parse(event.data);
        if (data.type === "start") {
          setQuestionActive(true);
          setWinner(null);
          setPauseInfo("");
          setPlaying(true);
          setTimeLeft(TIME_LIMIT);
        } else if (data.type === "buzz_result") {
          setWinner(data.user);
          setQuestionActive(false);
          setPlaying(false);
          clearInterval(timerRef.current);
          timerRef.current = null;
        } else if (data.type === "timeout") {
          setWinner(null);
          setQuestionActive(false);
          setPlaying(false);
          clearInterval(timerRef.current);
          timerRef.current = null;
        } else if (data.type === "answer") {
          setPlaying(false);
          setPauseInfo(`${data.user}さんが解答ボタンを押しました - 再生停止中`);
        } else if (data.type === "ready_state") {
          setReadyStates(data.readyUsers);
        }
        addMessage(event.data);
      },
      () => {
        send(JSON.stringify({ type: "join", user: userName }));
        fetch(`${API_URL}/api/youtube/random?playlistId=${encodeURIComponent(pid)}`)
          .then((res) => res.json())
          .then((data) => {
            if (data.videoId) {
              setVideoId(data.videoId);
            }
          });
      },
    );
    setReadyStates({});
    setJoined(true);
  };

  const sendBuzz = () => {
    send(JSON.stringify({ type: "buzz", user: name }));
    setPlaying(false);
    setPauseInfo(`${name}さんが解答ボタンを押しました - 再生停止中`);
  };

  const sendReady = () => {
    send(JSON.stringify({ type: "ready", user: name }));
  };

  // Fetch a new random video and start the next quiz round
  const nextQuestion = () => {
    fetch(`${API_URL}/api/youtube/random?playlistId=${encodeURIComponent(playlistId)}`)
      .then((res) => res.json())
      .then((data) => {
        if (data.videoId) {
          setVideoId(data.videoId);
          setPlayerReady(false);
          // Notify the server to start the question
          send(JSON.stringify({ type: "start" }));
        }
      });
  };

  useEffect(() => {
    if (questionActive) {
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
      timerRef.current = setInterval(() => {
        setTimeLeft((t) => (t > 0 ? t - 1 : 0));
      }, 1000);
    } else if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [questionActive]);

  return (
    <div>
      {joined ? (
        <div>
          {!readyStates[name] && !questionActive && (
            <button onClick={sendReady} disabled={!playerReady}>
              準備完了
            </button>
          )}
          {Object.entries(readyStates).map(([u, r]) => (
            <p key={u}>
              {u}さん：{r ? "準備完了" : "未準備"}
            </p>
          ))}
          {playing && <p>再生中…</p>}
          {pauseInfo && <p>{pauseInfo}</p>}
          <YouTubePlayer
            videoId={videoId}
            playing={playing}
            onPlayerReady={() => setPlayerReady(true)}
          />
          {questionActive && (
            <div>
              <p>制限時間: {timeLeft}秒</p>
              <button onClick={sendBuzz}>解答ボタン</button>
            </div>
          )}
          {winner && <p>{winner}さんが解答権を獲得しました</p>}
          {!questionActive && (
            <button onClick={nextQuestion}>次の問題</button>
          )}
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
