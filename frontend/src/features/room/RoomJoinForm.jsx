import { useState } from 'react'

export default function RoomJoinForm({ onJoin }) {
  const [name, setName] = useState('')

  const submit = (e) => {
    e.preventDefault()
    onJoin(name)
  }

  return (
    <form onSubmit={submit}>
      <input
        placeholder="Enter name"
        value={name}
        onChange={(e) => setName(e.target.value)}
      />
      <button type="submit">Join</button>
    </form>
  )
}
