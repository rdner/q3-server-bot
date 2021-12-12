package events

import (
	"context"
	"errors"
	"io/ioutil"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rdner/q3-server-bot/pkg/parsers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestCapturing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	logrus.SetOutput(ioutil.Discard)

	cases := []struct {
		name       string
		outputs    []string
		senderErrs []error
		expEvents  []AnyEvent
	}{
		{
			name: "emits `ServerStartedEvent`",
			outputs: []string{
				`print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "
^3 0^7 total players

"`,
			},
			senderErrs: []error{nil},
			expEvents: []AnyEvent{
				&ServerStartedEvent{},
			},
		},
		{
			name: "emits `ServerStoppedEvent`",
			outputs: []string{
				`print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "
^3 0^7 total players

"`,
				"",
			},
			senderErrs: []error{nil, errors.New("not available")},
			expEvents: []AnyEvent{
				&ServerStartedEvent{},
				&ServerStoppedEvent{},
			},
		},

		{
			name: "emits `PlayerJoinedEvent`",
			outputs: []string{
				`print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "
^3 0^7 total players

"`,
				`print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "NOTREADY ^5:^7   0^5:^7 rdner                  25000       30     20     1  ^3^7
"print "
^3 1^7 total players

"`,
			},
			senderErrs: []error{nil, nil},
			expEvents: []AnyEvent{
				&ServerStartedEvent{},
				&PlayerJoinedEvent{
					Player: parsers.HumanEntry{
						PlayerEntry: parsers.PlayerEntry{
							Status: "NOTREADY",
							ID:     0,
							Player: "rdner",
						},
						Rate:    25000,
						MaxPkts: 30,
						Snaps:   20,
						PLost:   1,
					},
				},
			},
		},
		{
			name: "emits `PlayerLeftEvent`",
			outputs: []string{
				`print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "NOTREADY ^5:^7   0^5:^7 rdner                  25000       30     20     1  ^3^7
"print "
^3 1^7 total players

"`,
				`print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "
^3 0^7 total players

"`,
			},
			senderErrs: []error{nil, nil},
			expEvents: []AnyEvent{
				&ServerStartedEvent{},
				&PlayerJoinedEvent{
					Player: parsers.HumanEntry{
						PlayerEntry: parsers.PlayerEntry{
							Status: "NOTREADY",
							ID:     0,
							Player: "rdner",
						},
						Rate:    25000,
						MaxPkts: 30,
						Snaps:   20,
						PLost:   1,
					},
				},
				&PlayerLeftEvent{
					Player: parsers.HumanEntry{
						PlayerEntry: parsers.PlayerEntry{
							Status: "NOTREADY",
							ID:     0,
							Player: "rdner",
						},
						Rate:    25000,
						MaxPkts: 30,
						Snaps:   20,
						PLost:   1,
					},
				},
			},
		},

		{
			name: "emits `PlayerLeftEvent` and `PlayerJoinedEvent` when another player takes the spot",
			outputs: []string{
				`print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "^3(READY)  ^5:^7   1^5:^7 100.Anarki                 [BOT]  ^3^7
"print "^3(READY)  ^5:^7   2^5:^7 75.Doom                   [BOT]  ^3^7
"print "         ^5:^7   3^5:^7 player             25000       30     20     1  ^3^7
"print "
^3 3^7 total players

"`,
				`print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "^3(READY)  ^5:^7   1^5:^7 100.Anarki                 [BOT]  ^3^7
"print "^3(READY)  ^5:^7   2^5:^7 75.Doom                   [BOT]  ^3^7
"print "         ^5:^7   3^5:^7 new_player             25000       30     20     0  ^3^7
"print "
^3 4^7 total players

"`,
			},
			senderErrs: []error{
				nil,
				nil,
			},
			expEvents: []AnyEvent{
				&ServerStartedEvent{},
				&PlayerJoinedEvent{
					Player: parsers.BotEntry{
						PlayerEntry: parsers.PlayerEntry{
							Status: "(READY)",
							ID:     1,
							Player: "Anarki",
						},
						Level: 100,
					},
				},
				&PlayerJoinedEvent{
					Player: parsers.BotEntry{
						PlayerEntry: parsers.PlayerEntry{
							Status: "(READY)",
							ID:     2,
							Player: "Doom",
						},
						Level: 75,
					},
				},
				&PlayerJoinedEvent{
					Player: parsers.HumanEntry{
						PlayerEntry: parsers.PlayerEntry{
							Status: "",
							ID:     3,
							Player: "player",
						},
						Rate:    25000,
						MaxPkts: 30,
						Snaps:   20,
						PLost:   1,
					},
				},
				&PlayerLeftEvent{
					Player: parsers.HumanEntry{
						PlayerEntry: parsers.PlayerEntry{
							Status: "",
							ID:     3,
							Player: "player",
						},
						Rate:    25000,
						MaxPkts: 30,
						Snaps:   20,
						PLost:   1,
					},
				},
				&PlayerJoinedEvent{
					Player: parsers.HumanEntry{
						PlayerEntry: parsers.PlayerEntry{
							Status: "",
							ID:     3,
							Player: "new_player",
						},
						Rate:    25000,
						MaxPkts: 30,
						Snaps:   20,
						PLost:   0,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			startTime := time.Now()
			c := make(chan bool)

			sender := &senderMock{
				outputs:     tc.outputs,
				errs:        tc.senderErrs,
				outOfEvents: c,
			}
			m := NewManager(sender, time.Millisecond, time.Millisecond)
			events := m.Subscribe()

			go func() {
				err := m.StartCapturing(ctx)
				require.NoError(t, err)
			}()

			go func() {
				// once we're out events we close
				<-c
				m.Close()
				close(c)
			}()

			receivedEvents := make([]AnyEvent, 0, len(tc.expEvents))
			for e := range events {
				receivedEvents = append(receivedEvents, e)
			}

			checkAndEmptyTimestamps(t, receivedEvents, startTime)

			require.Len(t, receivedEvents, len(tc.expEvents))
			for _, ee := range tc.expEvents {
				require.Contains(t, receivedEvents, ee)
			}
		})
	}

	t.Run("returns `ErrAlreadyRunning` when try to run twice", func(t *testing.T) {
		c := make(chan bool)

		m := NewManager(&senderMock{outOfEvents: c}, time.Millisecond, time.Millisecond)
		defer func() {
			m.Close()
			close(c)
		}()

		go func() {
			err := m.StartCapturing(ctx)
			require.NoError(t, err)
		}()

		<-c
		err := m.StartCapturing(ctx)
		require.Error(t, err)
		require.Equal(t, ErrAlreadyRunning, err)

	})

	t.Run("emits same events for all subscribers", func(t *testing.T) {
		startTime := time.Now()
		c := make(chan bool)

		sender := &senderMock{
			outputs: []string{
				`print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "^3(READY)  ^5:^7   1^5:^7 100.Anarki                 [BOT]  ^3^7
"print "^3(READY)  ^5:^7   2^5:^7 75.Doom                   [BOT]  ^3^7
"print "         ^5:^7   3^5:^7 player             25000       30     20     1  ^3^7
"print "
^3 3^7 total players

"`,
				`print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "^3(READY)  ^5:^7   1^5:^7 100.Anarki                 [BOT]  ^3^7
"print "^3(READY)  ^5:^7   2^5:^7 75.Doom                   [BOT]  ^3^7
"print "         ^5:^7   3^5:^7 new_player             25000       30     20     0  ^3^7
"print "
^3 4^7 total players

"`,
			},
			errs:        []error{nil, nil},
			outOfEvents: c,
		}

		m := NewManager(sender, time.Millisecond, time.Millisecond)
		events1 := m.Subscribe()
		events2 := m.Subscribe()
		events3 := m.Subscribe()

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			err := m.StartCapturing(ctx)
			require.NoError(t, err)
			wg.Done()
		}()

		go func() {
			<-c
			m.Close()
			close(c)
		}()

		var (
			receivedEvents1 []AnyEvent
			receivedEvents2 []AnyEvent
			receivedEvents3 []AnyEvent
		)

		wg.Add(1)
		go func() {
			for e := range events1 {
				receivedEvents1 = append(receivedEvents1, e)
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			for e := range events2 {
				receivedEvents2 = append(receivedEvents2, e)
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			for e := range events3 {
				receivedEvents3 = append(receivedEvents3, e)
			}
			wg.Done()
		}()

		// wait until all events are captured and stored
		wg.Wait()

		// since all subscribers are supposed to receive a pointer to the same event
		// we need to perform this only once
		checkAndEmptyTimestamps(t, receivedEvents1, startTime)

		require.Len(t, receivedEvents1, len(receivedEvents2))
		require.Len(t, receivedEvents2, len(receivedEvents3))

		for _, e := range receivedEvents1 {
			require.Contains(t, receivedEvents2, e)
			require.Contains(t, receivedEvents3, e)
		}
	})

	t.Run("does not emit events if channel is unsubscribed", func(t *testing.T) {
		c := make(chan bool)

		sender := &senderMock{
			outputs: []string{
				`print
print "
^3Status^5   : ^3ID^5 : ^3Player                      Rate  MaxPkts  Snaps PLost
"print "^5----------------------------------------------------------------------^7
"print "^3(READY)  ^5:^7   1^5:^7 100.Anarki                 [BOT]  ^3^7
"print "
^3 1^7 total players

"`,
			},
			errs:        []error{nil, nil},
			outOfEvents: c,
		}

		m := NewManager(sender, time.Millisecond, time.Millisecond)
		events1 := m.Subscribe()
		events2 := m.Subscribe()
		events3 := m.Subscribe()
		err := m.Unsubscribe(events2)
		require.NoError(t, err)
		err = m.Unsubscribe(events3)
		require.NoError(t, err)

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			err := m.StartCapturing(ctx)
			require.NoError(t, err)
			wg.Done()
		}()

		go func() {
			<-c
			m.Close()
			close(c)
		}()

		var (
			receivedEvents1 []AnyEvent
			receivedEvents2 []AnyEvent
			receivedEvents3 []AnyEvent
		)

		wg.Add(1)
		go func() {
			for e := range events1 {
				receivedEvents1 = append(receivedEvents1, e)
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			for e := range events2 {
				receivedEvents2 = append(receivedEvents2, e)
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			for e := range events3 {
				receivedEvents3 = append(receivedEvents3, e)
			}
			wg.Done()
		}()

		// wait until all events are captured and stored
		wg.Wait()

		require.NotEmpty(t, receivedEvents1)
		require.Empty(t, receivedEvents2)
		require.Empty(t, receivedEvents3)
	})

	t.Run("returns `ErrSubscriptionNotFound", func(t *testing.T) {
		sender := &senderMock{}

		m := NewManager(sender, time.Millisecond, time.Millisecond)
		c := make(chan AnyEvent)
		defer close(c)
		require.Equal(t, ErrSubscriptionNotFound, m.Unsubscribe(c))
		require.Equal(t, ErrSubscriptionNotFound, m.Unsubscribe(nil))
	})
}

func checkAndEmptyTimestamps(t *testing.T, events []AnyEvent, startTime time.Time) {
	// we need to compare timestamps and reset them separately, so we can compare the events
	for _, e := range events {
		var timestamp *time.Time

		switch te := e.(type) {
		case *ServerStartedEvent:
			timestamp = &te.When
		case *ServerStoppedEvent:
			timestamp = &te.When
		case *PlayerJoinedEvent:
			timestamp = &te.When
		case *PlayerLeftEvent:
			timestamp = &te.When
		default:
			require.FailNowf(t, "unexpected event type", "type %T", e)
		}

		require.WithinDuration(t, *timestamp, startTime, time.Since(startTime))
		*timestamp = time.Time{} // set empty timestamp
	}

}

type senderMock struct {
	outputs         []string
	index           int32
	errs            []error
	receivedCommand string
	outOfEvents     chan<- bool
}

func (s *senderMock) Send(ctx context.Context, command string) (output string, err error) {
	s.receivedCommand = command

	// we atomically increment the index but we need the previous value
	i := int(atomic.AddInt32(&s.index, 1) - 1)
	if i >= len(s.outputs) {
		// notify we don't have events anymore
		s.outOfEvents <- true

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		// block for some time just in case
		case <-time.After(100 * time.Millisecond):
			return "", nil
		}
	}
	return s.outputs[i], s.errs[i]
}
