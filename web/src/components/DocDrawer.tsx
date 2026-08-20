import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
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
      <SheetContent className="w-full sm:max-w-3xl! flex flex-col p-0" side="right">
        {doc && (
          <>
            {/* toolbar：独立 padding，固定顶部 */}
            <div className="shrink-0 px-6 py-4 border-b border-border">
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
            </div>
            {/* content：独立滚动 + 上下左右 padding；react-markdown 完整渲染 GFM */}
            <div className="flex-1 overflow-y-auto px-7 py-5">
              <article className="prose prose-sm prose-neutral dark:prose-invert max-w-none prose-headings:font-bold prose-headings:mt-6 prose-headings:mb-3 prose-p:my-2 prose-li:my-0.5 prose-pre:bg-gray-50 prose-pre:p-3 prose-blockquote:border-l-2 prose-blockquote:border-gray-300 prose-blockquote:pl-3 prose-blockquote:text-gray-600 prose-a:text-blue-600 prose-table:border prose-table:border-gray-200 prose-th:bg-gray-50 prose-th:p-2 prose-td:p-2">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>{doc.content}</ReactMarkdown>
              </article>
            </div>
          </>
        )}
      </SheetContent>
    </Sheet>
  )
}
