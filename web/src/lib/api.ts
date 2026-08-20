// API 层：拉取 /data，用 TanStack Query 缓存；提供 SSE 订阅触发失效
import type { Aggregate } from './types'

export async function fetchData(): Promise<Aggregate> {
  const r = await fetch('/data?t=' + Date.now())
  if (!r.ok) throw new Error('获取数据失败 ' + r.status)
  return r.json()
}

// 订阅 SSE /events，文档变化时触发 onRefresh
export function subscribeSSE(onRefresh: () => void): () => void {
  let es: EventSource | null = null
  try {
    es = new EventSource('/events')
    es.onmessage = (e) => {
      if (e.data === 'refresh') onRefresh()
    }
  } catch {
    // 无 SSE 时降级轮询
    const t = setInterval(onRefresh, 10000)
    return () => clearInterval(t)
  }
  return () => es?.close()
}
