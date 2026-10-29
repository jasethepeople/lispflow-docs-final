package billing

import (
    "fmt"
    "math"
)

type FuncImpl func(args []float64, metrics map[string]float64) (float64, error)

var Builtins map[string]FuncImpl

func init() {
    Builtins = map[string]FuncImpl{
        "+": func(args []float64, _ map[string]float64) (float64, error) {
            sum := 0.0
            for _, a := range args { sum += a }
            return sum, nil
        },
        "-": func(args []float64, _ map[string]float64) (float64, error) {
            if len(args) == 0 { return 0, fmt.Errorf("- requires args") }
            if len(args) == 1 { return -args[0], nil }
            res := args[0]
            for _, a := range args[1:] { res -= a }
            return res, nil
        },
        "*": func(args []float64, _ map[string]float64) (float64, error) {
            prod := 1.0
            for _, a := range args { prod *= a }
            return prod, nil
        },
        "/": func(args []float64, _ map[string]float64) (float64, error) {
            if len(args) == 0 { return 0, fmt.Errorf("/ requires args") }
            res := args[0]
            for _, a := range args[1:] {
                if a == 0 { return 0, fmt.Errorf("division by zero") }
                res /= a
            }
            return res, nil
        },
        "min": func(args []float64, _ map[string]float64) (float64, error) {
            if len(args) == 0 { return 0, fmt.Errorf("min requires args") }
            m := args[0]
            for _, a := range args[1:] { if a < m { m = a } }
            return m, nil
        },
        "max": func(args []float64, _ map[string]float64) (float64, error) {
            if len(args) == 0 { return 0, fmt.Errorf("max requires args") }
            m := args[0]
            for _, a := range args[1:] { if a > m { m = a } }
            return m, nil
        },
        "tiered": func(args []float64, _ map[string]float64) (float64, error) {
            // tiered usage price1 limit1 price2 limit2 ... e.g. (tiered api_calls 0.01 1000 0.008 10000 0.005)
            if len(args) < 3 || len(args)%2 == 0 { return 0, fmt.Errorf("tiered: (tiered usage price1 limit1 price2 limit2 ... final_price)") }
            usage := args[0]
            remaining := usage
            total := 0.0
            i := 1
            for i+1 < len(args)-1 {
                price := args[i]
                limit := args[i+1]
                tierSize := math.Min(remaining, limit)
                if tierSize < 0 { tierSize = 0 }
                total += tierSize * price
                remaining -= tierSize
                if remaining <= 0 { return total, nil }
                i += 2
            }
            finalPrice := args[len(args)-1]
            total += remaining * finalPrice
            return total, nil
        },
        "if": func(args []float64, _ map[string]float64) (float64, error) {
            if len(args) != 3 { return 0, fmt.Errorf("if requires cond then else") }
            if args[0] != 0 { return args[1], nil }
            return args[2], nil
        },
        ">": func(args []float64, _ map[string]float64) (float64, error) {
            if len(args) != 2 { return 0, fmt.Errorf("> requires 2 args") }
            if args[0] > args[1] { return 1, nil }
            return 0, nil
        },
        "<": func(args []float64, _ map[string]float64) (float64, error) {
            if len(args) != 2 { return 0, fmt.Errorf("< requires 2 args") }
            if args[0] < args[1] { return 1, nil }
            return 0, nil
        },
        ">=": func(args []float64, _ map[string]float64) (float64, error) {
            if len(args) != 2 { return 0, fmt.Errorf(">= requires 2 args") }
            if args[0] >= args[1] { return 1, nil }
            return 0, nil
        },
        "<=": func(args []float64, _ map[string]float64) (float64, error) {
            if len(args) != 2 { return 0, fmt.Errorf("<= requires 2 args") }
            if args[0] <= args[1] { return 1, nil }
            return 0, nil
        },
        "=": func(args []float64, _ map[string]float64) (float64, error) {
            if len(args) != 2 { return 0, fmt.Errorf("= requires 2 args") }
            if args[0] == args[1] { return 1, nil }
            return 0, nil
        },
        "volume": func(args []float64, _ map[string]float64) (float64, error) {
            // (volume metric tiers...) tiered but volume pricing (all units at tier price)
            if len(args) < 2 { return 0, fmt.Errorf("volume requires usage and tiers") }
            usage := args[0]
            // args[1:] are pairs: limit price? We'll interpret as limit price
            i := 1
            for i+1 < len(args) {
                limit := args[i]
                price := args[i+1]
                if usage <= limit { return usage * price, nil }
                i += 2
            }
            if len(args)%2 == 0 { // final price is last arg
                return usage * args[len(args)-1], nil
            }
            return usage * args[len(args)-1], nil
        },
    }
}
