import { useEffect, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchData, subscribeSSE } from './lib/api'
import type { Project } from './lib/types'
import ProjectBoard from './components/ProjectBoard'
import ProjectDetail from './components/ProjectDetail'
import Header from './components/Header'

export default function App() {
  const qc = useQueryClient()
  const { data, isLoading, error } = useQuery({
    queryKey: ['aggregate'],
    queryFn: fetchData,
  })
  // 当前项目名（null=首页多项目看板）
  const [curName, setCurName] = useState<string | null>(null)

  // SSE 自动刷新
  useEffect(() => {
    return subscribeSSE(() => qc.invalidateQueries({ queryKey: ['aggregate'] }))
  }, [qc])

  // 当前项目对象
  const curProject: Project | undefined =
    data?.projects.find((p) => p.name === curName) ?? undefined

  // 若所选项目已被移除，回首页
  useEffect(() => {
    if (curName && data && !data.projects.some((p) => p.name === curName)) {
      setCurName(null)
    }
  }, [data, curName])

  return (
    <div className="min-h-screen flex flex-col">
      <Header
        total={data?.total_projects ?? 0}
        globalDone={data?.global_tasks_done ?? 0}
        globalTotal={data?.global_tasks_total ?? 0}
        globalCompletion={data?.global_completion ?? 0}
        generatedAt={data?.generated_at ?? ''}
      />
      <div className="flex-1">
        {isLoading && <div className="p-8 text-center text-muted-foreground">加载中…</div>}
        {error && <div className="p-8 text-center text-destructive">加载失败: {(error as Error).message}</div>}
        {data && (
          curProject ? (
            <ProjectDetail
              project={curProject}
              projects={data.projects}
              onBack={() => setCurName(null)}
              onSwitch={setCurName}
            />
          ) : (
            <ProjectBoard projects={data.projects} onEnter={setCurName} />
          )
        )}
      </div>
    </div>
  )
}
