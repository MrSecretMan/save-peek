package stardew

import (
	"fmt"
	"sort"
)

type Recommendation struct {
	Title  string `json:"title"`
	Reason string `json:"reason"`
}

type scoredRecommendation struct {
	Recommendation
	score int
	order int
}

// addRecommendations derives a short, deterministic plan from facts already
// present in the save. It deliberately avoids guessing about inventory,
// bundles, quests, or mine floors that SavePeek does not parse yet.
func addRecommendations(p *Progress) {
	candidates := make([]scoredRecommendation, 0, 8)
	add := func(score, order int, title, reason string) {
		candidates = append(candidates, scoredRecommendation{
			Recommendation: Recommendation{Title: title, Reason: reason},
			score: score, order: order,
		})
	}

	if p.Day >= 24 && p.Day <= 28 && p.Season != "" {
		add(100, 0,
			fmt.Sprintf("Wrap up %s goals", p.Season),
			fmt.Sprintf("Only %d day(s) remain in the season; finish anything seasonal before it rolls over.", 29-p.Day))
	}

	skills := []struct {
		name  string
		level int
		order int
	}{
		{"farming", p.Skills.Farming, 0},
		{"mining", p.Skills.Mining, 1},
		{"fishing", p.Skills.Fishing, 2},
		{"foraging", p.Skills.Foraging, 3},
		{"combat", p.Skills.Combat, 4},
	}
	maxLevel := 0
	for _, skill := range skills {
		if skill.level > maxLevel {
			maxLevel = skill.level
		}
	}
	for _, skill := range skills {
		if skill.level >= 10 || maxLevel-skill.level < 2 {
			continue
		}
		add(70+(maxLevel-skill.level)*3, 10+skill.order,
			fmt.Sprintf("Work on %s", skill.name),
			fmt.Sprintf("%s is level %d while your highest skill is %d; a session here will round out the save.", titleCase(skill.name), skill.level, maxLevel))
	}

	for i, friend := range p.Relationships {
		if friend.Points <= 0 || friend.Hearts >= 10 {
			continue
		}
		toNext := 250 - friend.Points%250
		if toNext == 250 {
			continue
		}
		if toNext <= 80 {
			add(82, 30+i,
				fmt.Sprintf("Check in with %s", friend.Name),
				fmt.Sprintf("%s is only %d friendship points from the next heart.", friend.Name, toNext))
		}
	}

	if len(candidates) == 0 {
		for _, skill := range skills {
			if skill.level < 10 {
				add(50-skill.level, 50+skill.order,
					fmt.Sprintf("Work on %s", skill.name),
					fmt.Sprintf("%s is level %d and still has room to grow.", titleCase(skill.name), skill.level))
			}
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].order < candidates[j].order
	})

	limit := 4
	if len(candidates) < limit {
		limit = len(candidates)
	}
	p.Recommendations = make([]Recommendation, limit)
	for i := 0; i < limit; i++ {
		p.Recommendations[i] = candidates[i].Recommendation
	}
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	if b[0] >= 'a' && b[0] <= 'z' {
		b[0] -= 'a' - 'A'
	}
	return string(b)
}
