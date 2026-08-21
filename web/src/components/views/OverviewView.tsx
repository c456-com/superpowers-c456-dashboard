import type { Document, Project } from '../../lib/types'
import { TYPE_LABEL, normalizeStatus } from '../../lib/types'

interface Props {
  project: Project
  onOpen: (d: Document) => void
}

export default function OverviewView({ project, onOpen }: Props) {
  const s = project.stats
  const cards: [string, number | string, string][] = [
    ['文档总数', s.total_docs, '篇'],
    ['功能设计', s.specs_total, 'spec'],
    ['开发计划', s.plans_total, 'plan'],
    ['冲刺', s.sprints_total, 'sprint'],
    ['路线图', s.roadmaps_total, ''],
    ['任务完成率', s.completion, '%'],
  ]
  return (
    <div>
      <div className="flex items-center gap-2 mb-4">
        <h2 className="text-xl font-bold">总览 · {project.name}</h2>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-5">
        {cards.map(([k, v, unit]) => (
          <div key={k} className="bg-white border border-border rounded-xl p-4">
            <div className="text-xs text-muted-foreground mb-1">{k}</div>
            <div className="text-2xl font-bold">
              {v}<small className="text-sm text-muted-foreground font-medium ml-1">{unit}</small>
            </div>
          </div>
        ))}
      </div>

      {project.roadmap_stages.length > 0 && (
        <div className="bg-white border border-border rounded-xl p-4 mb-4">
          <h3 className="font-bold mb-3">🗺️ Roadmap</h3>
          <div className="flex flex-col gap-1.5">
            {project.roadmap_stages.map((st) => (
              <div key={st.id} className="flex items-baseline gap-2.5">
                <span className="shrink-0 inline-flex items-center rounded-full bg-blue-50 text-blue-600 text-xs font-semibold px-2.5 py-0.5">
                  {st.id}
                </span>
                <span className="font-semibold">{st.title}</span>
                <span className="text-xs text-muted-foreground truncate">{st.desc?.slice(0, 70)}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {['spec', 'plan', 'sprint', 'research'].map((t) => {
        const list = project.documents.filter((d) => d.type === t)
        if (list.length === 0) return null
        return (
          <div key={t} className="bg-white border border-border rounded-xl p-4 mb-4">
            <h3 className="font-bold mb-3">
              {TYPE_LABEL[t]}
              <span className="ml-2 inline-flex items-center rounded-full bg-blue-50 text-blue-600 text-xs font-semibold px-2 py-0.5">{list.length}</span>
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {list.map((d) => (
                <button key={d.path} onClick={() => onOpen(d)} className="text-left bg-white border border-border rounded-lg p-3.5 hover:border-primary transition-colors cursor-pointer">
                  <div className="font-semibold flex items-center gap-2">
                    <TypeTag type={d.type} />
                    <span className="truncate">
                      <span title={d.status || '文档'}>{StatusEmoji(d.status)} </span>
                      {d.title}
                    </span>
                  </div>
                  <div className="text-xs text-muted-foreground mt-1 line-clamp-2">{d.summary || ''}</div>
                  <div className="text-[11px] text-gray-400 mt-2">
                    {d.date || '无日期'}
                    {d.status ? ' · ' + d.status : ''}
                    {(d.tasks || []).length ? ` · ${(d.tasks || []).filter((x) => x.done).length}/${(d.tasks || []).length} 项` : ''}
                  </div>
                </button>
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}

export function TypeTag({ type }: { type: string }) {
  const map: Record<string, string> = {
    requirement: 'text-rose-600 bg-rose-50',
    research: 'text-gray-600 bg-gray-100',
    story: 'text-teal-600 bg-teal-50',
    product: 'text-fuchsia-600 bg-fuchsia-50',
    spec: 'text-blue-600 bg-blue-50',
    roadmap: 'text-amber-600 bg-amber-50',
    plan: 'text-green-600 bg-green-50',
    sprint: 'text-violet-600 bg-violet-50',
    doc: 'text-gray-600 bg-gray-100',
  }
  return (
    <span className={`inline-flex items-center rounded-md text-[11px] font-semibold px-1.5 py-0.5 ${map[type] || map.doc}`}>
      {TYPE_LABEL[type] || type}
    </span>
  )
}

// 文档状态 → emoji（快速识别已批准/草稿等）
export function StatusEmoji(status?: string): string {
  return normalizeStatus(status).emoji
}
