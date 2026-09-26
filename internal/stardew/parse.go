package stardew

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

func Parse(save Save) (Progress, error) {
	f, err := os.Open(save.Path)
	if err != nil {
		return Progress{}, err
	}
	defer f.Close()

	p, err := parseXML(f)
	if err != nil {
		return Progress{}, fmt.Errorf("parse %s: %w", save.Folder, err)
	}
	p.Source = Source{Folder: save.Folder, ModifiedAt: save.ModifiedAt}
	return p, nil
}

func parseXML(r io.Reader) (Progress, error) {
	dec := xml.NewDecoder(r)
	var p Progress
	var path []string
	var text strings.Builder
	var achievementDepth int
	var friendshipDepth int
	var friendName string
	var friendPoints int

	flush := func(name string) {
		value := strings.TrimSpace(text.String())
		text.Reset()
		if value == "" {
			return
		}

		switch name {
		case "name":
			if p.PlayerName == "" && !inside(path, "friendshipData") {
				p.PlayerName = value
			}
		case "farmName":
			p.FarmName = value
		case "money":
			p.Money = int64Value(value)
		case "currentSeason":
			p.Season = title(value)
		case "dayOfMonth":
			p.Day = intValue(value)
		case "year":
			p.Year = intValue(value)
		case "millisecondsPlayed":
			p.PlayTime = int64Value(value)
		case "farmingLevel":
			p.Skills.Farming = intValue(value)
		case "fishingLevel":
			p.Skills.Fishing = intValue(value)
		case "foragingLevel":
			p.Skills.Foraging = intValue(value)
		case "miningLevel":
			p.Skills.Mining = intValue(value)
		case "combatLevel":
			p.Skills.Combat = intValue(value)
		}
	}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Progress{}, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			path = append(path, t.Name.Local)
			text.Reset()
			if t.Name.Local == "achievements" {
				achievementDepth = len(path)
			}
			if t.Name.Local == "friendshipData" {
				friendshipDepth = len(path)
			}
		case xml.CharData:
			text.Write([]byte(t))
		case xml.EndElement:
			name := t.Name.Local
			value := strings.TrimSpace(text.String())

			if achievementDepth > 0 && len(path) > achievementDepth && name == "int" && value != "" {
				p.Achievements++
			}

			if friendshipDepth > 0 {
				if name == "string" && inside(path, "key") && value != "" {
					friendName = value
				}
				if name == "Points" || name == "points" {
					friendPoints = intValue(value)
				}
				if name == "item" && friendName != "" {
					p.Relationships = append(p.Relationships, Friend{Name: friendName, Points: friendPoints, Hearts: friendPoints / 250})
					friendName, friendPoints = "", 0
				}
			}

			flush(name)
			if achievementDepth == len(path) && name == "achievements" {
				achievementDepth = 0
			}
			if friendshipDepth == len(path) && name == "friendshipData" {
				friendshipDepth = 0
			}
			if len(path) > 0 {
				path = path[:len(path)-1]
			}
			text.Reset()
		}
	}

	sort.Slice(p.Relationships, func(i, j int) bool {
		if p.Relationships[i].Points == p.Relationships[j].Points {
			return p.Relationships[i].Name < p.Relationships[j].Name
		}
		return p.Relationships[i].Points > p.Relationships[j].Points
	})
	if len(p.Relationships) > 8 {
		p.Relationships = p.Relationships[:8]
	}
	return p, nil
}

func inside(path []string, want string) bool {
	for _, part := range path {
		if part == want {
			return true
		}
	}
	return false
}

func intValue(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func int64Value(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
