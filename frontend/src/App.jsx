import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useChatMessages } from './ChatMessagesProvider'

const QUERY_KEY = ['messages']

function fetchInitialMessages() {
  return Promise.resolve([])
}

export default function App() {
  const [text, setText] = useState('')
  const { sendMessage, canSendMessages } = useChatMessages()

  const { data: messages = [] } = useQuery({
    queryKey: QUERY_KEY,
    queryFn: fetchInitialMessages,
    staleTime: Infinity,
  })

  const mutation = useMutation({
    mutationFn: async (content) => {
      sendMessage(content)
    },
  })

  const submit = (e) => {
    e.preventDefault()
    if (!text) return
    mutation.mutate(text)
    setText('')
  }

  return (
    <div>
      <ul>
        {messages.map((m, i) => (
          <li key={i}>{m}</li>
        ))}
      </ul>
      <form onSubmit={submit}>
        <input value={text} onChange={(e) => setText(e.target.value)} />
        <button type="submit" disabled={!canSendMessages}>
          Send
        </button>
      </form>
    </div>
  )
}
