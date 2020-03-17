package events

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pragmader/q3-server-bot/pkg/parsers"
	"github.com/pragmader/q3-server-bot/pkg/rcon"
	"github.com/sirupsen/logrus"
)

var (
	// ErrSubscriptionNotFound occurs when the subscriber attempts to unsubscribe from
	// a subscription that does not exist
	ErrSubscriptionNotFound = errors.New("subscriber was not found")

	// ErrAlreadyRunning occurs when `StartCapturing` is called twice without closing
	ErrAlreadyRunning = errors.New("event capturing is in progress")
)

const (
	defaultChannelBuffer = 16 // 16 events per subscription before blocking
)

// AnyEvent tepresents any even on the server
type AnyEvent interface{}

// EventBase is an event that contains basic information
type EventBase struct {
	When time.Time
}

// ServerStartedEvent occurs when the server starts
type ServerStartedEvent EventBase

// ServerStoppedEvent occurs when the server stops
type ServerStoppedEvent EventBase

// PlayerEvent is a player-related event that contains information about the player
type PlayerEvent struct {
	EventBase

	// Player involved in the event
	Player parsers.Player
}

// PlayerJoinedEvent occurs when a player joins the server
type PlayerJoinedEvent PlayerEvent

// PlayerLeftEvent occurs when a player leaves the server
type PlayerLeftEvent PlayerEvent

type state struct {
	serverAvailable bool
	playersOutput   string
	// playersByKeys is a map of players by {ID:Name}, since it's the only unique identifier
	// Unfortunately, there is no way to distinguish whether the player changed its name
	// or a player left and another player joined. In both cases the subscriber will receive 2
	// events `PlayerLeftEvent` and `PlayerJoinedEvent`.
	playersByKeys map[string]parsers.Player
}

// Manager describes the event dispatcher and its available operations.
type Manager interface {
	// StartCapturing starts the loop of event capturing from the server.
	// This function can be run only once, following times it will return `ErrAlreadyRunning`.
	StartCapturing(ctx context.Context) error

	// Subscribe returns a read-only channel to a subscriber where it can consume server events.
	// This operation is thread-safe.
	Subscribe() <-chan AnyEvent

	// Unsubscribe takes the channel previously returned by `Subscribe` and removes it
	// from the list of subscriptions.
	// This operation is thread-safe.
	Unsubscribe(<-chan AnyEvent) error

	// Close closes all subscription channels and stops the event capturing loop.
	Close() error
}

// NewManager creates a new instance of the event manager.
func NewManager(sender rcon.Sender, refreshInterval, connectionTimeout time.Duration) Manager {
	return &manager{
		sender:            sender,
		refreshInterval:   refreshInterval,
		connectionTimeout: connectionTimeout,
		mutex:             &sync.RWMutex{},
		closed:            make(chan bool),
	}
}

type manager struct {
	sender            rcon.Sender
	subscriptions     []chan AnyEvent
	refreshInterval   time.Duration
	connectionTimeout time.Duration
	mutex             *sync.RWMutex
	state             state
	closed            chan bool
	runningCnt        int32
}

func (m *manager) Close() error {
	logrus := logrus.
		WithField("package", "events").
		WithField("module", "EventManager").
		WithField("function", "Close")

	logrus.Debug("closing...")

	m.closed <- true // signal the stop of the capturing

	logrus.Debug("closing signal sent")

	logrus.Debug("closing subscription channels...")
	m.mutex.Lock()
	for _, c := range m.subscriptions {
		close(c)
	}
	m.subscriptions = nil
	m.mutex.Unlock()

	logrus.Debug("all subscription channels are closed")

	return nil
}

func (m *manager) StartCapturing(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&m.runningCnt, 0, 1) {
		return ErrAlreadyRunning
	}

	defer func() {
		atomic.SwapInt32(&m.runningCnt, 0)
	}()

	logrus := logrus.
		WithField("package", "events").
		WithField("module", "EventManager").
		WithField("function", "StartCapturing")

	logrus.Debug("starting capturing loop...")
	// start the first iteration immediately
	s := m.computeState(ctx)
	m.applyState(ctx, s)

	// other iterations come after a refresh interval
	for {
		logrus.Debugf("starting next iteration in %s...", m.refreshInterval)
		select {
		case <-m.closed:
			logrus.Debug("got closing signal, stopped")
			return nil
		case <-ctx.Done():
			logrus.Debug("got context cancellation, stopped")
			return ctx.Err()
		case <-time.After(m.refreshInterval):
			logrus.Debug("starting iteration...")
			s := m.computeState(ctx)
			m.applyState(ctx, s)
		}
	}

	return nil
}

