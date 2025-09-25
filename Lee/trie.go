package Lee

import (
	"strings"
)

type node struct {
	pattern  string  // 待匹配路由，例如 /p/:lang
	part     string  // 路由中的一部分，例如 :lang
	children []*node // 子节点，例如 [doc, tutorial, intro]
	isWild   bool    // 是否精确匹配，part 含有 : 或 * 时为true

	// 性能优化：缓存子节点查找
	childrenMap map[string]*node // 静态路由快速查找
}

// 第一个匹配成功的节点，用于插入
func (n *node) matchChild(part string) *node {
	// 优先从map中查找静态路由
	if n.childrenMap != nil {
		if child := n.childrenMap[part]; child != nil {
			return child
		}
	}
	
	// 查找通配符路由
	for _, child := range n.children {
		if child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0, 2) // 预分配容量
	
	// 优先查找精确匹配
	if n.childrenMap != nil {
		if child := n.childrenMap[part]; child != nil {
			nodes = append(nodes, child)
		}
	}
	
	// 查找通配符匹配
	for _, child := range n.children {
		if child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
		
		// 为静态路由建立快速查找映射
		if !child.isWild {
			if n.childrenMap == nil {
				n.childrenMap = make(map[string]*node)
			}
			n.childrenMap[part] = child
		}
	}
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}
