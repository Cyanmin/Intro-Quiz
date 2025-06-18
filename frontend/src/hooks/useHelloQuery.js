import { useQuery } from '@tanstack/react-query'

const fetchHello = async () => {
  const res = await fetch('/api/hello')
  if (!res.ok) throw new Error('Network response was not ok')
  return res.json()
}

export default function useHelloQuery() {
  return useQuery({
    queryKey: ['hello'],
    queryFn: fetchHello,
  })
}
