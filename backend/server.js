const http = require('http')
const WebSocket = require('ws')

const server = http.createServer()
const wss = new WebSocket.Server({ server })
const messages = []

wss.on('connection', (ws) => {
  ws.send(JSON.stringify({ type: 'INITIAL_DATA', payload: messages }))

  ws.on('message', (data) => {
    let msg
    try {
      msg = JSON.parse(data)
    } catch (_) {
      return
    }
    if (msg.type === 'SEND_MESSAGE') {
      messages.push(msg.payload)
      const out = JSON.stringify({ type: 'NEW_MESSAGE', payload: msg.payload })
      wss.clients.forEach((client) => {
        if (client.readyState === WebSocket.OPEN) {
          client.send(out)
        }
      })
    }
  })
})

server.listen(3001, () => {
  console.log('WebSocket server running on ws://localhost:3001')
})
