import { useEffect, useState } from 'react'
import { Bot, Sparkles } from 'lucide-react'
import { Sheet, SheetContent, SheetTitle } from './ui/sheet'
import { Button } from './ui/button'
import {
  getAIConfig,
  saveAIConfig,
  analyseProject,
  getSuggestions,
  type AIConfig,
  type AISuggestion,
} from '../lib/api'

interface Props {
  open: boolean
  onOpenChange: (v: boolean) => void
  project?: string
}

const SEV: Record<string, string> = {
  success: 'text-green-600 border-green-200 bg-green-50',
  warning: 'text-amber-600 border-amber-200 bg-amber-50',
  info: 'text-blue-600 border-blue-200 bg-blue-50',
}

export default function AIPanel({ open, onOpenChange, project }: Props) {
  const [cfg, setCfg] = useState<AIConfig>({ base_url: '', model: '', has_key: false })
  const [baseUrl, setBaseUrl] = useState('')
  const [model, setModel] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null)
  const [running, setRunning] = useState(false)
  const [suggestions, setSuggestions] = useState<AISuggestion[]>([])
  const [at, setAt] = useState('')

  const load = async () => {
    const c = await getAIConfig().catch(() => ({ base_url: '', model: '', has_key: false }))
    setCfg(c)
    setBaseUrl(c.base_url)
    setModel(c.model)
  }
  useEffect(() => {
    if (open) {
      load()
      getSuggestions().then((r) => { setSuggestions(r.suggestions); setAt(r.at) }).catch(() => {})
    }
  }, [open])

  const save = async () => {
    setBusy(true); setMsg(null)
    try {
      await saveAIConfig({ base_url: baseUrl, model, api_key: apiKey })
      const c = await getAIConfig()
      setCfg(c)
      setMsg({ ok: true, text: 'AI 配置已保存' })
    } catch (e) {
      setMsg({ ok: false, text: (e as Error).message })
    } finally { setBusy(false) }
  }

  const analyse = async () => {
    setRunning(true); setMsg(null); setSuggestions([])
    try {
      await analyseProject(project)
      // 轮询结果（SSE 会推，这里兜底轮询）
      const poll = setInterval(async () => {
        const r = await getSuggestions().catch(() => null)
        if (r && r.at) {
          clearInterval(poll)
          setAt(r.at); setSuggestions(r.suggestions); setRunning(false)
        }
      }, 1500)
    } catch (e) {
      setRunning(false)
      setMsg({ ok: false, text: (e as Error).message })
    }
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="sm:max-w-xl flex flex-col p-0">
        <SheetTitle className="sr-only">AI 建议</SheetTitle>
        {/* toolbar */}
        <div className="shrink-0 px-6 py-4 border-b border-border flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Bot className="size-5 text-primary" />
            <span className="font-bold text-base">AI 开发顾问</span>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-5 flex flex-col gap-5">
          {/* AI 模型配置 */}
          <div className="flex flex-col gap-2.5">
            <div className="text-sm font-semibold flex items-center gap-1.5">
              <Sparkles className="size-4 text-primary" /> AI 模型配置
            </div>
            {!cfg.base_url && !cfg.has_key && (
              <div className="text-xs text-muted-foreground bg-muted/40 rounded-md px-3 py-2">
                未配置 AI 模型。填写 OpenAI 兼容端点（本地或远程）后即可让 AI 分析项目给建议。
              </div>
            )}
            <input value={baseUrl} onChange={(e) => setBaseUrl(e.target.value)} placeholder="API Base URL (OpenAI 兼容)，如 http://127.0.0.1:8888/v1"
              className="h-9 rounded-md border border-input bg-transparent px-3 text-sm" />
            <input value={model} onChange={(e) => setModel(e.target.value)} placeholder="模型名，如 Qwen3-35B-A3B"
              className="h-9 rounded-md border border-input bg-transparent px-3 text-sm" />
            <input type="password" value={apiKey} onChange={(e) => setApiKey(e.target.value)} placeholder={cfg.has_key ? 'API Key（已保存，留空表示不修改）' : 'API Key（本地端点可留空）'}
              className="h-9 rounded-md border border-input bg-transparent px-3 text-sm" />
            <Button size="sm" onClick={save} disabled={busy || !baseUrl.trim() || !model.trim()} className="self-start">
              保存配置
            </Button>
          </div>

          {/* AI 分析触发 */}
          <div className="border-t border-border pt-4 flex flex-col gap-2.5">
            <div className="text-sm font-semibold">AI 分析项目</div>
            <div className="text-xs text-muted-foreground">
              {project ? `分析项目「${project}」的文档结构、AGENTS.md、数据模型，给出改进建议。` : '分析全部项目的开发最佳实践合规性，给出改进建议。'}
            </div>
            <Button size="sm" onClick={analyse} disabled={running || !cfg.base_url || !cfg.model} className="self-start">
              {running ? '分析中…' : '开始分析'}
            </Button>
          </div>

          {/* 建议列表 */}
          {suggestions.length > 0 && (
            <div className="border-t border-border pt-4 flex flex-col gap-2">
              <div className="text-sm font-semibold flex items-center justify-between">
                建议 {at ? `(更新于 ${at})` : ''}
                <span className="text-xs text-muted-foreground font-normal">{suggestions.length} 条</span>
              </div>
              {suggestions.map((s, i) => (
                <div key={i} className={`rounded-lg border px-3.5 py-3 text-sm ${SEV[s.severity] || 'text-blue-600 border-blue-200 bg-blue-50'}`}>
                  <div className="font-semibold">{s.title}</div>
                  <div className="text-[13px] mt-1 opacity-90 whitespace-pre-wrap">{s.detail}</div>
                  {s.action && <div className="text-xs mt-1.5 font-medium">建议行动：{s.action}</div>}
                </div>
              ))}
            </div>
          )}

          {msg && (
            <div className={`text-sm ${msg.ok ? 'text-green-600' : 'text-destructive'}`}>{msg.text}</div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}