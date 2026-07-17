package exploration

import (
	"atd-tools/pkg/atom"
	"atd-tools/pkg/workspace"
	"errors"
	"fmt"
	"strings"
)

func (e *Explorer) CanonicalAtomID(refID string) (string, error) {
	if e.Resolver != nil {
		parsed, err := e.Resolver.Resolve(refID)
		if err != nil {
			return "", err
		}
		if parsed.Location != nil {
			if parsed.Type == workspace.ReferenceCrossProject {
				return fmt.Sprintf("%s:%s", parsed.Project, parsed.AtomID), nil
			}
			return parsed.AtomID, nil
		}
		return "", fmt.Errorf("atom '%s' not found", refID)
	}

	if _, ok := e.Graph.Atoms[refID]; ok {
		return refID, nil
	}

	// Retry once with a leading known type/layer token stripped (e.g.
	// "requirement_req_x" -> "req_x"), the same strip-and-retry
	// workspace.Resolver already does for workspace-aware projects -- a
	// standalone project (no e.Resolver) deserves the same forgiveness
	// instead of failing a redundantly-prefixed id exactly like a genuinely
	// nonexistent one.
	if stripped, hasPrefix := workspace.StripKnownPrefix(refID); hasPrefix {
		if _, ok := e.Graph.Atoms[stripped]; ok {
			return stripped, nil
		}
	}

	if suggestion := e.SuggestAtomID(refID); suggestion != "" {
		return "", fmt.Errorf("atom '%s' not found (did you mean '%s'?)", refID, suggestion)
	}
	return "", fmt.Errorf("atom '%s' not found", refID)
}

func (e *Explorer) SuggestAtomID(bareID string) string {
	if idx := strings.LastIndex(bareID, ":"); idx >= 0 {
		bareID = bareID[idx+1:]
	}

	if e.Index != nil {
		if loc := e.Index.FindAtom(bareID); loc != nil {
			if loc.Project != "" {
				return fmt.Sprintf("%s:%s", loc.Project, bareID)
			}
			return bareID
		}
	}

	for id := range e.Graph.Atoms {
		if idx := strings.LastIndex(id, ":"); idx >= 0 && id[idx+1:] == bareID {
			return id
		}
	}

	return ""
}

func (e *Explorer) ResolveAtom(refID string) (*atom.AtomData, error) {
	if e.Resolver != nil {
		parsed, err := e.Resolver.Resolve(refID)
		if err == workspace.ErrUnknownProject || errors.Is(err, workspace.ErrAmbiguousAtom) {
			return nil, err
		}
		if err == nil && parsed.Location != nil {
			id := parsed.AtomID
			if parsed.Type == workspace.ReferenceCrossProject {
				id = fmt.Sprintf("%s:%s", parsed.Project, parsed.AtomID)
			}
			if node, ok := e.Graph.Atoms[id]; ok {
				return node, nil
			}

			a, err := atom.Parse(parsed.Location.Path)
			if err != nil {
				return nil, err
			}
			node := a
			if e.Graph != nil {
				e.Graph.Atoms[id] = &node
			}
			return &node, nil
		}
	}

	if node, ok := e.Graph.Atoms[refID]; ok {
		return node, nil
	}

	// Retry once with a leading known type/layer token stripped -- see the
	// matching comment in CanonicalAtomID above; this is the same
	// generalization applied to ResolveAtom's standalone path.
	if stripped, hasPrefix := workspace.StripKnownPrefix(refID); hasPrefix {
		if node, ok := e.Graph.Atoms[stripped]; ok {
			return node, nil
		}
	}

	if suggestion := e.SuggestAtomID(refID); suggestion != "" {
		return nil, fmt.Errorf("atom '%s' not found (did you mean '%s'?)", refID, suggestion)
	}
	return nil, fmt.Errorf("atom '%s' not found", refID)
}