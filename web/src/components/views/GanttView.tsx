import type { Project } from '../../lib/types'

interface Props {
  project: Project
}

// 简单甘特：按日期排序，渲染横向时间条
export default function GanttView({ project }: Props) {
  const plans = ['plan', 'sprint']
    .map((t) => project.documents.filter((d) => d.type === t))
    .flat()
  const today = new Date().toISOString().slice(0, 10)

  if (plans.length === 0) {
    return (
      <div className="text-center text-muted-foreground p-8 bg-card border border-border rounded-xl">
        暂无开发计划文档
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center gap-2 mb-4">
        <h2 className="text-xl font-bold">📅 开发计划</h2>
        <span className="text-sm text-muted-foreground">plan + sprint 时间安排</span>
      </div>
      <div className="bg-card border border-border rounded-xl p-5">
        {plans.map((d) => {
          const done = (d.tasks || []).filter((t) => t.done).length
          const total = (d.tasks || []).length
          const pct = total ? Math.round((done / total) * 100) : 0
          const start = d.date || today
          return (
            <div key={d.path} className="flex items-center gap-3 py-2.5 border-b border-border last:border-0">
              <div className="w-40 shrink-0 truncate text-[13px] font-medium">{d.title.slice(0, 20)}</div>
              <div className="text-[11px] text-muted-foreground shrink-0 w-24">{start}</div>
              <div className="flex-1 h-5 bg-muted rounded relative overflow-hidden">
                <div className="h-full bg-blue-500/30 rounded" style={{ width: `${pct}%` }} />
                <div
                  className="absolute inset-y-0 left-0 bg-primary/80 rounded"
                  style={{ width: `${Math.min(100, pct || 8)}%` }}
                />
              </div>
              <span className="inline-flex items-center rounded-full bg-primary/10 text-primary text-xs font-semibold px-2 py-0.5 shrink-0">
                {done}/{total} {pct}%
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
