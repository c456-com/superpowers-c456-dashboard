import type { Document, Project } from '../../lib/types'
import { StatusEmoji, TypeTag } from './OverviewView'

interface Props {
  project: Project
  onOpen: (d: Document) => void
}

// 冲刺视图：置顶展示（频繁看），Kanban 风格卡墙，重状态与任务进度
export default function SprintView({ project, onOpen }: Props) {
  const sprints = project.documents
    .filter((d) => d.type === 'sprint')
    .sort((a, b) => (b.date || '').localeCompare(a.date || ''))

  if (sprints.length === 0) {
    return (
      <div className="p-6 text-center text-muted-foreground text-sm">
        暂无冲刺文档。在开发工作流最后阶段创建冲刺来拆解任务。
      </div>
    )
  }

  return (
    <div className="p-6">
      <h2 className="text-xl font-bold mb-4">🗓️ 冲刺（{sprints.length}）</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {sprints.map((s) => {
          const done = (s.tasks || []).filter((x) => x.done).length
          const total = (s.tasks || []).length
          const pct = total ? Math.round((done / total) * 100) : 0
          return (
            <button
              key={s.path}
              onClick={() => onOpen(s)}
              className="text-left bg-card border border-border rounded-xl p-4 hover:border-primary transition-colors cursor-pointer"
            >
              <div className="flex items-center gap-2 mb-2">
                <TypeTag type="sprint" />
                <span>{StatusEmoji(s.status)}</span>
              </div>
              <div className="font-semibold leading-snug mb-2 line-clamp-2">{s.title}</div>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span>{s.date || '无日期'}</span>
                {total ? <span>{done}/{total} 项 · {pct}%</span> : <span>无任务</span>}
              </div>
              {s.status && <div className="text-[11px] text-muted-foreground mt-1">状态：{s.status}</div>}
              <div className="mt-2 h-1.5 rounded-full bg-muted overflow-hidden">
                <div className="h-full bg-violet-500 rounded-full" style={{ width: pct + '%' }} />
              </div>
            </button>
          )
        })}
      </div>
    </div>
  )
}
