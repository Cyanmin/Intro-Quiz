import { useQuery } from '@tanstack/react-query'

const fetchTitle = async () => {
  const res = await fetch('/api/youtube/test')
  if (!res.ok) throw new Error('Network response was not ok')
  return res.json()
}

export default function useYouTubeTitleQuery() {
  return useQuery({
    queryKey: ['youtube', 'firstTitle'],
    queryFn: fetchTitle,
    staleTime: 1000 * 60 * 5,
  })
}
