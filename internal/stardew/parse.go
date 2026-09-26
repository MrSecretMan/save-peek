package stardew

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
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
	var text []byte

	depth := 0
	achievementDepth := 0
	friendshipDepth := 0
	keyDepth := 0
	captureDepth := 0
	captureName := ""
	friendName := ""
	friendPoints := 0

	for {
		tok, err := dec.RawToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Progress{}, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			name := t.Name.Local

			if name == "achievements" && achievementDepth == 0 {
				achievementDepth = depth
			}
			if name == "friendshipData" && friendshipDepth == 0 {
				friendshipDepth = depth
			}
			if name == "key" && friendshipDepth > 0 && keyDepth == 0 {
				keyDepth = depth
			}

			if shouldCapture(name, p.PlayerName == "", achievementDepth, friendshipDepth, keyDepth, depth) {
				captureDepth = depth
				captureName = name
				text = text[:0]
			}

		case xml.CharData:
			if captureDepth == depth {
				text = append(text, t...)
			}

		case xml.EndElement:
			name := t.Name.Local
			if captureDepth == depth && captureName == name {
				value := bytes.TrimSpace(text)
				if len(value) > 0 {
					switch name {
					case "name":
						if p.PlayerName == "" && friendshipDepth == 0 {
							p.PlayerName = string(value)
						}
					case "farmName":
						p.FarmName = string(value)
					case "money":
						p.Money = int64ValueBytes(value)
					case "currentSeason":
						p.Season = seasonName(value)
					case "dayOfMonth":
						p.Day = int(int64ValueBytes(value))
					case "year":
						p.Year = int(int64ValueBytes(value))
					case "millisecondsPlayed":
						p.PlayTime = int64ValueBytes(value)
					case "farmingLevel":
						p.Skills.Farming = int(int64ValueBytes(value))
					case "fishingLevel":
						p.Skills.Fishing = int(int64ValueBytes(value))
					case "foragingLevel":
						p.Skills.Foraging = int(int64ValueBytes(value))
					case "miningLevel":
						p.Skills.Mining = int(int64ValueBytes(value))
					case "combatLevel":
						p.Skills.Combat = int(int64ValueBytes(value))
					case "int":
						if achievementDepth > 0 && depth > achievementDepth {
							p.Achievements++
						}
					case "string":
						if friendshipDepth > 0 && keyDepth > 0 {
							friendName = string(value)
						}
					case "Points", "points":
						if friendshipDepth > 0 {
							friendPoints = int(int64ValueBytes(value))
						}
					}
				}
				captureDepth = 0
				captureName = ""
				text = text[:0]
			}

			if name == "item" && friendshipDepth > 0 && friendName != "" {
				p.Relationships = addTopFriend(p.Relationships, Friend{
					Name: friendName, Points: friendPoints, Hearts: friendPoints / 250,
				})
				friendName, friendPoints = "", 0
			}
			if name == "key" && keyDepth == depth {
				keyDepth = 0
			}
			if name == "achievements" && achievementDepth == depth {
				achievementDepth = 0
			}
			if name == "friendshipData" && friendshipDepth == depth {
				friendshipDepth = 0
			}
			depth--
		}
	}

	return p, nil
}

func shouldCapture(name string, needPlayerName bool, achievementDepth, friendshipDepth, keyDepth, depth int) bool {
	switch name {
	case "farmName", "money", "currentSeason", "dayOfMonth", "year", "millisecondsPlayed",
		"farmingLevel", "fishingLevel", "foragingLevel", "miningLevel", "combatLevel":
		return true
	case "name":
		return needPlayerName && friendshipDepth == 0
	case "int":
		return achievementDepth > 0 && depth > achievementDepth
	case "string":
		return friendshipDepth > 0 && keyDepth > 0
	case "Points", "points":
		return friendshipDepth > 0
	default:
		return false
	}
}

func addTopFriend(top []Friend, friend Friend) []Friend {
	const limit = 8
	if cap(top) < limit {
		grown := make([]Friend, len(top), limit)
		copy(grown, top)
		top = grown
	}

	at := len(top)
	for i := range top {
		if friend.Points > top[i].Points || (friend.Points == top[i].Points && friend.Name < top[i].Name) {
			at = i
			break
		}
	}
	if at >= limit {
		return top
	}
	if len(top) < limit {
		top = append(top, Friend{})
	}
	copy(top[at+1:], top[at:len(top)-1])
	top[at] = friend
	return top
}

func int64ValueBytes(b []byte) int64 {
	var n int64
	sign := int64(1)
	if len(b) > 0 && b[0] == '-' {
		sign = -1
		b = b[1:]
	}
	for _, c := range b {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int64(c-'0')
	}
	return n * sign
}

func seasonName(b []byte) string {
	switch {
	case bytes.Equal(b, []byte("spring")):
		return "Spring"
	case bytes.Equal(b, []byte("summer")):
		return "Summer"
	case bytes.Equal(b, []byte("fall")):
		return "Fall"
	case bytes.Equal(b, []byte("winter")):
		return "Winter"
	default:
		return string(b)
	}
}
