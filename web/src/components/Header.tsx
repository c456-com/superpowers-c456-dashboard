import { Activity } from 'lucide-react'

interface Props {
  total: number
  globalDone: number
  globalTotal: number
  globalCompletion: number
  generatedAt: string
}

export default function Header({ total, globalDone, globalTotal, globalCompletion, generatedAt }: Props) {
  return (
    <header className="sticky top-0 z-50 bg-white border-b border-border px-6 py-3 flex items-center gap-4 flex-wrap">
      <div className="font-bold text-base tracking-wide flex items-center gap-2">
        <Activity className="size-4 text-primary" />
        多项目监控 · {total} 个项目
      </div>
      <span className="inline-flex items-center rounded-full bg-blue-50 text-primary text-xs font-semibold px-3 py-1">
        全局任务 {globalDone}/{globalTotal} · {globalCompletion}%
      </span>
      <div className="flex-1" />
      <div className="text-xs text-muted-foreground">更新于 {generatedAt}</div>
      <div className="text-xs text-muted-foreground flex items-center gap-1">
        <span className="inline-block size-2 rounded-full bg-green-500" />
        自动刷新
      </div>
    </header>
  )
}
