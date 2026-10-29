'use client'
export default function Dashboard(){
  return <div><h1>Dashboard</h1><div style={{display:'grid', gridTemplateColumns:'repeat(3,1fr)', gap:12}}><div style={{background:'#161616', padding:16, borderRadius:8}}>Total Evals: 12,482</div><div style={{background:'#161616', padding:16, borderRadius:8}}>Active Plans: 3</div><div style={{background:'#161616', padding:16, borderRadius:8}}>Avg Revenue: $42.50</div></div></div>
}
