package stardew

import (
	"strings"
	"testing"
)

func TestParseXML(t *testing.T) {
	xml := `<SaveGame>
<player><name>Jepson</name><farmingLevel>9</farmingLevel><fishingLevel>7</fishingLevel><foragingLevel>6</foragingLevel><miningLevel>8</miningLevel><combatLevel>5</combatLevel><money>84233</money><millisecondsPlayed>9000000</millisecondsPlayed><achievements><int>1</int><int>2</int></achievements><friendshipData><item><key><string>Leah</string></key><value><Friendship><Points>1750</Points></Friendship></value></item></friendshipData></player>
<farmName>Cranberry</farmName><currentSeason>fall</currentSeason><dayOfMonth>18</dayOfMonth><year>2</year>
</SaveGame>`

	p, err := parseXML(strings.NewReader(xml))
	if err != nil {
		t.Fatal(err)
	}
	if p.PlayerName != "Jepson" || p.FarmName != "Cranberry" {
		t.Fatalf("names: %#v", p)
	}
	if p.Season != "Fall" || p.Day != 18 || p.Year != 2 {
		t.Fatalf("date: %#v", p)
	}
	if p.Money != 84233 || p.Skills.Farming != 9 || p.Achievements != 2 {
		t.Fatalf("progress: %#v", p)
	}
	if len(p.Relationships) != 1 || p.Relationships[0].Name != "Leah" || p.Relationships[0].Hearts != 7 {
		t.Fatalf("relationships: %#v", p.Relationships)
	}
}
