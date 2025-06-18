import { createContext, useContext, useCallback, useEffect } from 'react'
import { ReadyState, useWebSocket } from 'react-use-websocket'
import { queryClient } from './queryClient'

const MessagesContext = createContext({ sendMessage: () => {}, canSendMessages: false })
const QUERY_KEY = ['messages']

export function ChatMessagesProvider({ children }) {
  const { sendJsonMessage, lastJsonMessage, readyState } = useWebSocket('ws://localhost:3001')

  const canSendMessages = readyState === ReadyState.OPEN

  const sendMessage = useCallback(
    (content) => {
      sendJsonMessage({ type: 'SEND_MESSAGE', payload: content })
    },
    [sendJsonMessage],
  )

  useEffect(() => {
    if (!lastJsonMessage) return
    const { type, payload } = lastJsonMessage
    if (type === 'INITIAL_DATA') {
      queryClient.setQueryData(QUERY_KEY, payload)
    } else if (type === 'NEW_MESSAGE') {
      queryClient.setQueryData(QUERY_KEY, (old = []) => [...old, payload])
    }
  }, [lastJsonMessage])

  return (
    <MessagesContext.Provider value={{ sendMessage, canSendMessages }}>
      {children}
    </MessagesContext.Provider>
  )
}

export function useChatMessages() {
  return useContext(MessagesContext)
}
