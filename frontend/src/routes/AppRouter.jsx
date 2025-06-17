import { BrowserRouter, Routes, Route } from 'react-router-dom'
import RoomPage from '../pages/RoomPage'

export default function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<RoomPage />} />
      </Routes>
    </BrowserRouter>
  )
}
