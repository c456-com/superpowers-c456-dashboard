import { useState } from 'react'
import type { Document, Project } from '../../lib/types'
import { TYPE_LABEL, TYPE_ORDER, normalizeStatus, TYPE_COLOR } from '../../lib/types'
import { ChevronRight } from 'lucide-react'

interface Props {
  project: Project
  onOpen: (d: Document) => void
}

// 工作流阶段（排除 doc）
const FLOW = TYPE_ORDER.filter((t) => t !== 'doc')

export default function OverviewView({ project, onOpen }: Props) {
  const [sel, setSel] = useState<string | null>(null)

  return (
    <div className="p-6">
      <div className="flex items-center gap-2 mb-5">
        <h2 className="text-xl font-bold">开发最佳实践工作流 · {project.name}</h2>
      </div>

      {/* 工作流 8 阶段卡片（横向串联） */}
      <div className="flex items-stretch gap-0 overflow-x-auto pb-2">
        {FLOW.map((t, i) => {
          const list = project.documents.filter((d) => d.type === t)
          const done = list.reduce((acc, d) => acc + (d.tasks || []).filter((x) => x.done).length, 0)
          const total = list.reduce((acc, d) => acc + (d.tasks || []).length, 0)
          const pct = total ? Math.round((done / total) * 100) : 0
          const active = sel === t
          // 该类型色（dark 可读）
          const color = TYPE_COLOR[t] || { border: 'border-border', text: 'text-muted-foreground', tag: '' }

          return (
            <div key={t} className="flex items-stretch shrink-0">
              <button
                type="button"
                onClick={() => setSel(active ? null : t)}
                className={`w-44 text-left rounded-xl border-2 bg-card p-4 flex flex-col gap-2 transition-colors cursor-pointer ${color.border} ${color.text} ${active ? 'ring-2 ring-primary/30' : ''}`}
              >
                <div className="font-bold text-sm">{TYPE_LABEL[t]}</div>
                {list.length === 0 ? (
                  <div className="text-xs text-destructive/80">缺文档 · 建议创建</div>
                ) : (
                  <>
                    <div className="text-xs text-muted-foreground">{list.length} 篇</div>
                    <div className="flex items-center gap-1.5">
                      <div className="flex-1 h-1.5 rounded-full bg-muted overflow-hidden">
                        <div className="h-full bg-current rounded-full" style={{ width: pct + '%' }} />
                      </div>
                      <span className="text-[10px] text-muted-foreground">{pct}%</span>
                    </div>
                  </>
                )}
                <div className="text-[10px] text-muted-foreground flex items-center gap-0.5 mt-auto pt-1">
                  查看{list.length ? ` ${list.length} 篇` : ''} <ChevronRight className="size-3" />
                </div>
              </button>
              {/* 阶段间连线箭头 */}
              {i < FLOW.length - 1 && (
                <div className="flex items-center px-0.5 text-muted-foreground/70">
                  <ChevronRight className="size-5" />
                </div>
              )}
            </div>
          )
        })}
      </div>

      {/* 阶段下钻：点击阶段卡后展示该类型文档列表 */}
      {sel && (
        <ListPanel type={sel} list={project.documents.filter((d) => d.type === sel)} onOpen={onOpen} />
      )}

      {/* 底部：冲刺（置顶语义下最常看）+ 统计 */}
      <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-6 gap-3 mt-6">
        <StatRow project={project} />
      </div>
    </div>
  )
}

function StatRow({ project }: { project: Project }) {
  const s = project.stats
  const research = s.by_type?.['research'] ?? 0
  const items: [string, number|string, string][] = [
    ['文档总数', s.total_docs, '篇'],
    ['任务', `${s.tasks_done}/${s.tasks_total}`, '项'],
    ['任务完成率', s.completion, '%'],
    ['Roadmap', project.roadmap_stages.length, '阶段'],
    ['冲刺', s.sprints_total, '个'],
    ['调研', research, '篇'],
  ]
  return (
    <>
      {items.map(([k, v, u]) => (
        <div key={k} className="bg-card border border-border rounded-xl p-3.5">
          <div className="text-xs text-muted-foreground mb-1">{k}</div>
          <div className="text-xl font-bold">{v}<small className="text-xs text-muted-foreground ml-0.5">{u}</small></div>
        </div>
      ))}
    </>
  )
}

function ListPanel({ type, list, onOpen }: { type: string; list: Document[]; onOpen: (d: Document) => void }) {
  return (
    <div className="bg-card border border-border rounded-xl p-4 mt-5">
      <h3 className="font-bold mb-3">
        {TYPE_LABEL[type]}
        <span className="ml-2 inline-flex items-center rounded-full bg-muted text-muted-foreground text-xs font-semibold px-2 py-0.5">{list.length}</span>
      </h3>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        {list.map((d) => (
          <button key={d.path} onClick={() => onOpen(d)} className="text-left bg-card border border-border rounded-lg p-3.5 hover:border-primary transition-colors cursor-pointer">
            <div className="font-semibold flex items-center gap-2">
              <TypeTag type={d.type} />
              <span className="truncate">
                <span title={d.status || '文档'}>{StatusEmoji(d.status)} </span>
                {d.title}
              </span>
            </div>
            <div className="text-xs text-muted-foreground mt-1 line-clamp-2">{d.summary || ''}</div>
            <div className="text-[11px] text-muted-foreground mt-2 flex items-center gap-2">
              <span>{d.date || '无日期'}</span>
              {(d.tasks || []).length ? <span>{d.tasks.filter((x) => x.done).length}/{d.tasks.length} 项</span> : null}
            </div>
          </button>
        ))}
      </div>
    </div>
  )
}

export function TypeTag({ type }: { type: string }) {
  const c = TYPE_COLOR[type] || TYPE_COLOR.doc
  return (
    <span className={`inline-flex items-center rounded-md text-[11px] font-semibold px-1.5 py-0.5 ${c.tag}`}>
      {TYPE_LABEL[type] || type}
    </span>
  )
}

// 文档状态 → emoji（快速识别已批准/草稿等）
export function StatusEmoji(status?: string): string {
  return normalizeStatus(status).emoji
}
