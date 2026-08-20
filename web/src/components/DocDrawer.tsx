import type { Document } from '../lib/types'
import { TYPE_LABEL } from '../lib/types'
import { Sheet, SheetContent, SheetTitle } from './ui/sheet'
import { TypeTag } from './views/OverviewView'

interface Props {
  doc: Document | null
  onClose: () => void
}

export default function DocDrawer({ doc, onClose }: Props) {
  return (
    <Sheet open={!!doc} onOpenChange={(open) => !open && onClose()}>
      <SheetContent className="w-full sm:max-w-3xl! overflow-y-auto px-7" side="right">
        {doc && (
          <>
            <SheetTitle className="text-base font-bold flex items-center gap-2 pr-8">
              <TypeTag type={doc.type} />
              {doc.title}
            </SheetTitle>
            <div className="text-xs text-muted-foreground mt-2 flex items-center gap-2 flex-wrap">
              <span>{TYPE_LABEL[doc.type]}</span>
              <span className="break-all">{doc.path}</span>
              <span>{doc.date || ''}</span>
              {doc.status ? ` · ${doc.status}` : ''}
            </div>
            <div
              className="markdown mt-4 px-1"
              // 简单 markdown 渲染（完整渲染可选 marked/渲染库）
            >
              <RenderMd content={doc.content} />
            </div>
          </>
        )}
      </SheetContent>
    </Sheet>
  )
}

// 轻量 markdown→HTML（覆盖标题/粗体/列表/代码，够用）
function RenderMd({ content }: { content: string }) {
  const lines = content.split('\n')
  let html = ''
  for (const ln of lines) {
    const s = ln
    if (/^#{1,6}\s/.test(s)) {
      const level = s.match(/^#+/)![0].length
      html += `<h${Math.min(level, 6)}>${esc(s.replace(/^#+\s/, ''))}</h${Math.min(level, 6)}>\n`
    } else if (/^```/.test(s)) {
      html += '<pre>'
    } else if (/^```$/.test(s) && /<pre>$/.test(html)) {
      html += '</pre>\n'
    } else if (/^[-*]\s/.test(s)) {
      html += `<li>${esc(s.replace(/^[-*]\s/, ''))}</li>\n`
    } else if (/^\d+\.\s/.test(s)) {
      html += `<li>${esc(s.replace(/^\d+\.\s/, ''))}</li>\n`
    } else if (s.trim() === '') {
      html += '\n'
    } else if (/<\/?pre>/.test(html)) {
      html += esc(s) + '\n'
    } else {
      html += `<p>${esc(s)}</p>\n`
    }
  }
  // 简单包裹列表
  html = html.replace(/(<li>[^]*?)(?=<p>|<li>|$)/g, '<ul>$&</ul>').replace(/<\/li>\s*<\/ul>/g, '</li></ul>')
  return <div dangerouslySetInnerHTML={{ __html: html }} />
}

function esc(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}
