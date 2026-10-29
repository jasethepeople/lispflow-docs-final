'use client'
import { ASTNode } from '../../lib/ast'
export function NodeView({node, depth=0}:{node:ASTNode, depth?:number}){
  if(node.type==='atom'){
    return <span style={{background: '#222', padding:'2px 6px', borderRadius:4, margin:2, display:'inline-block', fontFamily:'monospace'}}>{node.value}</span>
  }
  return <div style={{marginLeft: depth?16:0, paddingLeft: depth?8:0, marginTop:4}}><div style={{display:'flex', flexWrap:'wrap', gap:4, background:'#1a1a1a', padding:6, borderRadius:6, border:'1px solid #2a2a2a'}}>{node.children.map((c,i)=><NodeView key={i} node={c} depth={depth+1} />)}</div></div>
}
