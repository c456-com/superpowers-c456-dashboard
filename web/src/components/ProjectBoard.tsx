import { FolderOpen } from 'lucide-react'
import type { Project } from '../lib/types'

interface Props {
  projects: Project[]
  onEnter: (name: string) => void
}

export default function ProjectBoard({ projects, onEnter }: Props) {
  return (
    <main className="p-6">
      <div className="flex items-center gap-2 mb-4">
        <h2 className="text-xl font-bold">📁 全部项目</h2>
        <span className="text-sm text-muted-foreground">点击进入单个项目 · 共 {projects.length} 个</span>
      </div>

      {projects.length === 0 && (
        <div className="text-center text-muted-foreground p-8">
          <FolderOpen className="size-8 mx-auto mb-2 opacity-40" />
          未配置项目，请在 projects.yaml 里添加
        </div>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {projects.map((p) => {
          const s = p.stats
          return (
            <button
              key={p.name}
              onClick={() => onEnter(p.name)}
              className="text-left bg-card border border-border rounded-xl p-4 hover:border-primary hover:shadow-md transition-all cursor-pointer"
            >
              <div className="flex items-center gap-2 mb-2">
                <span className="font-bold">{p.name}</span>
                <span className="text-xs text-muted-foreground">{s.total_docs} 文档</span>
              </div>
              <div className="text-[11px] text-muted-foreground mb-2 break-all">{p.root}</div>
              <div className="flex items-center gap-2 mb-2">
                <div className="h-1.5 flex-1 bg-muted rounded-full overflow-hidden">
                  <div className="h-full bg-primary rounded-full" style={{ width: `${s.completion || 0}%` }} />
                </div>
                <span className="text-xs font-semibold text-primary whitespace-nowrap">{s.completion || 0}%</span>
              </div>
              <div className="text-xs text-muted-foreground mb-2">
                任务 {s.tasks_done}/{s.tasks_total} · spec {s.specs_total} · plan {s.plans_total} · sprint {s.sprints_total}
              </div>
              <div className="flex gap-1.5 flex-wrap">
                {p.type && <span className="px-2 py-0.5 rounded-full bg-violet-500/10 text-violet-700 dark:text-violet-400 text-xs">{p.type}</span>}
                {p.status && (
                  <span className={`px-2 py-0.5 rounded-full text-xs ${p.status.includes('暂停') ? 'bg-amber-500/10 text-amber-700 dark:text-amber-400' : 'bg-green-500/10 text-green-700 dark:text-green-400'}`}>
                    {p.status}
                  </span>
                )}
              </div>
            </button>
          )
        })}
      </div>
    </main>
  )
}
