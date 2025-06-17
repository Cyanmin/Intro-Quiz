import { useState } from 'react'
import RoomJoinForm from '../features/room/RoomJoinForm'
import { useRoomStore } from '../stores/roomStore'
import useWebSocket from '../hooks/useWebSocket'
import { WS_URL } from '../services/websocket'

export default function RoomPage() {
  const addMessage = useRoomStore((state) => state.addMessage)
  const messages = useRoomStore((state) => state.messages)
  const [joined, setJoined] = useState(false)
  const { connect, send } = useWebSocket(WS_URL)

  const handleJoin = () => {
    connect((event) => addMessage(event.data))
    setJoined(true)
  }

  const sendTest = () => {
    send('test')
  }

  return (
    <div>
      {joined ? (
        <div>
          <button onClick={sendTest}>Send Test</button>
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
  )
}
