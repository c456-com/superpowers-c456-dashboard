import { useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchData, subscribeSSE } from './lib/api'
import type { Project } from './lib/types'
import ProjectBoard from './components/ProjectBoard'
import ProjectDetail from './components/ProjectDetail'
import Header from './components/Header'
import ManagePanel from './components/ManagePanel'
import AIPanel from './components/AIPanel'

// hash 路由：`#/` 或空 = 首页多项目看板；`#/项目名` = 单项目（默认总览）；`#/项目名/视图`
function parseHash(): { name: string | null; view: string } {
  const h = window.location.hash.replace(/^#\/?/, '') // 去 #/ 前缀
  if (!h) return { name: null, view: 'overview' }
  const parts = h.split('/').filter(Boolean)
  if (parts.length === 0) return { name: null, view: 'overview' }
  const name = decodeURIComponent(parts[0])
  const view = parts[1] || 'overview'
  return { name, view }
}

export default function App() {
  const qc = useQueryClient()
  const { data, isLoading, error } = useQuery({
    queryKey: ['aggregate'],
    queryFn: fetchData,
  })
  // 从 hash 派生的路由状态
  const [route, setRoute] = useState(parseHash)

  // 监听 hashchange（点浏览器后退/前进、手动改 URL）
  useEffect(() => {
    const onHash = () => setRoute(parseHash())
    window.addEventListener('hashchange', onHash)
    return () => window.removeEventListener('hashchange', onHash)
  }, [])

  // SSE 自动刷新
  useEffect(() => {
    return subscribeSSE(() => qc.invalidateQueries({ queryKey: ['aggregate'] }))
  }, [qc])

  const routes = useMemo(() => ({ home: '', project: (name: string, view?: string) => `#/${encodeURIComponent(name)}${view && view !== 'overview' ? '/' + view : ''}` }), [])

  const navigate = (name: string | null, view?: string) => {
    const target = name ? routes.project(name, view) : routes.home
    if (window.location.hash === target) return
    window.location.hash = target
  }

  // 当前项目（hash 指向但数据中不存在 → 视为首页）
  const curProject: Project | undefined =
    route.name ? data?.projects.find((p) => p.name === route.name) : undefined

  // 若 hash 指向的项目已被移除，回首页（用 replace 避免污染历史）
  useEffect(() => {
    if (route.name && data && !data.projects.some((p) => p.name === route.name)) {
      window.location.replace(window.location.pathname + window.location.search + '#/')
    }
  }, [route.name, data])

  const view = curProject ? route.view : 'overview'

  const handleSwitch = (name: string) => navigate(name, view)
  const handleViewChange = (v: string) => navigate(route.name, v)
  const handleBack = () => navigate(null)

  const [manageOpen, setManageOpen] = useState(false)
  const [aiOpen, setAiOpen] = useState(false)
  const handleChanged = () => {
    qc.invalidateQueries({ queryKey: ['aggregate'] })
  }
  const handleEnter = (name: string) => {
    setManageOpen(false)
    navigate(name)
  }

  return (
    <div className="min-h-screen flex flex-col">
      <Header
        projects={data?.projects ?? []}
        currentProject={curProject ?? null}
        globalDone={data?.global_tasks_done ?? 0}
        globalTotal={data?.global_tasks_total ?? 0}
        globalCompletion={data?.global_completion ?? 0}
        generatedAt={data?.generated_at ?? ''}
        onBack={handleBack}
        onSwitch={handleSwitch}
        onManage={() => setManageOpen(true)}
        onAI={() => setAiOpen(true)}
      />
      <div className="flex-1">
        {isLoading && <div className="p-8 text-center text-muted-foreground">加载中…</div>}
        {error && <div className="p-8 text-center text-destructive">加载失败: {(error as Error).message}</div>}
        {data && (
          curProject ? (
            <ProjectDetail
              project={curProject}
              view={view}
              onViewChange={handleViewChange}
            />
          ) : (
            <ProjectBoard projects={data.projects} onEnter={(n) => navigate(n)} />
          )
        )}
      </div>
      <ManagePanel
        projects={data?.projects ?? []}
        open={manageOpen}
        onClose={() => setManageOpen(false)}
        onChanged={handleChanged}
        onEnter={handleEnter}
      />
      <AIPanel
        open={aiOpen}
        onOpenChange={setAiOpen}
        project={curProject?.name}
      />
    </div>
  )
}
