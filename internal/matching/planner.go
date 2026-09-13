// Package matching computes private candidate plans from bounded temporary input.
// It performs no network I/O and exposes no public counts or participant records.
package matching

import (
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/geography"
	"sort"
)

type Signal struct {
	Hash                 string
	Area                 geography.ParticipantArea
	CreatedAt, ExpiresAt int64
	// Assigned signals and live reservations cannot found a competing gathering.
	Assigned      bool
	ReservedUntil int64
	Declined      map[string]bool
}
type Gathering struct {
	ID       string
	Crossing geography.Crossing
	EndsAt   int64
}
type Proposal struct {
	Crossing geography.Crossing
	Founders []Signal
}
type Planner struct {
	Index  *geography.Index
	Config config.Matching
}

// Offer prefers the closest reachable open gathering, with stable ID tie-breaking.
// It never modifies willingness or attendance. Final admission must revalidate.
func (p Planner) Offer(now int64, s Signal, open []Gathering) (Gathering, bool) {
	best := Gathering{}
	distance := 0.0
	minimum := now + int64(p.Config.LateJoinMinRemainingMinutes)*60000
	if s.ExpiresAt < minimum {
		return best, false
	}
	for _, g := range open {
		if g.EndsAt < minimum || s.Declined[g.ID] || !p.Index.CanReach(s.Area, g.Crossing.ID) {
			continue
		}
		d := geography.Distance(p.Index.Grid.Center(s.Area.Cell), g.Crossing.Point)
		if best.ID == "" || d < distance || (d == distance && g.ID < best.ID) {
			best = g
			distance = d
		}
	}
	return best, best.ID != ""
}

// Propose reserves exactly the activation threshold's oldest compatible signals.
// Other compatible people can receive the activated gathering through Offer. This
// bounds an atomic reservation/activation; a large crowd need not be written in a
// single transaction. Counts rank candidate crossings but are not public output.
func (p Planner) Propose(now int64, signals []Signal, open []Gathering) (Proposal, bool) {
	var none Proposal
	eligible := make([]Signal, 0, len(signals))
	minimum := now + int64(p.Config.MinimumRemainingMinutes)*60000
	for _, s := range signals {
		if s.Assigned || s.ReservedUntil > now || s.ExpiresAt < minimum {
			continue
		}
		if _, ok := p.Index.Grid.Parse(p.Index.Grid.ID(s.Area.Cell)); ok != nil {
			continue
		}
		if p.Config.PreferOpenGatherings {
			if _, ok := p.Offer(now, s, open); ok {
				continue
			}
		}
		eligible = append(eligible, s)
	}
	if len(eligible) < p.Config.ActivationCount {
		return none, false
	}
	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].CreatedAt != eligible[j].CreatedAt {
			return eligible[i].CreatedAt < eligible[j].CreatedAt
		}
		return eligible[i].Hash < eligible[j].Hash
	})
	type bucket struct {
		count  int
		oldest int64
	}
	areas := map[geography.ParticipantArea]bucket{}
	for _, s := range eligible {
		b := areas[s.Area]
		if b.count == 0 {
			b.oldest = s.CreatedAt
		}
		b.count++
		areas[s.Area] = b
	}
	stats := make([]bucket, len(p.Index.Crossings))
	for area, b := range areas {
		for _, n := range p.Index.Reachable(area) {
			v := stats[n]
			if v.count == 0 || b.oldest < v.oldest {
				v.oldest = b.oldest
			}
			v.count += b.count
			stats[n] = v
		}
	}
	candidates := []int{}
	for n, b := range stats {
		if b.count >= p.Config.ActivationCount {
			candidates = append(candidates, n)
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := stats[candidates[i]], stats[candidates[j]]
		if a.count != b.count {
			return a.count > b.count
		}
		if a.oldest != b.oldest {
			return a.oldest < b.oldest
		}
		return p.Index.Crossings[candidates[i]].ID < p.Index.Crossings[candidates[j]].ID
	})
	if len(candidates) == 0 {
		return none, false
	}
	seed := p.Index.Crossings[candidates[0]]
	group := make([]Signal, 0, p.Config.ActivationCount)
	groupAreas := make([]geography.ParticipantArea, 0, p.Config.ActivationCount)
	reachable := map[geography.ParticipantArea]bool{}
	for area := range areas {
		reachable[area] = p.Index.CanReach(area, seed.ID)
	}
	for _, s := range eligible {
		if reachable[s.Area] {
			group = append(group, s)
			groupAreas = append(groupAreas, s.Area)
			if len(group) == p.Config.ActivationCount {
				break
			}
		}
	}
	crossing, ok := p.Index.Closest(groupAreas)
	if !ok {
		return none, false
	}
	return Proposal{crossing, group}, true
}

// Deadline is frozen at activation. Later admission cannot call this to renew it.
func (p Planner) Deadline(now int64, founders []Signal) int64 {
	deadline := now + int64(p.Config.MaximumGatheringMinutes)*60000
	for _, s := range founders {
		if s.ExpiresAt < deadline {
			deadline = s.ExpiresAt
		}
	}
	return deadline
}
