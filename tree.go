package muxy

import (
	"fmt"
	"net/http"
	"strings"
)

// node represents a single node in the tree ordered by priority (static > param > wildcard)
type node struct {
	static   map[string]*node
	param    *node
	wildcard *node

	paramKey string
	handlers map[string]http.Handler
}

func (n *node) match(method string, segments []string, params map[string]string) http.Handler {
	current := n

	for _, seg := range segments {
		if current.static != nil {
			if next, ok := current.static[seg]; ok {
				current = next
				continue
			}
		}

		if current.param != nil {
			params[current.param.paramKey] = seg
			current = current.param
			continue
		}

		if current.wildcard != nil {
			current = current.wildcard
			break
		}

		return nil
	}

	if current.handlers == nil {
		return nil
	}

	if h := current.handlers[method]; h != nil {
		return h
	}

	if h := current.handlers["*"]; h != nil {
		return h
	}

	return nil
}

func (n *node) insert(method, path string, handler http.Handler) {
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

	if current.handlers == nil {
		current.handlers = make(map[string]http.Handler)
	}

	if _, ok := current.handlers[method]; ok {
		panic(fmt.Sprintf("muxy: route already exists -> %s %s", method, path))
	}

	current.handlers[method] = handler
}
