import type { Project } from '../../lib/types'

interface Props {
  project: Project
}

export default function RoadmapView({ project }: Props) {
  const stages = project.roadmap_stages
  return (
    <div>
      <div className="flex items-center gap-2 mb-4">
        <h2 className="text-xl font-bold">🗺️ Roadmap 时间线</h2>
        <span className="text-sm text-muted-foreground">{project.name} 开发阶段演进</span>
      </div>
      {stages.length === 0 ? (
        <div className="text-center text-muted-foreground p-8 bg-card border border-border rounded-xl">
          未识别到 Roadmap 阶段（需 roadmap 文档含「阶段X」小节）
        </div>
      ) : (
        <div className="bg-card border border-border rounded-xl p-5">
          <ol className="relative border-l-2 border-border ml-2 pl-6 space-y-6">
            {stages.map((st) => (
              <li key={st.id} className="relative">
                <span className="absolute -left-[34px] flex items-center justify-center size-5 rounded-full bg-primary text-primary-foreground text-[11px] font-bold">
                  {st.id}
                </span>
                <div className="font-semibold">{st.title}</div>
                {st.desc && <div className="text-xs text-muted-foreground mt-0.5">{st.desc}</div>}
              </li>
            ))}
          </ol>
        </div>
      )}
    </div>
  )
}
