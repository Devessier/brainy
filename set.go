package brainy

type stateNodesSet struct {
	stateNodes []*StateNode
}

func (set *stateNodesSet) Has(s *StateNode) bool {
	for _, node := range set.stateNodes {
		if node == s {
			return true
		}
	}

	return false
}

func (set *stateNodesSet) Add(nodes ...*StateNode) {
	for _, nodeToAdd := range nodes {
		if set.Has(nodeToAdd) {
			continue
		}

		set.stateNodes = append(set.stateNodes, nodeToAdd)
	}
}

func (set *stateNodesSet) ToSlice() []*StateNode {
	stateNodesCopy := make([]*StateNode, 0, len(set.stateNodes))

	copy(stateNodesCopy, set.stateNodes)

	return stateNodesCopy
}
