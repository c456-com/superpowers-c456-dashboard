import { useState } from 'react'
import { Plus, Trash2, ScanSearch, FolderPlus } from 'lucide-react'
import { Sheet, SheetContent, SheetTitle } from './ui/sheet'
import { Button } from './ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from './ui/alert-dialog'
import { addProject, removeProject, scanDir, type ScanCandidate } from '../lib/api'
import type { Project } from '../lib/types'

interface Props {
  projects: Project[]
  open: boolean
  onClose: () => void
  onChanged: () => void // 数据已变，触发 invalidate
  onEnter: (name: string) => void
}

export default function ManagePanel({ projects, open, onClose, onChanged, onEnter }: Props) {
  const [addPath, setAddPath] = useState('')
  const [addName, setAddName] = useState('')
  const [scanPath, setScanPath] = useState('')
  const [candidates, setCandidates] = useState<ScanCandidate[] | null>(null)
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null)
  const [confirmRemove, setConfirmRemove] = useState<Project | null>(null)

  const flash = (ok: boolean, text: string) => setMsg({ ok, text })

  const handleAdd = async () => {
    if (!addPath.trim()) return
    setBusy(true); setMsg(null)
    try {
      const name = await addProject(addPath.trim(), addName.trim() || undefined)
      flash(true, `已添加项目「${name}」`)
      setAddPath(''); setAddName('')
      onChanged()
    } catch (e) {
      flash(false, (e as Error).message)
    } finally {
      setBusy(false)
    }
  }

  const handleScan = async () => {
    if (!scanPath.trim()) return
    setBusy(true); setMsg(null); setCandidates(null)
    try {
      const c = await scanDir(scanPath.trim())
      if (c.length === 0) { flash(true, '未发现 superpowers 项目（含 specs/plans/roadmap/sprint 文档的目录）'); setCandidates([]) }
      else setCandidates(c)
    } catch (e) {
      flash(false, (e as Error).message); setCandidates([])
    } finally {
      setBusy(false)
    }
  }

  const handleAddCandidate = async (c: ScanCandidate) => {
    if (c.already) return
    setBusy(true); setMsg(null)
    try {
      const name = await addProject(c.path)
      flash(true, `已添加「${name}」`)
      setCandidates((prev) => prev?.map((x) => (x.path === c.path ? { ...x, already: true } : x)) ?? null)
      onChanged()
    } catch (e) {
      flash(false, (e as Error).message)
    } finally {
      setBusy(false)
    }
  }

  const doRemove = async (name: string) => {
    setBusy(true); setMsg(null)
    try {
      await removeProject(name)
      flash(true, `已移除「${name}」`)
      onChanged()
    } catch (e) {
      flash(false, (e as Error).message)
    } finally {
      setBusy(false)
      setConfirmRemove(null)
    }
  }

  return (
    <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
      <SheetContent className="w-full sm:max-w-lg! flex flex-col p-0" side="left">
        <div className="shrink-0 px-6 py-4 border-b border-border">
          <SheetTitle className="text-base font-bold pr-8">⚙️ 管理项目</SheetTitle>
          <div className="text-xs text-muted-foreground mt-1">全局配置随面板自动读写，改动即时生效</div>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-5 flex flex-col gap-6">
          {/* 一键扫描目录 */}
          <section>
            <div className="flex items-center gap-1.5 text-sm font-semibold mb-2">
              <ScanSearch className="size-4 text-primary" /> 扫描目录识别项目
            </div>
            <div className="flex gap-2">
              <input
                value={scanPath}
                onChange={(e) => setScanPath(e.target.value)}
                placeholder="输入目录绝对路径，如 ~/Codes"
                className="flex-1 h-9 rounded-md border border-input bg-transparent px-3 text-sm"
              />
              <Button size="sm" onClick={handleScan} disabled={busy || !scanPath.trim()}>扫描</Button>
            </div>
            {candidates && (
              <div className="mt-3 flex flex-col gap-2">
                {candidates.length === 0 && <div className="text-xs text-muted-foreground">没有匹配结果</div>}
                {candidates.map((c) => (
                  <div key={c.path} className="flex items-center gap-2 rounded-md border border-border p-2 text-sm">
                    <div className="flex-1 min-w-0">
                      <div className="font-medium truncate">{c.name}</div>
                      <div className="text-xs text-muted-foreground truncate">{c.path} · {c.doc_count} 篇文档</div>
                    </div>
                    {c.already ? (
                      <span className="text-xs text-muted-foreground shrink-0">已在面板</span>
                    ) : (
                      <Button size="sm" variant="secondary" onClick={() => handleAddCandidate(c)} disabled={busy}>
                        <FolderPlus className="size-4 mr-1" /> 添加
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* 手动添加 */}
          <section>
            <div className="flex items-center gap-1.5 text-sm font-semibold mb-2">
              <Plus className="size-4 text-primary" /> 手动添加项目
            </div>
            <div className="flex flex-col gap-2">
              <input
                value={addPath}
                onChange={(e) => setAddPath(e.target.value)}
                placeholder="项目目录绝对路径（必填）"
                className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
              />
              <input
                value={addName}
                onChange={(e) => setAddName(e.target.value)}
                placeholder="显示名（可选，默认取目录名）"
                className="h-9 rounded-md border border-input bg-transparent px-3 text-sm"
              />
              <Button size="sm" onClick={handleAdd} disabled={busy || !addPath.trim()}>添加项目</Button>
            </div>
          </section>

          {/* 当前项目列表 + 移除 */}
          <section>
            <div className="text-sm font-semibold mb-2">当前项目（{projects.length}）</div>
            <div className="flex flex-col gap-1.5">
              {projects.map((p) => (
                <div key={p.name} className="flex items-center gap-2 rounded-md border border-border p-2 text-sm group">
                  <button
                    type="button"
                    onClick={() => onEnter(p.name)}
                    className="flex-1 text-left min-w-0 hover:text-primary"
                  >
                    <div className="font-medium truncate">{p.name}</div>
                    <div className="text-xs text-muted-foreground truncate">{p.root}</div>
                  </button>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="text-destructive opacity-60 group-hover:opacity-100"
                    onClick={() => setConfirmRemove(p)}
                    disabled={busy}
                  >
                    <Trash2 className="size-4" />
                  </Button>
                </div>
              ))}
            </div>
          </section>
        </div>

        {msg && (
          <div className={`shrink-0 px-6 py-3 border-t border-border text-sm ${msg.ok ? 'text-green-600' : 'text-destructive'}`}>
            {msg.text}
          </div>
        )}
      </SheetContent>

      {/* 移除项目确认 */}
      <AlertDialog open={!!confirmRemove} onOpenChange={(o) => !o && setConfirmRemove(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认移除项目？</AlertDialogTitle>
            <AlertDialogDescription>
              将把「{confirmRemove?.name}」从监控面板移除（全局配置同步更新）。
              <br />
              <span className="block mt-1 text-xs text-muted-foreground break-all">{confirmRemove?.root}</span>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={busy}>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => confirmRemove && doRemove(confirmRemove.name)} disabled={busy}>
              移除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Sheet>
  )
}
