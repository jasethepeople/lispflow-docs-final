package billing

import (
    "fmt"
    "strconv"
    "strings"
    "sync"
)

const MaxPlanSize = 16 * 1024

type Engine struct {
    mu sync.RWMutex
}

func NewEngine() *Engine { return &Engine{} }

func (e *Engine) Validate(expr string) error {
    if len(expr) > MaxPlanSize {
        return fmt.Errorf("plan exceeds max size %d", MaxPlanSize)
    }
    if strings.TrimSpace(expr) == "" {
        return fmt.Errorf("empty expression")
    }
    _, err := e.parse(expr)
    return err
}

func (e *Engine) Evaluate(expr string, metrics map[string]float64) (float64, map[string]float64, error) {
    if len(expr) > MaxPlanSize {
        return 0, nil, fmt.Errorf("plan too large")
    }
    node, err := e.parse(expr)
    if err != nil { return 0, nil, err }
    breakdown := make(map[string]float64)
    val, err := e.evalNode(node, metrics, breakdown)
    return val, breakdown, err
}

// AST node
type node struct {
    atom  string
    isList bool
    children []*node
}

func (e *Engine) parse(s string) (*node, error) {
    tokens := tokenize(s)
    if len(tokens)==0 { return nil, fmt.Errorf("empty") }
    idx := 0
    n, err := parseTokens(tokens, &idx)
    if err != nil { return nil, err }
    return n, nil
}

func tokenize(s string) []string {
    s = strings.ReplaceAll(s, "(", " ( ")
    s = strings.ReplaceAll(s, ")", " ) ")
    return strings.Fields(s)
}

func parseTokens(toks []string, idx *int) (*node, error) {
    if *idx >= len(toks) { return nil, fmt.Errorf("unexpected EOF") }
    tok := toks[*idx]
    *idx++
    if tok == "(" {
        n := &node{isList: true}
        for *idx < len(toks) && toks[*idx] != ")" {
            child, err := parseTokens(toks, idx)
            if err != nil { return nil, err }
            n.children = append(n.children, child)
        }
        if *idx >= len(toks) { return nil, fmt.Errorf("unclosed (") }
        *idx++ // consume )
        return n, nil
    }
    if tok == ")" { return nil, fmt.Errorf("unexpected )") }
    return &node{atom: tok}, nil
}

func (e *Engine) evalNode(n *node, metrics map[string]float64, breakdown map[string]float64) (float64, error) {
    if !n.isList {
        // atom: number or metric key (defaults to 0)
        if f, err := strconv.ParseFloat(n.atom, 64); err == nil {
            return f, nil
        }
        // metric lookup, default 0 per spec
        if v, ok := metrics[n.atom]; ok {
            breakdown[n.atom] = v
            return v, nil
        }
        // missing key defaults to 0
        return 0, nil
    }
    if len(n.children)==0 { return 0, fmt.Errorf("empty list") }
    opNode := n.children[0]
    if opNode.isList { return 0, fmt.Errorf("operator must be atom") }
    op := opNode.atom
    impl, ok := Builtins[op]
    if !ok { return 0, fmt.Errorf("unknown function %s", op) }
    args := make([]float64, 0, len(n.children)-1)
    for _, c := range n.children[1:] {
        v, err := e.evalNode(c, metrics, breakdown)
        if err != nil { return 0, err }
        args = append(args, v)
    }
    return impl(args, metrics)
}
