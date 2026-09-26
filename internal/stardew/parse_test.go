package stardew

import (
	"bytes"
	"strings"
	"testing"
)

const sampleSave = `<SaveGame>
<player><name>Jepson</name><farmingLevel>9</farmingLevel><fishingLevel>7</fishingLevel><foragingLevel>6</foragingLevel><miningLevel>8</miningLevel><combatLevel>5</combatLevel><money>84233</money><millisecondsPlayed>9000000</millisecondsPlayed><achievements><int>1</int><int>2</int></achievements><friendshipData><item><key><string>Leah</string></key><value><Friendship><Points>1750</Points></Friendship></value></item></friendshipData></player>
<locations><GameLocation><name>Farm</name></GameLocation></locations>
<farmName>Cranberry</farmName><currentSeason>fall</currentSeason><dayOfMonth>18</dayOfMonth><year>2</year>
</SaveGame>`

func TestParseXML(t *testing.T) {
	p, err := parseXML(strings.NewReader(sampleSave))
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

func BenchmarkParseXML(b *testing.B) {
	data := []byte(sampleSave)
	r := bytes.NewReader(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		if _, err := parseXML(r); err != nil {
			b.Fatal(err)
		}
	}
}

func TestParseKeepsOnlyTopEightRelationships(t *testing.T) {
	var s strings.Builder
	s.WriteString(`<SaveGame><player><name>A</name><friendshipData>`)
	for i := 0; i < 10; i++ {
		s.WriteString(`<item><key><string>Friend`)
		s.WriteString(string(rune('A' + i)))
		s.WriteString(`</string></key><value><Friendship><Points>`)
		s.WriteString([]string{"250", "2500", "500", "2250", "750", "2000", "1000", "1750", "1250", "1500"}[i])
		s.WriteString(`</Points></Friendship></value></item>`)
	}
	s.WriteString(`</friendshipData></player></SaveGame>`)

	p, err := parseXML(strings.NewReader(s.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Relationships) != 8 {
		t.Fatalf("got %d relationships", len(p.Relationships))
	}
	if p.Relationships[0].Name != "FriendB" || p.Relationships[0].Points != 2500 {
		t.Fatalf("first relationship = %#v", p.Relationships[0])
	}
	if p.Relationships[7].Name != "FriendE" || p.Relationships[7].Points != 750 {
		t.Fatalf("last relationship = %#v", p.Relationships[7])
	}
}

func TestParseDecodesEscapedNames(t *testing.T) {
	p, err := parseXML(strings.NewReader(`<SaveGame><player><name>A &amp; B</name></player><farmName>Rock &amp; Roll</farmName></SaveGame>`))
	if err != nil {
		t.Fatal(err)
	}
	if p.PlayerName != "A & B" || p.FarmName != "Rock & Roll" {
		t.Fatalf("names = %q / %q", p.PlayerName, p.FarmName)
	}
}
