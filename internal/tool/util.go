package tool

func getInt(args map[string]any, key string, def int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64: return int(n)
		case int: return n
		case int64: return int(n)
		}
	}
	return def
}
