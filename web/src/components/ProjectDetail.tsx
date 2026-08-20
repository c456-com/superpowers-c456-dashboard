import { useState } from 'react'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from './ui/accordion'
import { DOC_TYPES, TYPE_LABEL } from '../lib/types'
import type { Document, Project } from '../lib/types'
import DocDrawer from './DocDrawer'
import OverviewView from './views/OverviewView'
import RoadmapView from './views/RoadmapView'
import MindmapView from './views/MindmapView'
import GanttView from './views/GanttView'
import TasksView from './views/TasksView'

interface Props {
  project: Project
  view: string
  onViewChange: (v: string) => void
}

const VIEW_TABS: [string, string][] = [
  ['overview', '📊 总览'],
  ['roadmap', '🗺️ Roadmap'],
  ['mindmap', '🕸️ 功能图谱'],
  ['gantt', '📅 开发计划'],
  ['tasks', '✅ 任务清单'],
]

export default function ProjectDetail({ project, view, onViewChange }: Props) {
  const [openDoc, setOpenDoc] = useState<Document | null>(null)

  return (
    <div className="flex h-[calc(100vh-56px)]">
      {/* 左侧：手风琴文档索引 */}
      <aside className="w-60 shrink-0 bg-white border-r border-border overflow-y-auto p-3">
        <Accordion defaultValue={['spec']}>
          {DOC_TYPES.map((t) => {
            const list = project.documents.filter((d) => d.type === t).sort((a, b) => (b.date || '').localeCompare(a.date || ''))
            if (list.length === 0) return null
            return (
              <AccordionItem key={t} value={t}>
                <AccordionTrigger className="py-2 text-[13px] font-semibold">
                  {TYPE_LABEL[t]}
                  <span className="ml-auto mr-1 text-xs bg-gray-100 rounded-full px-2">{list.length}</span>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="flex flex-col gap-0.5">
                    {list.map((d) => (
                      <button
                        key={d.path}
                        onClick={() => setOpenDoc(d)}
                        className="text-left text-[13px] text-gray-700 hover:bg-gray-50 rounded px-2 py-1.5 flex items-center gap-2"
                      >
                        <span className="truncate flex-1">{d.title.slice(0, 18)}{d.title.length > 18 ? '…' : ''}</span>
                        <span className="text-[11px] text-gray-400 shrink-0">{d.date || ''}</span>
                      </button>
                    ))}
                  </div>
                </AccordionContent>
              </AccordionItem>
            )
          })}
        </Accordion>
      </aside>

      {/* 主区 */}
      <main className="flex-1 overflow-y-auto p-5 min-w-0">
        {/* 视图页签 */}
        <div className="flex gap-1.5 mb-4 flex-wrap">
          {VIEW_TABS.map(([k, label]) => (
            <button
              key={k}
              onClick={() => onViewChange(k)}
              className={`px-3.5 py-1.5 rounded-lg text-[13px] font-medium border transition-colors ${
                view === k
                  ? 'bg-primary text-white border-primary'
                  : 'bg-white text-gray-500 border-border hover:border-primary hover:text-primary'
              }`}
            >
              {label}
            </button>
          ))}
        </div>

        {view === 'overview' && <OverviewView project={project} onOpen={setOpenDoc} />}
        {view === 'roadmap' && <RoadmapView project={project} />}
        {view === 'mindmap' && <MindmapView project={project} onOpen={setOpenDoc} />}
        {view === 'gantt' && <GanttView project={project} />}
        {view === 'tasks' && <TasksView project={project} onOpen={setOpenDoc} />}
      </main>

      <DocDrawer doc={openDoc} onClose={() => setOpenDoc(null)} />
    </div>
  )
}
