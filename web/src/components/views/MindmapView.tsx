import { useMemo } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  Handle,
  Position,
  type Node,
  type Edge,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import dagre from 'dagre'
import type { Document, Project } from '../../lib/types'
import { TypeTag } from './OverviewView'

interface Props {
  project: Project
  onOpen: (d: Document) => void
}

const NODE_W = 240
const NODE_H = 90

// 自定义节点类型（模块级常量，避免 React Flow 每次渲染重建类型引用）
const NODE_TYPES = { root: RootNode, spec: SpecNode }

// dagre 自动分层布局（左右树形）
function layout(nodes: Node[], edges: Edge[]) {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'LR', nodesep: 40, ranksep: 120 })
  nodes.forEach((n) => g.setNode(n.id, { width: NODE_W, height: NODE_H }))
  edges.forEach((e) => g.setEdge(e.source, e.target))
  dagre.layout(g)
  return nodes.map((n) => {
    const pos = g.node(n.id)
    return { ...n, position: { x: pos.x - NODE_W / 2, y: pos.y - NODE_H / 2 } }
  })
}

export default function MindmapView({ project, onOpen }: Props) {
  const specs = project.documents.filter((d) => d.type === 'spec')

  const { nodes, edges } = useMemo(() => {
    if (specs.length === 0) return { nodes: [], edges: [] }
    const rootId = 'root'
    const rawNodes: Node[] = [
      {
        id: rootId,
        type: 'root',
        position: { x: 0, y: 0 },
        data: { label: `${project.name} · 功能设计 (${specs.length})` },
      },
      ...specs.map((d) => ({
        id: d.path,
        type: 'spec',
        position: { x: 0, y: 0 },
        data: { doc: d, onOpen },
      })),
    ]
    const rawEdges: Edge[] = specs.map((d) => ({
      id: 'e-' + d.path,
      source: rootId,
      target: d.path,
      type: 'smoothstep',
    }))
    return { nodes: layout(rawNodes, rawEdges), edges: rawEdges }
  }, [project, specs])

  if (specs.length === 0) {
    return (
      <div className="p-6 text-center text-muted-foreground text-sm">
        暂无功能设计文档
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 px-6 py-3">
        <h2 className="text-lg font-bold">🕸️ 功能设计图谱</h2>
        <span className="text-xs text-muted-foreground">
          spec 图谱（可拖拽/缩放）· 点击功能节点查看详情
        </span>
      </div>
      <div className="flex-1 bg-card border border-border rounded-xl overflow-hidden">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          fitView
          proOptions={{ hideAttribution: true }}
          nodeTypes={NODE_TYPES}
          nodesDraggable
          minZoom={0.2}
        >
          <Background gap={16} />
          <Controls />
        </ReactFlow>
      </div>
    </div>
  )
}

// 根节点
function RootNode({ data }: { data: { label: string } }) {
  return (
    <div className="bg-primary text-white rounded-xl px-4 py-3 font-semibold text-sm shadow text-center">
      {data.label}
      <Handle type="source" position={Position.Right} style={{ opacity: 0 }} />
    </div>
  )
}

// 功能 spec 节点（点开详情）
function SpecNode({ data }: { data: { doc: Document; onOpen: (d: Document) => void } }) {
  const d = data.doc
  return (
    <button
      type="button"
      onClick={() => data.onOpen(d)}
      className="block w-60 text-left bg-card border-2 border-blue-200 rounded-xl p-3 shadow-sm hover:border-blue-400 hover:shadow transition-all cursor-pointer"
    >
      <Handle type="target" position={Position.Left} style={{ opacity: 0 }} />
      <div className="flex items-center gap-1.5 mb-1.5">
        <TypeTag type="spec" />
      </div>
      <div className="text-[13px] font-semibold leading-snug line-clamp-2" title={d.title}>
        {d.title}
      </div>
      <div className="mt-1.5 flex flex-col gap-0.5">
        {(d.sections || []).slice(0, 4).map((s, j) => (
          <div key={j} className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
            <span className="size-1 rounded-full bg-blue-300 shrink-0" />
            <span className="truncate">{s.title}</span>
          </div>
        ))}
      </div>
      <div className="text-[10px] text-muted-foreground mt-1">{d.date || ''}</div>
      <Handle type="source" position={Position.Right} style={{ opacity: 0 }} />
    </button>
  )
}
