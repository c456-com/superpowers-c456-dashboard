// 与 Go 后端 /data 返回结构对应的类型定义

export interface Task {
  text: string
  done: boolean
}

export interface Section {
  level: number
  title: string
}

export interface RoadmapStage {
  id: string
  title: string
  desc: string
}

export interface Meta {
  [key: string]: string
}

export interface Document {
  path: string
  title: string
  date: string
  type: string
  status: string
  meta: Meta
  summary: string
  sections: Section[]
  tasks: Task[]
  content: string
  mtime: number
}

export interface Stats {
  total_docs: number
  by_type: Record<string, number>
  tasks_total: number
  tasks_done: number
  completion: number
  specs_total: number
  plans_total: number
  sprints_total: number
  roadmaps_total: number
  last_scan: string
}

export interface Project {
  name: string
  root: string
  generated_at: string
  type?: string
  status?: string
  stats: Stats
  documents: Document[]
  roadmap_stages: RoadmapStage[]
}

export interface Aggregate {
  projects: Project[]
  generated_at: string
  total_projects: number
  global_tasks_total: number
  global_tasks_done: number
  global_completion: number
  global_docs_total: number
}

export const TYPE_LABEL: Record<string, string> = {
  requirement: '客户需求',
  research: '调研',
  story: '用户故事',
  product: '产品设计',
  spec: '功能设计',
  roadmap: '路线图',
  plan: '开发计划',
  sprint: '冲刺',
  doc: '文档',
}

// 类型语义色（dark 变体）：明暗都可读的边框/Tag 色，一处维护
export const TYPE_COLOR: Record<string, { border: string; text: string; tag: string }> = {
  requirement: { border: 'border-rose-300 dark:border-rose-500/50', text: 'text-rose-700 dark:text-rose-400', tag: 'text-rose-700 dark:text-rose-300 bg-rose-500/10' },
  research: { border: 'border-border', text: 'text-muted-foreground', tag: 'text-muted-foreground bg-muted' },
  story: { border: 'border-teal-300 dark:border-teal-500/50', text: 'text-teal-700 dark:text-teal-400', tag: 'text-teal-700 dark:text-teal-300 bg-teal-500/10' },
  product: { border: 'border-fuchsia-300 dark:border-fuchsia-500/50', text: 'text-fuchsia-700 dark:text-fuchsia-400', tag: 'text-fuchsia-700 dark:text-fuchsia-300 bg-fuchsia-500/10' },
  spec: { border: 'border-blue-300 dark:border-blue-500/50', text: 'text-blue-700 dark:text-blue-400', tag: 'text-blue-700 dark:text-blue-300 bg-blue-500/10' },
  roadmap: { border: 'border-amber-300 dark:border-amber-500/50', text: 'text-amber-700 dark:text-amber-400', tag: 'text-amber-700 dark:text-amber-300 bg-amber-500/10' },
  plan: { border: 'border-green-300 dark:border-green-500/50', text: 'text-green-700 dark:text-green-400', tag: 'text-green-700 dark:text-green-300 bg-green-500/10' },
  sprint: { border: 'border-violet-300 dark:border-violet-500/50', text: 'text-violet-700 dark:text-violet-400', tag: 'text-violet-700 dark:text-violet-300 bg-violet-500/10' },
  doc: { border: 'border-border', text: 'text-muted-foreground', tag: 'text-muted-foreground bg-muted' },
}

// 工作流 8 阶段展示顺序（首页工作流 + 侧栏索引排序）
export const TYPE_ORDER = [
  'requirement',
  'research',
  'story',
  'product',
  'spec',
  'roadmap',
  'plan',
  'sprint',
  'doc',
]

export const DOC_TYPES = TYPE_ORDER

// 文档状态 → emoji/label/色 归一化映射（辉哥定：列表标题前加 emoji 快速识别）
export interface StatusMeta {
  label: string
  emoji: string
  color: string
}
export const STATUS_META: { match: string[]; meta: StatusMeta }[] = [
  { match: ['approved', 'approve', '正式方案', '已验证', '已定稿', '已批准', 'accepted'], meta: { label: '已批准', emoji: '✅', color: 'green' } },
  { match: ['draft', '未验证', '未定稿', '草案', '草稿', 'draft（未验证）'], meta: { label: '草稿', emoji: '🔶', color: 'amber' } },
  { match: ['提案', 'proposal', '建议'], meta: { label: '提案', emoji: '💡', color: 'blue' } },
  { match: ['废弃', 'deprecated', '已删除', '已作废', 'obsolete'], meta: { label: '已废弃', emoji: '🗑️', color: 'gray' } },
  { match: ['评审中', 'review', '审核中'], meta: { label: '评审中', emoji: '🔄', color: 'blue' } },
]

export const STATUS_DEFAULT: StatusMeta = { label: '文档', emoji: '📄', color: 'neutral' }

// 归一化任意状态字符串 → StatusMeta（大小写/中英/带** 的忽略）
export function normalizeStatus(status?: string): StatusMeta {
  if (!status) return STATUS_DEFAULT
  const s = status.toLowerCase().replace(/\*\*/g, '').trim()
  for (const group of STATUS_META) {
    for (const m of group.match) {
      if (s.includes(m.toLowerCase())) return group.meta
    }
  }
  return STATUS_DEFAULT
}
