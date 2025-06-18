import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClientProvider } from '@tanstack/react-query'
import App from './App'
import { queryClient } from './queryClient'
import { ChatMessagesProvider } from './ChatMessagesProvider'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <ChatMessagesProvider>
        <App />
      </ChatMessagesProvider>
    </QueryClientProvider>
  </React.StrictMode>
)