func (m *manager) applyState(ctx context.Context, new state) {
	logrus := logrus.
		WithField("package", "events").
		WithField("module", "EventManager").
		WithField("function", "applyState")

	logrus.Debug("calculating events and saving state...")
	defer logrus.Debug("state has been updated")

	old := m.state

	// we ignore the errors returned from `emitEvent` on purpose,
	// the only possible error there is cancellation, which we don't have
	// to handle

	if !old.serverAvailable && new.serverAvailable {
		go m.emitEvent(ctx, &ServerStartedEvent{
			When: time.Now(),
		})
	}

	if old.serverAvailable && !new.serverAvailable {
		go m.emitEvent(ctx, &ServerStoppedEvent{
			When: time.Now(),
		})
	}

	if old.playersOutput != new.playersOutput {
		// we have to create a go routine per subscription, otherwise,
		// if the subscriber did not read events from the channel it would block
		// the entire event distribution
		for key, p := range old.playersByKeys {
			if _, ok := new.playersByKeys[key]; ok {
				continue
			}
			go m.emitEvent(ctx, &PlayerLeftEvent{
				EventBase: EventBase{
					When: time.Now(),
				},
				Player: p,
			})
		}

		for key, p := range new.playersByKeys {
			if _, ok := old.playersByKeys[key]; ok {
				continue
			}
			go m.emitEvent(ctx, &PlayerJoinedEvent{
				EventBase: EventBase{
					When: time.Now(),
				},
				Player: p,
			})
		}
	}

	m.state = new
}

func (m *manager) emitEvent(ctx context.Context, e AnyEvent) error {
	logrus := logrus.
		WithField("package", "events").
		WithField("module", "EventManager").
		WithField("function", "emitEvent").
		WithField("event.type", fmt.Sprintf("%T", e))

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	logrus.Debugf("distributing event to %d subscriptions", len(m.subscriptions))

	for _, s := range m.subscriptions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			s <- e
		}
	}
	return nil
}

func (m *manager) computeState(ctx context.Context) state {
	logrus := logrus.
		WithField("package", "events").
		WithField("module", "EventManager").
		WithField("function", "computeState")

	logrus.Debug("computing the new state...")
	defer logrus.Debug("the new state has been computed")

	subCtx, cancel := context.WithTimeout(ctx, m.connectionTimeout)
	defer cancel()

	logrus.Debugf("trying to reach the server with timeout of %s", m.connectionTimeout)
	output, err := m.sender.Send(subCtx, "players")
	if err != nil {
		logrus.Warn(err)
		return state{
			serverAvailable: false,
		}
	}

	logrus.Debug("parsing the command output...")
	players := parsers.ParsePlayers(output)

	s := state{
		serverAvailable: true,
		playersOutput:   output,
		playersByKeys:   make(map[string]parsers.Player, len(players)),
	}

	logrus.Debug("organizing the player list...")
	for _, p := range players {
		key := fmt.Sprintf("%d:%s", p.GetID(), p.GetName())
		s.playersByKeys[key] = p
	}

	return s
}

func (m *manager) Subscribe() <-chan AnyEvent {
	logrus := logrus.
		WithField("package", "events").
		WithField("module", "EventManager").
		WithField("function", "Subscribe")

	logrus.Debug("adding a subscription...")
	defer logrus.Debug("subscription added...")

	c := make(chan AnyEvent, defaultChannelBuffer)
	m.mutex.Lock()
	m.subscriptions = append(m.subscriptions, c)
	m.mutex.Unlock()

	return c
}

func (m *manager) Unsubscribe(c <-chan AnyEvent) error {
	logrus := logrus.
		WithField("package", "events").
		WithField("module", "EventManager").
		WithField("function", "Unsubscribe")

	logrus.Debug("looking for a subscription to remove...")

	newSubs := make([]chan AnyEvent, 0, len(m.subscriptions))
	m.mutex.RLock()
	for _, s := range m.subscriptions {
		if s == c {
			logrus.Debug("subscription found, closing the channel and removing...")
			close(s)
			continue
		}
		newSubs = append(newSubs, s)
	}
	m.mutex.RUnlock()

	if len(newSubs) == len(m.subscriptions) {
		logrus.Debug("subscription was not found")
		return ErrSubscriptionNotFound
	}

	m.mutex.Lock()
	m.subscriptions = newSubs
	m.mutex.Unlock()
	logrus.Debug("subscription removed successfully")
	return nil
}
