import { Activity } from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select'
import type { Project } from '../lib/types'

interface Props {
  projects: Project[]
  currentProject: Project | null
  globalDone: number
  globalTotal: number
  globalCompletion: number
  generatedAt: string
  onBack: () => void
  onSwitch: (name: string) => void
}

export default function Header({
  projects,
  currentProject,
  globalDone,
  globalTotal,
  globalCompletion,
  generatedAt,
  onBack,
  onSwitch,
}: Props) {
  const inProject = !!currentProject

  return (
    <header className="sticky top-0 z-50 bg-white border-b border-border px-6 py-3 flex items-center gap-4 flex-wrap">
      {/* 左上角品牌：点击返回全部项目 */}
      <button
        onClick={onBack}
        className="font-bold text-base tracking-wide flex items-center gap-2 cursor-pointer hover:text-primary transition-colors"
        title="返回全部项目"
      >
        <Activity className="size-4 text-primary" />
        多项目监控
        {!inProject && (
          <span className="text-sm font-medium text-muted-foreground">· {projects.length} 个项目</span>
        )}
      </button>

      {/* 进入项目：品牌旁显示项目切换下拉（两行：名称+URL） */}
      {inProject && currentProject && (
        <Select value={currentProject.name} onValueChange={(v) => v && onSwitch(v)}>
          <SelectTrigger className="w-72 h-auto py-1.5 text-left" aria-label="切换项目">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {projects.map((p) => (
              <SelectItem key={p.name} value={p.name}>
                <div className="flex flex-col leading-tight">
                  <span className="font-semibold text-[13px]">
                    {p.name}
                    {p.status ? <span className="text-muted-foreground font-normal"> · {p.status}</span> : null}
                  </span>
                  <span className="text-[11px] text-muted-foreground truncate max-w-[260px]">{p.root}</span>
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}

      {/* 进度徽标：首页=全局任务；项目页=当前项目进度+URL */}
      {inProject && currentProject ? (
        <span className="inline-flex items-center gap-1.5 rounded-full bg-blue-50 text-primary text-xs font-semibold px-3 py-1 max-w-[45%]">
          进度 {currentProject.stats.tasks_done}/{currentProject.stats.tasks_total} · {currentProject.stats.completion}%
          <span className="text-primary/60 font-normal truncate" title={currentProject.root}>
            {currentProject.root}
          </span>
        </span>
      ) : (
        <span className="inline-flex items-center rounded-full bg-blue-50 text-primary text-xs font-semibold px-3 py-1">
          全局任务 {globalDone}/{globalTotal} · {globalCompletion}%
        </span>
      )}

      <div className="flex-1" />
      <div className="text-xs text-muted-foreground">更新于 {generatedAt}</div>
      <div className="text-xs text-muted-foreground flex items-center gap-1">
        <span className="inline-block size-2 rounded-full bg-green-500" />
        自动刷新
      </div>
    </header>
  )
}
