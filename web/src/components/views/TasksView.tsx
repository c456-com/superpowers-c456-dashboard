import type { Document, Project } from '../../lib/types'
import { TypeTag } from './OverviewView'

interface Props {
  project: Project
  onOpen: (d: Document) => void
}

export default function TasksView({ project, onOpen }: Props) {
  const docs = ['plan', 'sprint', 'spec']
    .map((t) => project.documents.filter((d) => d.type === t))
    .flat()
    .filter((d) => (d.tasks || []).length)
  return (
    <div>
      <div className="flex items-center gap-2 mb-4">
        <h2 className="text-xl font-bold">✅ 任务清单</h2>
        <span className="text-sm text-muted-foreground">{project.name} 开发任务逐项进度（checkbox）</span>
      </div>
      {docs.length === 0 ? (
        <div className="text-center text-muted-foreground p-8 bg-card border border-border rounded-xl">暂无任务</div>
      ) : (
        docs.map((d) => {
          const done = (d.tasks || []).filter((t) => t.done).length
          const pct = (d.tasks || []).length ? Math.round((done / (d.tasks || []).length) * 100) : 0
          return (
            <div key={d.path} className="bg-card border border-border rounded-xl p-4 mb-4">
              <div className="flex items-center gap-2 mb-2">
                <TypeTag type={d.type} />
                <button onClick={() => onOpen(d)} className="font-semibold hover:underline cursor-pointer">{d.title}</button>
                <span className="text-xs text-muted-foreground">{d.date || ''}</span>
                <div className="h-1.5 flex-1 bg-muted rounded-full overflow-hidden ml-auto">
                  <div className="h-full bg-primary rounded-full" style={{ width: `${pct}%` }} />
                </div>
                <span className="inline-flex items-center rounded-full bg-primary/10 text-primary text-xs font-semibold px-2 py-0.5 shrink-0">
                  {done}/{d.tasks.length}
                </span>
              </div>
              <ul className="grid grid-cols-1 md:grid-cols-2 gap-x-6 gap-y-0.5">
                {d.tasks.map((t, i) => (
                  <li key={i} className={`flex items-start gap-2 text-[13px] py-0.5 ${t.done ? 'text-muted-foreground line-through' : ''}`}>
                    <input type="checkbox" checked={t.done} disabled className="mt-0.5 size-3.5" />
                    {t.text}
                  </li>
                ))}
              </ul>
            </div>
          )
        })
      )}
    </div>
  )
}
