package parsers

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	// matches the row of the players table
	playerEntryRowRE = regexp.MustCompile(`^\s+"(?:\^\d)*(.*)\s+(?:\^\d)*:(?:\^\d)*\s+(\d+)(?:\^\d)*:(?:\^\d)*\s+(\S+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(?:\^\d)*\n"$`)
	// matches the bot row of the players table
	botEntryRowRE = regexp.MustCompile(`^\s+"(?:\^\d)*(.*)\s+(?:\^\d)*:(?:\^\d)*\s+(\d+)(?:\^\d)*:(?:\^\d)*\s+(\d+)\.(\S+)\s+\[BOT\]\s+(?:\^\d)*\s*\n"$`)
)

// Player represents a player on the server
type Player interface {

	// GetStatus gets the status of the player (e.g. `READY` or `NOTREADY`).
	GetStatus() string
	// GetID gets a numeric ID of the player
	GetID() int
	// GetName gets the name of the player
	GetName() string
}

// PlayerEntry represents basic information about the player.
// Implements `Player` interface
type PlayerEntry struct {
	Status string
	ID     int
	Player string
}

func (p PlayerEntry) GetStatus() string {
	return p.Status
}
func (p PlayerEntry) GetID() int {
	return p.ID
}
func (p PlayerEntry) GetName() string {
	return p.Player
}

// HumanEntry represents the human player entry on the server
type HumanEntry struct {
	PlayerEntry

	// Network-related information

	Rate    int
	MaxPkts int
	Snaps   int
	PLost   int
}

type BotEntry struct {
	PlayerEntry

	// Level of the bot (0-100)
	Level int
}

// ParsePlayers tries to find information about human players and bots parsing the output
// of `rcon <password> players`.
func ParsePlayers(str string) (players []Player) {
	lines := strings.Split(str, "print")

	for _, line := range lines {
		human, err := tryParseHuman(line)
		if err == nil {
			players = append(players, human)
			continue
		}
		bot, err := tryParseBot(line)
		if err == nil {
			players = append(players, bot)
			continue
		}
	}
	return players
}

func tryParseHuman(line string) (player HumanEntry, err error) {
	submatches := playerEntryRowRE.FindAllStringSubmatch(line, -1)
	if len(submatches) != 1 {
		return player, errors.New("the row does not match the player row")
	}
	row := submatches[0]
	if len(row) != 8 {
		return player, errors.New("the row does not have the right amount of columns")
	}

	row = row[1:]

	player.Status = strings.TrimSpace(row[0])

	player.ID, err = strconv.Atoi(row[1])
	if err != nil {
		return player, err
	}

	player.Player = strings.TrimSpace(row[2])

	player.Rate, err = strconv.Atoi(row[3])
	if err != nil {
		return player, err
	}
	player.MaxPkts, err = strconv.Atoi(row[4])
	if err != nil {
		return player, err
	}
	player.Snaps, err = strconv.Atoi(row[5])
	if err != nil {
		return player, err
	}
	player.PLost, err = strconv.Atoi(row[6])
	if err != nil {
		return player, err
	}

	return player, nil
}

func tryParseBot(line string) (bot BotEntry, err error) {
	submatches := botEntryRowRE.FindAllStringSubmatch(line, -1)
	if len(submatches) != 1 {
		return bot, errors.New("the row does not match the bot row")
	}
	row := submatches[0]
	if len(row) != 5 {
		return bot, errors.New("the row does not have the right amount of columns")
	}

	row = row[1:]

	bot.Status = strings.TrimSpace(row[0])

	bot.ID, err = strconv.Atoi(row[1])
	if err != nil {
		return bot, err
	}

	bot.Level, err = strconv.Atoi(row[2])
	if err != nil {
		return bot, err
	}
	bot.Player = strings.TrimSpace(row[3])

	return bot, nil
}
