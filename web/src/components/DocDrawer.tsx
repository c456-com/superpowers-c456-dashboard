import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeHighlight from 'rehype-highlight'
import 'highlight.js/styles/github.css'
import type { Document } from '../lib/types'
import { TYPE_LABEL } from '../lib/types'
import { Sheet, SheetContent, SheetTitle } from './ui/sheet'
import { TypeTag } from './views/OverviewView'
import FileViewer from './FileViewer'

interface Props {
  doc: Document | null
  onClose: () => void
  projectName?: string
}

export default function DocDrawer({ doc, onClose, projectName }: Props) {
  const [file, setFile] = useState<{ project: string; path: string; dir?: string } | null>(null)
  // 文档所在目录（markdown 相对链接基准）
  const docDir = doc ? (doc.path.includes('/') ? doc.path.slice(0, doc.path.lastIndexOf('/')) : '') : ''

  return (
    <>
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
                <article className="prose prose-sm prose-neutral dark:prose-invert max-w-none prose-headings:font-bold prose-headings:mt-6 prose-headings:mb-3 prose-p:my-2 prose-li:my-0.5 prose-pre:bg-muted prose-pre:p-3 prose-blockquote:border-l-2 prose-blockquote:border-border prose-blockquote:pl-3 prose-blockquote:text-muted-foreground prose-a:text-primary prose-table:border prose-table:border-border prose-th:bg-muted prose-th:p-2 prose-td:p-2">
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    rehypePlugins={[rehypeHighlight]}
                    components={{
                      pre: (props) => (
                        <pre
                          className="text-[13px] text-foreground leading-relaxed bg-muted/60 border border-border rounded-lg p-4 my-3 overflow-x-auto [&_code]:bg-transparent [&_code]:p-0 [&_code]:text-[13px] [&_code]:leading-relaxed [&_code]:text-foreground"
                          {...props}
                        />
                      ),
                      a: ({ href, children }) => {
                        // 相对路径链接（文档里的文件路径）→ 点击打开文件预览，不跳转
                        if (href && projectName && !/^(https?:|mailto:|#|\/|\?)/.test(href)) {
                          return (
                            <button
                              type="button"
                              onClick={(e) => {
                                e.preventDefault()
                                setFile({ project: projectName, path: href, dir: docDir })
                              }}
                              className="text-primary hover:underline cursor-pointer font-medium"
                            >
                              {children}
                            </button>
                          )
                        }
                        return <a href={href}>{children}</a>
                      },
                    }}
                  >
                    {doc.content}
                  </ReactMarkdown>
                </article>
              </div>
            </>
          )}
        </SheetContent>
      </Sheet>
      <FileViewer project={file?.project ?? null} path={file?.path ?? null} dir={file?.dir ?? null} onClose={() => setFile(null)} />
    </>
  )
}
