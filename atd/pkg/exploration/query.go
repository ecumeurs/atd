package exploration

import (
	"atd-tools/pkg/atom"
	"strings"
)

func (e *Explorer) Query(field, search string) []*atom.AtomData {
	var matches []*atom.AtomData
	search = strings.ToLower(search)

	for _, a := range e.Graph.Atoms {
		match := false
		if field == "" {
			if strings.Contains(strings.ToLower(a.ID), search) ||
				strings.Contains(strings.ToLower(a.HumanName), search) {
				match = true
			} else {
				for _, t := range a.Tags {
					if strings.Contains(strings.ToLower(t), search) {
						match = true
						break
					}
				}
			}
		} else {
			var val string
			switch strings.ToLower(field) {
			case "id":
				val = a.ID
			case "human_name", "name":
				val = a.HumanName
			case "status":
				val = a.Status
			case "layer":
				val = a.Layer
			case "type":
				val = a.Type
			case "tags":
				val = strings.Join(a.Tags, ",")
			}
			if strings.Contains(strings.ToLower(val), search) {
				match = true
			}
		}

		if match {
			matches = append(matches, a)
		}
	}
	return matches
}