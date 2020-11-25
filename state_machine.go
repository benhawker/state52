package state52

import (
	"fmt"
	"strings"
	"sync"
)

var validglobalCallbacks = []string{"before_all_events", "after_all_events", "ensure_all_events"}
var validEventCallbacks = []string{"before", "after", "ensure"}
var validTransitionCallbacks = []string{"after", "success"}

// State52 defines your Finite State Machine.
type State52 struct {
	// initialState is the initial state.
	initialState string

	// events defines all potential events that can be performed.
	events map[string]Event

	// globalCallbacks defines callbacks that will called for all Events.
	globalCallbacks map[string]callback

	// persistFn defines a fn that will be called at the
	// appropriate moment in each event to persist the state.
	persistFn func(string) error

	// currentState represents the current state.
	currentState string

	// stateMutex locks/unlocks access to the current state.
	stateMutex sync.RWMutex

	// states holds a map of all possible states
	states map[string]struct{}
}

// Event provides the format for defining an event when creating a State Machine.
type Event struct {
	// name is the event name to be used when calling for a transition.
	Name string

	// transitions contains a slice of Transition structs.
	// Each Transition defines available transitions for an event.
	Transitions []Transition

	// guard is a function that returns a bool, if you want to provide a
	// guard for all transitions within an event. A guard is a condition
	// that must be met for the event transition's to execute.
	Guards []func() bool

	// Callbacks is a map of transition `Callback`(s) specifically
	// to be run for an event. The code refers to these as Event Callbacks.
	// callbacks map[string]callback
	Callbacks map[string]callback

	// err is an optional error that can be returned from a callback.
	err error
}

// Transition defines a transition that can be made (within an event).
type Transition struct {
	// from is a slice of `from` states that the state machine must
	// be in (i.e. CurrentState) to perform the transition.
	From []string

	// fo is the state that the state machine will be in if the transition succeds.
	To string

	// guard is a function that returns a bool, if you want to provide
	// a guard for this specific transition. A guard is a condition
	// that must be met for the transition to execute.
	Guards []func() bool

	// callbacks is a map of transition `Callback`(s) specifically run for this
	// specific transition. The code refers to these as Transition Callbacks.
	Callbacks map[string]tCallback
}

// Events -> Syntax for building the state machine
type Events []Event

// Transitions -> syntax for building the state machine
type Transitions []Transition

// Callbacks -> Syntax for building the state machine
type Callbacks map[string]callback

// callback is a function type that all Global and Event callbacks should use.
// Event is the current event data passed as the callback happens.
type callback func(*State52, *Event) error

// TransitionCallbacks -> Syntax for building the state machine
type TransitionCallbacks map[string]tCallback

// tCallback is function type that all Transaction callbacks should use.
// This allows us access to Transition & associated Event data.
type tCallback func(*State52, *Event, *Transition) error

// Guards -> Syntax for building the state machine
type Guards []func() bool

// SetInitial sets the initialState.
func SetInitial(state string) SetupFunc {
	return func(sm *State52) error {
		sm.initialState = state
		sm.currentState = state
		return nil
	}
}

// SetPersistFn sets the persistFn.
func SetPersistFn(fn func(string) error) SetupFunc {
	return func(c *State52) error {
		c.persistFn = fn
		return nil
	}
}

// SetGlobalCallbacks sets any 'global' callbacks you may seek to add.
func SetGlobalCallbacks(callbacks Callbacks) SetupFunc {
	return func(sm *State52) error {
		sm.globalCallbacks = callbacks
		return nil
	}
}

// SetEvents sets all of the events in your state machine.
func SetEvents(events Events) SetupFunc {
	return func(sm *State52) error {
		sm.events = mapEvents(events)
		return nil
	}
}

// SetupFunc is a function that configures a State52 (state machine).
type SetupFunc func(*State52) error

// NewStateMachine allows initialisation of a StateMachine.
func NewStateMachine(options ...SetupFunc) *State52 {
	sm := &State52{}

	// Apply passed options.
	for _, option := range options {
		if err := option(sm); err != nil {
			panic(err)
		}
	}

	// Always build states
	sm.states = mapStates(sm.events)

	sm.validate()
	return sm
}

func mapEvents(events []Event) map[string]Event {
	mapppedEvents := map[string]Event{}

	for _, e := range events {
		if len(e.Guards) > 0 {
			// An event level guard can specify a single guard to be applied to all transitions within an event.
			// So we append event guards to every transition within the event.
			newTransitions := []Transition{}
			for _, transition := range e.Transitions {
				transition.Guards = append(transition.Guards, e.Guards...)
				newTransitions = append(newTransitions, transition)
			}

			e.Transitions = newTransitions
		}

		mapppedEvents[e.Name] = e
	}

	return mapppedEvents
}

func mapStates(events map[string]Event) map[string]struct{} {
	allRegisteredStates := map[string]struct{}{}

	for _, event := range events {
		for _, transition := range event.Transitions {
			for _, fromState := range transition.From {
				allRegisteredStates[fromState] = struct{}{}
			}
			allRegisteredStates[transition.To] = struct{}{}
		}

	}

	return allRegisteredStates
}

func (sm *State52) validate() {
	// Validate presence of initialState
	if sm.initialState == "" {
		panic("You must set an initial state.")
	}

	// Validate the initial state is included in at least one event transition to/from.
	// Note that this checks both to & from attributes whereas it would require being present in `to` in reality.
	if _, ok := sm.states[sm.initialState]; !ok {
		panic("initialState was not found in the registered states.")
	}

	// Validate at least 1 event
	if len(sm.events) == 0 {
		panic("You must define at least 1 event.")
	}

	// Validate globalCallbacks
	for name := range sm.globalCallbacks {
		if !stringInSlice(name, validglobalCallbacks) {
			panic(fmt.Sprintf("%s is not a valid Global Callback. The following are valid: %s.", name, strings.Join(validglobalCallbacks, ",")))
		}
	}

	// Validate Event & Transition Callbacks
	for _, event := range sm.events {
		event.validate()
	}
}

func (event *Event) validate() {
	// Validates all event level callbacks.
	for callbackName := range event.Callbacks {
		if !stringInSlice(callbackName, validEventCallbacks) {
			panic(fmt.Sprintf("%s is not a valid Event Callback. The following are valid: %s.", callbackName, strings.Join(validEventCallbacks, ",")))
		}
	}

	// Validates all transition level callbacks.
	for _, transition := range event.Transitions {
		for callbackName := range transition.Callbacks {
			if !stringInSlice(callbackName, validTransitionCallbacks) {
				panic(fmt.Sprintf("%s is not a valid Transition Callback. The following are valid: %s.", callbackName, strings.Join(validTransitionCallbacks, ",")))
			}
		}
	}
}

// CurrentState returns the current state of the sm.
func (sm *State52) CurrentState() string {
	sm.stateMutex.RLock()
	defer sm.stateMutex.RUnlock()
	return sm.currentState
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
