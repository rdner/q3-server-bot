package parsers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePlayers(t *testing.T) {
	cases := []struct {
		name  string
		value string
		exp   []Player
	}{
		{
			name: "returns valid list of players when the table has values",
			value: `print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "NOTREADY ^5:^7   0^5:^7 rdner                  25000       30     20     1  ^3^7
"print "^3(READY)  ^5:^7   1^5:^7 100.Anarki                 [BOT]  ^3^7
"print "^3(READY)  ^5:^7   2^5:^7 75.Doom                   [BOT]  ^3^7
"print "^3(READY)  ^5:^7   3^5:^7 50.Slash                  [BOT]  ^3^7
"print "         ^5:^7   4^5:^7 100.Slash                  [BOT]  ^3^7
"print "         ^5:^7   5^5:^7 another_player             25000       30     20     0  ^3^7
"print "
^3 5^7 total players

"`,
			exp: []Player{
				HumanEntry{
					PlayerEntry: PlayerEntry{
						Status: "NOTREADY",
						ID:     0,
						Player: "rdner",
					},
					Rate:    25000,
					MaxPkts: 30,
					Snaps:   20,
					PLost:   1,
				},
				BotEntry{
					PlayerEntry: PlayerEntry{
						Status: "(READY)",
						ID:     1,
						Player: "Anarki",
					},
					Level: 100,
				},
				BotEntry{
					PlayerEntry: PlayerEntry{
						Status: "(READY)",
						ID:     2,
						Player: "Doom",
					},
					Level: 75,
				},
				BotEntry{
					PlayerEntry: PlayerEntry{
						Status: "(READY)",
						ID:     3,
						Player: "Slash",
					},
					Level: 50,
				},
				BotEntry{
					PlayerEntry: PlayerEntry{
						Status: "",
						ID:     4,
						Player: "Slash",
					},
					Level: 100,
				},
				HumanEntry{
					PlayerEntry: PlayerEntry{
						Status: "",
						ID:     5,
						Player: "another_player",
					},
					Rate:    25000,
					MaxPkts: 30,
					Snaps:   20,
					PLost:   0,
				},
			},
		},
		{
			name: "returns empty list if no players",
			value: `print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "
^3 0^7 total players

"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.EqualValues(t, tc.exp, ParsePlayers(tc.value))
		})
	}
}
