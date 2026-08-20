import { Activity, Settings } from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
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
  onManage: () => void
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
  onManage,
}: Props) {
  const inProject = !!currentProject

  return (
    <header className="sticky top-0 z-50 bg-white border-b border-border h-14 px-6 flex items-center gap-4">
      {/* 左上角品牌：点击返回全部项目 */}
      <button
        onClick={onBack}
        className="font-bold text-base tracking-wide flex items-center gap-2 cursor-pointer hover:text-primary transition-colors whitespace-nowrap"
        title="返回全部项目"
      >
        <Activity className="size-4 text-primary" />
        多项目监控
        {!inProject && (
          <span className="text-sm font-medium text-muted-foreground">· {projects.length} 个项目</span>
        )}
      </button>

      {/* 进入项目：品牌旁显示项目切换下拉（trigger 单行；选项两行：名称+URL） */}
      {inProject && currentProject && (
        <Select value={currentProject.name} onValueChange={(v) => v && onSwitch(v)}>
          <SelectTrigger className="w-56 h-8 text-left px-2.5" aria-label="切换项目">
            <span className="truncate">{currentProject.name}</span>
            <span className="sr-only">{currentProject.name}</span>
          </SelectTrigger>
          <SelectContent className="w-[440px]">
            {projects.map((p) => (
              <SelectItem key={p.name} value={p.name}>
                <div className="flex flex-col leading-tight w-full">
                  <span className="font-semibold text-[13px]">
                    {p.name}
                    {p.status ? <span className="text-muted-foreground font-normal"> · {p.status}</span> : null}
                  </span>
                  <span className="text-[11px] text-muted-foreground truncate" title={p.root}>{p.root}</span>
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}

      {/* 进度徽标：首页=全局任务；项目页=当前项目进度+URL */}
      {inProject && currentProject ? (
        <span className="inline-flex items-center gap-1.5 rounded-full bg-blue-50 text-primary text-xs font-semibold px-3 py-1 max-w-[40%]">
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
      <div className="text-xs text-muted-foreground whitespace-nowrap">更新于 {generatedAt}</div>
      <div className="text-xs text-muted-foreground flex items-center gap-1 whitespace-nowrap">
        <span className="inline-block size-2 rounded-full bg-green-500" />
        自动刷新
      </div>
      <button
        type="button"
        onClick={onManage}
        className="inline-flex items-center gap-1 rounded-md border border-border px-2.5 py-1.5 text-xs font-medium hover:bg-muted transition-colors whitespace-nowrap"
        title="管理项目（添加 / 移除 / 扫描目录）"
      >
        <Settings className="size-3.5" /> 管理项目
      </button>
    </header>
  )
}
