export type ASTNode = { type: 'atom', value: string } | { type: 'list', children: ASTNode[], op?: string }
export function lispToAst(expr: string): ASTNode {
  const tokens = expr.replace(/\(/g,' ( ').replace(/\)/g,' ) ').trim().split(/\s+/).filter(Boolean)
  let idx=0
  function parse(): ASTNode {
    const tok = tokens[idx++]
    if(tok==='('){
      const children: ASTNode[]=[]
      while(idx < tokens.length && tokens[idx]!==')'){ children.push(parse()) }
      idx++
      return { type:'list', children, op: (children[0] as any)?.value }
    }
    return { type:'atom', value: tok }
  }
  try { return parse() } catch { return {type:'atom', value: expr} }
}
export function astToLisp(node: ASTNode): string {
  if(node.type==='atom') return node.value
  return '(' + node.children.map(astToLisp).join(' ') + ')'
}
export const BLOCKS = [
  { label:'Add', template:'(+ api_calls 10)' },
  { label:'Multiply', template:'(* api_calls 0.01)' },
  { label:'Tiered', template:'(tiered api_calls 0.01 1000 0.008 10000 0.005)' },
  { label:'Cap', template:'(max 5 (min 100 (* api_calls 0.01)))' },
  { label:'If discount', template:'(if (> api_calls 1000) (* api_calls 0.009) (* api_calls 0.01))' },
]
