package muxy

import "net/http"

// node represents a single node in the tree ordered by priority (static > param > wildcard)
type node struct {
	static   map[string]*node
	param    *node
	wildcard *node
	handler  http.Handler
	paramKey string
}

func (n *node) match(segments []string, params map[string]string) http.Handler {
	if len(segments) == 0 {
		return n.handler
	}

	seg := segments[0]

	if next, ok := n.static[seg]; ok {
		if h := next.match(segments[1:], params); h != nil {
			return h
		}
	}

	if n.param != nil {
		params[n.param.paramKey] = seg
		if h := n.param.match(segments[1:], params); h != nil {
			return h
		}
		delete(params, n.param.paramKey)
	}

	if n.wildcard != nil {
		return n.wildcard.handler
	}

	return nil
}
