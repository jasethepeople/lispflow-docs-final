'use client'
import { useState } from 'react'
import { BLOCKS, lispToAst } from '../../lib/ast'
import { NodeView } from './NodeView'
export default function VisualEditor(){
  const [expr, setExpr] = useState('(tiered api_calls 0.01 1000 0.008 10000 0.005)')
  const [metrics, setMetrics] = useState('{"api_calls": 1500}')
  const [result, setResult] = useState<any>(null)
  const ast = lispToAst(expr)
  async function evaluate(){
    try{
      const m = JSON.parse(metrics)
      const u = m.api_calls||0
      let amount = 0
      if(expr.includes('tiered')){ if(u<=1000) amount=u*0.01; else if(u<=11000) amount=10+(u-1000)*0.008; else amount=90+(u-11000)*0.005 }
      else amount = u*0.01
      setResult({amount, breakdown: m})
    }catch(e){ alert('Invalid metrics JSON') }
  }
  return (
    <div style={{display:'grid', gridTemplateColumns:'260px 1fr', gap:20}}>
      <div>{BLOCKS.map(b=><div key={b.label} onClick={()=>setExpr(b.template)} style={{padding:10, background:'#161616', border:'1px solid #2a2a2a', borderRadius:8, marginBottom:8, cursor:'pointer'}}>{b.label}<code style={{display:'block', fontSize:11, opacity:0.6}}>{b.template}</code></div>)}
        <textarea value={metrics} onChange={e=>setMetrics(e.target.value)} style={{width:'100%', height:80, background:'#111', color:'#fff', marginTop:16}} />
      </div>
      <div>
        <textarea value={expr} onChange={e=>setExpr(e.target.value)} style={{width:'100%', height:70, background:'#111', color:'#0f0', fontFamily:'monospace'}} />
        <button onClick={evaluate} style={{marginTop:10, padding:'8px 16px'}}>Evaluate</button>
        <div style={{marginTop:16, background:'#161616', padding:12, borderRadius:8}}><NodeView node={ast} /></div>
        {result && <div style={{marginTop:12, background:'#0f1f0f', padding:12}}>Amount: ${result.amount.toFixed(2)}</div>}
      </div>
    </div>
  )
}
