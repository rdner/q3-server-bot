package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rdner/q3-server-bot/pkg/events"
	"github.com/rdner/q3-server-bot/pkg/parsers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type testEvent struct {
	event      events.AnyEvent
	delayAfter time.Duration
}

func TestServerEventsBot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	logrus.SetOutput(ioutil.Discard)

	cases := []struct {
		name           string
		events         []testEvent
		expMessages    []string
		token          string
		chatID         string
		serverAddr     string
		throttling     time.Duration
		apiAvailableIn time.Duration
	}{
		{
			name: "sends 'server started' message",
			events: []testEvent{
				{
					event:      &events.ServerStartedEvent{},
					delayAfter: 5 * time.Millisecond,
				},
			},
			expMessages: []string{
				"Server started. Join the game!\n\nUse this console command to connect:\n```\n\\connect localhost:27960\n```"},
			token:      "token",
			chatID:     "chatID",
			serverAddr: "localhost:27960",
			throttling: time.Millisecond,
		},
		{
			name: "sends 'server stopped' message",
			events: []testEvent{
				{
					event:      &events.ServerStoppedEvent{},
					delayAfter: 5 * time.Millisecond,
				},
			},
			expMessages: []string{"Server stopped."},
			token:       "token",
			chatID:      "chatID",
			serverAddr:  "localhost:27960",
			throttling:  time.Millisecond,
		},
		{
			name: "sends only single server status-related message",
			events: []testEvent{
				{
					event:      &events.ServerStartedEvent{},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event:      &events.ServerStoppedEvent{},
					delayAfter: 5 * time.Millisecond,
				},
			},
			expMessages: []string{"Server stopped."},
			token:       "token",
			chatID:      "chatID",
			serverAddr:  "localhost:27960",
			throttling:  5 * time.Millisecond,
		},
		{
			name: "sends only single server status-related message with blinking events",
			events: []testEvent{
				{
					event:      &events.ServerStartedEvent{},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event:      &events.ServerStoppedEvent{},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event:      &events.ServerStartedEvent{},
					delayAfter: 5 * time.Millisecond,
				},
			},
			expMessages: []string{
				"Server started. Join the game!\n\nUse this console command to connect:\n```\n\\connect localhost:27960\n```"},
			token:      "token",
			chatID:     "chatID",
			serverAddr: "localhost:27960",
			throttling: 10 * time.Millisecond,
		},

		{
			name: "sends 'player joined' message",
			events: []testEvent{
				{
					event: &events.PlayerJoinedEvent{
						Player: parsers.PlayerEntry{
							Player: "player1",
						},
					},
					delayAfter: 5 * time.Millisecond,
				},
			},
			expMessages: []string{
				"`player1` joined the game\nUse this console command to connect:\n```\n\\connect localhost:27960\n```",
			},
			token:      "token",
			chatID:     "chatID",
			serverAddr: "localhost:27960",
			throttling: time.Millisecond,
		},
		{
			name: "sends 'players joined' message for a list of players",
			events: []testEvent{
				{
					event: &events.PlayerJoinedEvent{
						Player: parsers.PlayerEntry{
							Player: "player1",
						},
					},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event: &events.PlayerJoinedEvent{
						Player: parsers.PlayerEntry{
							Player: "player2",
						},
					},
					delayAfter: 10 * time.Millisecond,
				},
			},
			expMessages: []string{
				"`player1`, `player2` joined the game\nUse this console command to connect:\n```\n\\connect localhost:27960\n```",
			},
			token:      "token",
			chatID:     "chatID",
			serverAddr: "localhost:27960",
			throttling: 3 * time.Millisecond,
		},
		{
			name: "sends 'player left' message",
			events: []testEvent{
				{
					event: &events.PlayerLeftEvent{
						Player: parsers.PlayerEntry{
							Player: "player1",
						},
					},
					delayAfter: 5 * time.Millisecond,
				},
			},
			expMessages: []string{
				"`player1` left the game\nUse this console command to connect:\n```\n\\connect localhost:27960\n```",
			},
			token:      "token",
			chatID:     "chatID",
			serverAddr: "localhost:27960",
			throttling: time.Millisecond,
		},
		{
			name: "sends 'players left' message for a list of players",
			events: []testEvent{
				{
					event: &events.PlayerLeftEvent{
						Player: parsers.PlayerEntry{
							Player: "player1",
						},
					},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event: &events.PlayerLeftEvent{
						Player: parsers.PlayerEntry{
							Player: "player2",
						},
					},
					delayAfter: 10 * time.Millisecond,
				},
			},
			expMessages: []string{
				"`player1`, `player2` left the game\nUse this console command to connect:\n```\n\\connect localhost:27960\n```",
			},
			token:      "token",
			chatID:     "chatID",
			serverAddr: "localhost:27960",
			throttling: 3 * time.Millisecond,
		},
		{
			name: "groups all kinds of messages together",
			events: []testEvent{
				{
					event:      &events.ServerStartedEvent{},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event: &events.PlayerJoinedEvent{
						Player: parsers.PlayerEntry{
							Player: "player1",
						},
					},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event: &events.PlayerJoinedEvent{
						Player: parsers.PlayerEntry{
							Player: "player2",
						},
					},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event: &events.PlayerLeftEvent{
						Player: parsers.PlayerEntry{
							Player: "player3",
						},
					},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event: &events.PlayerLeftEvent{
						Player: parsers.PlayerEntry{
							Player: "player4",
						},
					},
					delayAfter: 50 * time.Millisecond,
				},
			},
			expMessages: []string{
				"Server started. Join the game!\n\n`player1`, `player2` joined the game\n`player3`, `player4` left the game\nUse this console command to connect:\n```\n\\connect localhost:27960\n```",
			},
			token:      "token",
			chatID:     "chatID",
			serverAddr: "localhost:27960",
			throttling: 10 * time.Millisecond,
		},

		{
			name: "does not lose accumulated messages if the API is not available",
			events: []testEvent{
				{
					event:      &events.ServerStartedEvent{},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event: &events.PlayerJoinedEvent{
						Player: parsers.PlayerEntry{
							Player: "player1",
						},
					},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event: &events.PlayerJoinedEvent{
						Player: parsers.PlayerEntry{
							Player: "player2",
						},
					},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event: &events.PlayerLeftEvent{
						Player: parsers.PlayerEntry{
							Player: "player3",
						},
					},
					delayAfter: 1 * time.Millisecond,
				},
				{
					event: &events.PlayerLeftEvent{
						Player: parsers.PlayerEntry{
							Player: "player4",
						},
					},
					delayAfter: 300 * time.Millisecond,
				},
			},
			expMessages: []string{
				"Server started. Join the game!\n\n`player1`, `player2` joined the game\n`player3`, `player4` left the game\nUse this console command to connect:\n```\n\\connect localhost:27960\n```",
			},
			token:          "token",
			chatID:         "chatID",
			serverAddr:     "localhost:27960",
			throttling:     2 * time.Millisecond,
			apiAvailableIn: 100 * time.Millisecond,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var messages []string
			now := time.Now()
			// mock Telegram API
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if time.Since(now) <= tc.apiAvailableIn {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				require.Equal(t, fmt.Sprintf("/bot%s/sendMessage", tc.token), r.URL.Path)
				require.Equal(t, "application/json", r.Header.Get("Content-Type"))

				defer r.Body.Close()
				payload, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				var msg telegramMessage
				err = json.Unmarshal(payload, &msg)
				require.NoError(t, err)
				require.Equal(t, tc.chatID, msg.ChatID)
				require.NotNil(t, msg.ParseMode)
				require.Equal(t, Mardown, *msg.ParseMode)
				messages = append(messages, msg.Text)
				w.WriteHeader(http.StatusOK)
			}))
			defer s.Close()

			em := eventMngrMock{
				testEvents:  tc.events,
				events:      make(chan events.AnyEvent),
				outOfEvents: make(chan bool),
			}
			b := NewServerEventsBot(
				em,
				tc.token,
				tc.chatID,
				s.URL,
				tc.serverAddr,
				tc.throttling,
			)

			go func() {
				err := em.StartCapturing(ctx)
				require.NoError(t, err)
			}()
			go func() {
				<-em.outOfEvents
				err := em.Close()
				require.NoError(t, err)
				err = b.Close()
				require.NoError(t, err)
			}()

			err := b.Start(ctx)
			require.NoError(t, err)

			require.Len(t, messages, len(tc.expMessages))
			for _, m := range tc.expMessages {
				require.Contains(t, messages, m)
			}
		})
	}

	t.Run("returns `ErrBotRunning` when try to run twice", func(t *testing.T) {
		em := eventMngrMock{
			testEvents: []testEvent{
				{
					event:      &events.ServerStartedEvent{},
					delayAfter: 10 * time.Millisecond,
				},
			},
			events:      make(chan events.AnyEvent),
			outOfEvents: make(chan bool),
		}
		b := NewServerEventsBot(
			em,
			"token",
			"chatID",
			"localhost",
			"localhost:27960",
			1*time.Millisecond,
		)

		defer func() {
			b.Close()
		}()

		go func() {
			err := b.Start(ctx)
			require.NoError(t, err)
		}()
		go func() {
			err := em.StartCapturing(ctx)
			require.NoError(t, err)
		}()

		<-em.outOfEvents

		err := b.Start(ctx)
		require.Error(t, err)
		require.Equal(t, ErrBotRunning, err)
	})
}

type eventMngrMock struct {
	testEvents  []testEvent
	events      chan events.AnyEvent
	outOfEvents chan bool
}

func (m eventMngrMock) StartCapturing(ctx context.Context) error {
	for _, e := range m.testEvents {
		m.events <- e.event
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(e.delayAfter):
		}
	}

	m.outOfEvents <- true
	close(m.outOfEvents)
	return nil
}

func (m eventMngrMock) Subscribe() <-chan events.AnyEvent { return m.events }
func (m eventMngrMock) Unsubscribe(<-chan events.AnyEvent) error {
	close(m.events)
	return nil
}

func (m eventMngrMock) Close() error { return nil }
