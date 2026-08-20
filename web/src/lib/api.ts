// API 层：拉取 /data，用 TanStack Query 缓存；提供 SSE 订阅触发失效
import type { Aggregate, Document, Project } from './types'

export async function fetchData(): Promise<Aggregate> {
  const r = await fetch('/data?t=' + Date.now())
  if (!r.ok) throw new Error('获取数据失败 ' + r.status)
  const raw = (await r.json()) as Aggregate
  return normalize(raw)
}

// 读取项目内文件（文档里文件路径链接闭环）。
export interface FileContent {
  project: string
  path: string
  abs: string
  content: string
}

export async function fetchFile(project: string, path: string, dir?: string): Promise<FileContent> {
  const q = new URLSearchParams({ project, path })
  if (dir) q.set('dir', dir)
  const r = await fetch(`/api/file?${q.toString()}`)
  const data = await r.json().catch(() => ({}))
  if (!r.ok) throw new Error((data as { error?: string }).error || ('读取失败 ' + r.status))
  return data as FileContent
}

// 防御性归一化：确保数组/对象字段非 null（后端曾因 nil slice 序列化为 null 导致前端崩）
// 即使后端漏修，前端也不会因数据形状白屏。
function normalize(agg: Aggregate): Aggregate {
  for (const p of agg.projects as (Omit<Project, 'documents' | 'roadmap_stages'> & {
    documents?: Document[] | null
    roadmap_stages?: Project['roadmap_stages'] | null
    stats?: Project['stats'] | null
  })[]) {
    p.documents = p.documents ?? []
    p.roadmap_stages = p.roadmap_stages ?? []
    p.stats = (p.stats ?? {}) as Project['stats']
    for (const d of p.documents as (Omit<Document, 'tasks' | 'sections' | 'meta'> & {
      tasks?: Document['tasks'] | null
      sections?: Document['sections'] | null
      meta?: Document['meta'] | null
    })[]) {
      d.tasks = d.tasks ?? []
      d.sections = d.sections ?? []
      d.meta = d.meta ?? {}
    }
  }
  return agg
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
