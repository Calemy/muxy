package muxy

import (
	"net/http"
	"strings"
)

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

func (n *node) insert(path string, handler http.Handler) {
	segments := split(path)

	current := n

	for _, seg := range segments {

		switch {
		case strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}"):
			key := seg[1 : len(seg)-1]

			if current.param == nil {
				current.param = &node{paramKey: key}
			}

			current = current.param

		case seg == "*":
			if current.wildcard == nil {
				current.wildcard = &node{}
			}

			current = current.wildcard

		default:
			if current.static == nil {
				current.static = make(map[string]*node)
			}

			if current.static[seg] == nil {
				current.static[seg] = &node{}
			}

			current = current.static[seg]
		}
	}

	current.handler = handler
}
