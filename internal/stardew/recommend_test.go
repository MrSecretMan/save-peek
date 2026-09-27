package stardew

import "testing"

func TestRecommendationsPrioritizeSeasonDeadline(t *testing.T) {
	p := Progress{
		Season: "Fall", Day: 27,
		Skills: Skills{Farming: 10, Fishing: 10, Foraging: 5, Mining: 4, Combat: 3},
	}
	addRecommendations(&p)
	if len(p.Recommendations) == 0 || p.Recommendations[0].Title != "Wrap up Fall goals" {
		t.Fatalf("recommendations = %#v", p.Recommendations)
	}
}

func TestRecommendationsUseSkillImbalance(t *testing.T) {
	p := Progress{Skills: Skills{Farming: 7, Fishing: 10, Foraging: 6, Mining: 4, Combat: 3}}
	addRecommendations(&p)
	if len(p.Recommendations) < 2 {
		t.Fatalf("recommendations = %#v", p.Recommendations)
	}
	if p.Recommendations[0].Title != "Work on combat" {
		t.Fatalf("primary = %#v", p.Recommendations[0])
	}
	if p.Recommendations[1].Title != "Work on mining" {
		t.Fatalf("secondary = %#v", p.Recommendations[1])
	}
}

func TestRecommendationsNoticeNearHeart(t *testing.T) {
	p := Progress{
		Skills: Skills{Farming: 5, Fishing: 5, Foraging: 5, Mining: 5, Combat: 5},
		Relationships: []Friend{{Name: "Leah", Points: 745, Hearts: 2}},
	}
	addRecommendations(&p)
	if len(p.Recommendations) == 0 || p.Recommendations[0].Title != "Check in with Leah" {
		t.Fatalf("recommendations = %#v", p.Recommendations)
	}
}

func TestRecommendationsAreCapped(t *testing.T) {
	p := Progress{Skills: Skills{Farming: 10, Fishing: 9, Foraging: 1, Mining: 2, Combat: 3}}
	addRecommendations(&p)
	if len(p.Recommendations) > 4 {
		t.Fatalf("got %d recommendations", len(p.Recommendations))
	}
}
