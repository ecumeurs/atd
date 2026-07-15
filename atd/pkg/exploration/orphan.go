package exploration

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
)

func (e *Explorer) IsOrphan(node *atom.AtomData) bool {
	if node.Status != "STABLE" {
		return false
	}

	if len(node.Implementations) > 0 {
		return false
	}

	if config.GetOrphanExcludedTypes()[node.Type] {
		return false
	}

	if e.Config.BusinessLayerException && node.Layer == "BUSINESS" {
		return false
	}

	if e.Config.HierarchicalOrphanCheck {
		for _, depID := range node.Dependents {
			if depNode, ok := e.Graph.Atoms[depID]; ok {
				if len(depNode.Implementations) > 0 {
					return false
				}
			}
		}
	}

	return true
}