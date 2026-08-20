import type { Document, Project } from '../../lib/types'
import { TypeTag } from './OverviewView'

interface Props {
  project: Project
  onOpen: (d: Document) => void
}

export default function MindmapView({ project, onOpen }: Props) {
  const specs = project.documents.filter((d) => d.type === 'spec')
  return (
    <div>
      <div className="flex items-center gap-2 mb-4">
        <h2 className="text-xl font-bold">🕸️ 功能设计图谱</h2>
        <span className="text-sm text-muted-foreground">spec → 章节，点击节点查看功能详情</span>
      </div>
      {specs.length === 0 ? (
        <div className="text-center text-muted-foreground p-8 bg-white border border-border rounded-xl">暂无功能设计文档</div>
      ) : (
        <div>
          {/* 根节点 */}
          <div className="flex justify-center mb-6">
            <div className="bg-primary text-white rounded-lg px-5 py-2.5 font-semibold">
              {project.name} · 功能设计 ({specs.length})
            </div>
          </div>
          {/* 功能节点网格 */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {specs.map((d) => (
              <div key={d.path} className="bg-white border border-border rounded-xl p-4">
                <button
                  onClick={() => onOpen(d)}
                  className="text-left w-full font-semibold text-blue-600 hover:underline flex items-center gap-2 cursor-pointer"
                >
                  <TypeTag type={d.type} />
                  <span className="truncate">{d.title}</span>
                </button>
                <div className="mt-2 flex flex-col gap-1">
                {(d.sections || []).slice(0, 6).map((sec, j) => (
                    <div key={j} className="flex items-center gap-2 text-[13px] text-gray-600 pl-2 border-l-2 border-gray-100">
                      <span className="size-1.5 rounded-full bg-gray-300 shrink-0" />
                      <span className="truncate">{sec.title}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
