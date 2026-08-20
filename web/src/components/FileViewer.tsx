import { useEffect, useMemo, useState } from 'react'
import hljs from 'highlight.js/lib/common'
import 'highlight.js/styles/github.css'
import { Sheet, SheetContent, SheetTitle } from './ui/sheet'
import { fetchFile, type FileContent } from '../lib/api'

interface Props {
  project: string | null
  path: string | null
  dir?: string | null
  onClose: () => void
}

// 文件预览：点击文档里的文件路径链接 → 后端安全读取 → 显示源码
export default function FileViewer({ project, path, dir, onClose }: Props) {
  const open = !!project && !!path
  const [data, setData] = useState<FileContent | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!project || !path) return
    setData(null)
    setError(null)
    fetchFile(project, path, dir ?? undefined)
      .then(setData)
      .catch((e) => setError(e.message))
  }, [project, path, dir])

  // 标题取路径最后一段
  const filename = path ? path.split('/').pop() : ''

  // 根据扩展名高亮（ruby/go/yml/json/erb/md → hljs 语言类）
  const highlighted = useMemo(() => {
    if (!data) return ''
    const ext = (filename || '').toLowerCase()
    const lang =
      ext.endsWith('.rb') || ext.endsWith('.erb') ? 'ruby'
      : ext.endsWith('.go') ? 'go'
      : ext.endsWith('.yml') || ext.endsWith('.yaml') ? 'yaml'
      : ext.endsWith('.json') ? 'json'
      : ext.endsWith('.md') ? 'markdown'
      : ext.endsWith('.tf') ? 'hcl'
      : ext.endsWith('.html') ? 'xml'
      : ext.endsWith('.css') ? 'css'
      : ext.endsWith('.js') || ext.endsWith('.ts') ? 'typescript'
      : ''
    try {
      return lang && hljs.getLanguage(lang) ? hljs.highlight(data.content, { language: lang }).value : data.content
    } catch {
      return data.content
    }
  }, [data, filename])

  return (
    <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
      <SheetContent className="w-full sm:max-w-3xl! flex flex-col p-0" side="right">
        <div className="shrink-0 px-6 py-4 border-b border-border">
          <SheetTitle className="text-base font-bold flex items-center gap-2 pr-8">
            📄 {filename}
            {data && <span className="text-xs font-normal text-muted-foreground">{data.project}</span>}
          </SheetTitle>
          <div className="text-xs text-muted-foreground mt-2 break-all">{data?.abs || path}</div>
        </div>
        <div className="flex-1 overflow-y-auto px-7 py-5">
          {error && <div className="text-destructive text-sm">{error}</div>}
          {!data && !error && <div className="text-muted-foreground text-sm">加载中…</div>}
          {data && (
            <pre className="text-[13px] leading-relaxed whitespace-pre-wrap break-words bg-muted/60 p-4 rounded-lg border border-border overflow-x-auto">
              <code
                className="hljs text-[13px] leading-relaxed bg-transparent p-0"
                dangerouslySetInnerHTML={{ __html: highlighted }}
              />
            </pre>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
