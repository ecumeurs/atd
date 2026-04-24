package exploration

import (
	"atd-tools/pkg/atom"
	"fmt"
	"os/exec"
	"strings"
)

type HeatState string

const (
	HeatCold    HeatState = "cold"
	HeatOptimal HeatState = "optimal"
	HeatWarm    HeatState = "warm"
	HeatHot     HeatState = "hot"
	HeatStable  HeatState = "stable" // Used for Update heat when 0 updates
)

type HeatMetrics struct {
	Parents          int    `json:"parents"`
	Dependents       int    `json:"dependents"`
	CodeFilesLinked  int    `json:"codeFilesLinked"`
	RecentUpdates    int    `json:"recentUpdates"`
	LastUpdated      string `json:"lastUpdated"`
}

type HeatMapResult struct {
	Atom            string      `json:"atom"`
	Layer           string      `json:"layer"`
	DependencyState HeatState   `json:"dependency_state"`
	CodeState       HeatState   `json:"code_state"`
	UpdateState     HeatState   `json:"update_state"`
	Metrics         HeatMetrics `json:"metrics"`
	Recommendations []string    `json:"recommendations"`
}

// CalculateDependencyHeat returns the heat state based on parent/dependent counts.
func (e *Explorer) CalculateDependencyHeat(a *atom.AtomData) HeatState {
	p := len(a.Parents)
	d := len(a.Dependents)

	// Special Rules:
	// IMPLEMENTATION layer: May have 0 dependents (leaf nodes acceptable)
	// CUSTOMER layer: May have 0 parents (top-level requirements acceptable)
	isLeafOk := a.Layer == "IMPLEMENTATION"
	isRootOk := a.Layer == "CUSTOMER"

	if p == 0 && d == 0 {
		return HeatCold
	}

	if (!isRootOk && p == 0) || (!isLeafOk && d == 0) {
		// Isolated but not 0/0, still feels "cold" or "isolated"
		// but let's stick to the issue table:
		// Too Cold: 0 / 0
	}

	max := p
	if d > max {
		max = d
	}

	if max >= 10 {
		return HeatHot
	}
	if max >= 4 {
		return HeatWarm
	}
	if max >= 1 {
		return HeatOptimal
	}

	return HeatCold
}

// GetCodeFileHeat returns the heat state for a single code file based on linked atom count.
func (e *Explorer) GetCodeFileHeat(path string) HeatState {
	count := 0
	for _, sl := range e.SpecLinks {
		if sl.FilePath == path {
			count++
		}
	}

	if count >= 10 {
		return HeatHot
	}
	if count >= 6 {
		return HeatWarm
	}
	if count >= 2 {
		return HeatOptimal
	}
	return HeatCold
}

// CalculateCodeHeat returns the heat state for an atom based on the heat of linked code files.
// Atom takes the maximum heat of all its linked code files.
func (e *Explorer) CalculateCodeHeat(a *atom.AtomData) HeatState {
	if len(a.Implementations) == 0 {
		return HeatCold
	}

	maxHeat := HeatCold
	heatRank := map[HeatState]int{
		HeatCold:    0,
		HeatOptimal: 1,
		HeatWarm:    2,
		HeatHot:     3,
	}

	for _, impl := range a.Implementations {
		// impl is "path:line"
		parts := strings.Split(impl, ":")
		path := parts[0]
		heat := e.GetCodeFileHeat(path)
		if heatRank[heat] > heatRank[maxHeat] {
			maxHeat = heat
		}
	}

	return maxHeat
}

// CalculateUpdateHeat returns the heat state based on git history.
func (e *Explorer) CalculateUpdateHeat(a *atom.AtomData) (HeatState, int, string) {
	// git log -n 20 --pretty=format:"%as" -- <file>
	// %as: author date, short (YYYY-MM-DD)
	cmd := exec.Command("git", "log", "-n", "20", "--pretty=format:%as", "--", a.FilePath)
	cmd.Dir = e.ProjectRoot
	out, err := cmd.Output()
	if err != nil {
		return HeatStable, 0, ""
	}

	dates := strings.Split(strings.TrimSpace(string(out)), "\n")
	count := 0
	lastUpdated := ""
	if len(dates) > 0 && dates[0] != "" {
		lastUpdated = dates[0]
		count = len(dates)
	}

	if count == 0 {
		return HeatStable, 0, ""
	}
	if count >= 7 {
		return HeatHot, count, lastUpdated
	}
	if count >= 4 {
		return HeatWarm, count, lastUpdated
	}
	if count >= 1 {
		return HeatOptimal, count, lastUpdated
	}

	return HeatStable, 0, ""
}

// GetHeatMapResult returns a full heat map result for a single atom.
func (e *Explorer) GetHeatMapResult(atomID string) (*HeatMapResult, error) {
	a, ok := e.Graph.Atoms[atomID]
	if !ok {
		return nil, fmt.Errorf("atom not found: %s", atomID)
	}

	updateState, updateCount, lastUpdated := e.CalculateUpdateHeat(a)

	res := &HeatMapResult{
		Atom:            a.ID,
		Layer:           a.Layer,
		DependencyState: e.CalculateDependencyHeat(a),
		CodeState:       e.CalculateCodeHeat(a),
		UpdateState:     updateState,
		Metrics: HeatMetrics{
			Parents:         len(a.Parents),
			Dependents:      len(a.Dependents),
			CodeFilesLinked: len(a.Implementations),
			RecentUpdates:   updateCount,
			LastUpdated:     lastUpdated,
		},
		Recommendations: []string{},
	}

	// Add recommendations based on heat
	if res.DependencyState == HeatHot {
		res.Recommendations = append(res.Recommendations, "Excessive coupling detected. Consider splitting this atom.")
	} else if res.DependencyState == HeatCold {
		if a.Layer != "CUSTOMER" && a.Layer != "IMPLEMENTATION" {
			res.Recommendations = append(res.Recommendations, "This atom is isolated. Ensure it is part of the dependency graph.")
		}
	}

	if res.CodeState == HeatHot {
		res.Recommendations = append(res.Recommendations, "Linked code files are overloaded. Consider decomposing the implementation.")
	}

	if res.UpdateState == HeatHot {
		res.Recommendations = append(res.Recommendations, "Frequent updates detected. This atom may be unstable.")
	}

	return res, nil
}
