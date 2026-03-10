package muxy

import (
	"fmt"
	"net/http"
	"strings"
)

type edge struct {
	segment string
	node    *node
}

// node represents a single node in the routing tree
// ordered by priority: static > param > wildcard
type node struct {
	static   []edge
	param    *node
	wildcard *node

	paramKey string

	handlers        map[string]http.Handler // exact match
	subtreeHandlers map[string]http.Handler // /path/ subtree match
}

func (n *node) match(method string, segments []string, params map[string]string) http.Handler {
	current := n
	var subtree *node

	for _, seg := range segments {

		if current.subtreeHandlers != nil {
			subtree = current
		}

		for i := range current.static {
			if current.static[i].segment == seg {
				current = current.static[i].node
				goto matched
			}
		}

		if current.param != nil {
			params[current.param.paramKey] = seg
			current = current.param
			goto matched
		}

		if current.wildcard != nil {
			current = current.wildcard
			goto matched
		}

		goto fallback

	matched:
	}

	if current.handlers != nil {
		if h := current.handlers[method]; h != nil {
			return h
		}

		if h := current.handlers["*"]; h != nil {
			return h
		}
	}

fallback:
	if subtree != nil {
		if h := subtree.subtreeHandlers[method]; h != nil {
			return h
		}

		if h := subtree.subtreeHandlers["*"]; h != nil {
			return h
		}
	}

	return nil
}

func (n *node) insert(method, path string, handler http.Handler) {
	isSubtree := strings.HasSuffix(path, "/") && path != "/"

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
			var next *node

			for i := range current.static {
				if current.static[i].segment == seg {
					next = current.static[i].node
					break
				}
			}

			if next == nil {
				next = &node{}
				current.static = append(current.static, edge{
					segment: seg,
					node:    next,
				})
			}

			current = next
		}
	}

	if isSubtree {
		if current.subtreeHandlers == nil {
			current.subtreeHandlers = make(map[string]http.Handler)
		}

		if _, ok := current.subtreeHandlers[method]; ok {
			panic(fmt.Sprintf("muxy: subtree route already exists -> %s %s", method, path))
		}

		current.subtreeHandlers[method] = handler
		return
	}

	if current.handlers == nil {
		current.handlers = make(map[string]http.Handler)
	}

	if _, ok := current.handlers[method]; ok {
		panic(fmt.Sprintf("muxy: route already exists -> %s %s", method, path))
	}

	current.handlers[method] = handler
}
