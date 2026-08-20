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
  spec: '功能设计',
  plan: '开发计划',
  sprint: '冲刺',
  roadmap: '路线图',
  research: '调研',
  doc: '文档',
}

export const DOC_TYPES = ['spec', 'plan', 'sprint', 'roadmap', 'research', 'doc']
