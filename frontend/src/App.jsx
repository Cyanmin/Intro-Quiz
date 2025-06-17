import { useEffect, useRef, useState } from 'react'

// Minimal component that connects to WebSocket and displays messages
export default function App() {
  const [messages, setMessages] = useState([])
  const socketRef = useRef(null)

  // Establish WebSocket connection on mount
  useEffect(() => {
    socketRef.current = new WebSocket('ws://localhost:8080/ws')
    socketRef.current.onmessage = (event) => {
      // Append incoming message to the list
      setMessages((prev) => [...prev, event.data])
    }
    return () => {
      socketRef.current?.close()
    }
  }, [])

  // Send a test message
  const sendTest = () => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send('test')
    }
  }

  return (
    <div>
      <button onClick={sendTest}>Send Test</button>
      <ul>
        {messages.map((msg, i) => (
          <li key={i}>{msg}</li>
        ))}
      </ul>
    </div>
  )
}
