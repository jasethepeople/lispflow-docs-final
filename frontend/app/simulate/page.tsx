'use client'
import { useState } from 'react'
export default function Simulate(){
  const [plan, setPlan] = useState('(tiered api_calls 0.01 1000 0.008 10000 0.005)')
  const [scenarios, setScenarios] = useState('[{"api_calls":500},{"api_calls":1500}]')
  const [res, setRes] = useState<any>(null)
  function run(){
    const sc = JSON.parse(scenarios)
    const results = sc.map((s:any)=>{ const u=s.api_calls||0; let amt=u<=1000?u*0.01:90; return {amt} })
    setRes({total: results.reduce((a:any,b:any)=>a+b.amt,0), results})
  }
  return <div><h1>Simulate</h1><textarea value={plan} onChange={e=>setPlan(e.target.value)} style={{width:'100%', height:60, background:'#111', color:'#0f0'}}/><textarea value={scenarios} onChange={e=>setScenarios(e.target.value)} style={{width:'100%', height:60, background:'#111', color:'#fff', marginTop:8}}/><button onClick={run} style={{marginTop:8}}>Run</button>{res && <pre>{JSON.stringify(res,null,2)}</pre>}</div>
}
