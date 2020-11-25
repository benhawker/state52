package state52

import (
	"fmt"
	"reflect"
)

// Event performs the first available transition that is found.
func (sm *State52) Event(event string, args ...interface{}) error {
	selectedEvent, ok := sm.events[event]
	if !ok {
		return EventNotRegisteredError{event}
	}

	// defer (i.e. ensure) that any ensure_on_all_events callback will be called.
	defer func() {
		sm.ensureEventCallback(&selectedEvent)
		sm.ensureAllEventsCallback(&selectedEvent)
	}()

	err := sm.beforeAllEventsCallback(&selectedEvent)
	if err != nil {
		return err
	}

	err = sm.beforeEventCallback(&selectedEvent)
	if err != nil {
		return err
	}

	selectedTransition := Transition{}
	for _, transition := range selectedEvent.Transitions {
		// If the 'from' states do not include the CurrentState
		// we continue to next iteration.
		if !stringInSlice(sm.CurrentState(), transition.From) {
			continue
		}

		// If there is no guard we select this transition
		if len(transition.Guards) == 0 { // No Guards not defined
			selectedTransition = transition
			break
		} else {
			guardsResult := false

			for _, guard := range transition.Guards {
				if guard() == true { // Guard defined
					guardsResult = true
				} else {
					guardsResult = false
					break
				}
			}

			if guardsResult == true {
				selectedTransition = transition
				break
			}
		}
	}

	// If we could not select a transition to execute we
	// return a CannotTransitionError
	if reflect.DeepEqual(selectedTransition, Transition{}) {
		return CannotTransitionError{sm.CurrentState(), event}
	}

	// Transition after
	sm.afterTransitionCallback(selectedTransition, &selectedEvent)

	// Perform the transition
	sm.setCurrentState(selectedTransition.To)

	// Call the persistFn if it has been passed
	if sm.persistFn != nil {
		err = sm.persistFn(selectedTransition.To)
		if err != nil {
			return PersistFailedError{err, event}
		}
	}

	// Transition success
	sm.successTransitionCallback(selectedTransition, &selectedEvent)

	sm.afterEventCallback(&selectedEvent)
	sm.afterAllEventsCallback(&selectedEvent)

	return selectedEvent.err
}

func (sm *State52) setCurrentState(state string) {
	sm.stateMutex.Lock()
	sm.currentState = state
	sm.stateMutex.Unlock()
}

// beforeEventCallback
func (sm *State52) beforeEventCallback(e *Event) error {
	if fn, ok := e.Callbacks["before"]; ok {
		err := fn(sm, e)
		if err != nil {
			return err
		}
	}
	return nil
}

// beforeAllEventsCallback
func (sm *State52) beforeAllEventsCallback(e *Event) error {
	if fn, ok := sm.globalCallbacks["before_all_events"]; ok {
		err := fn(sm, e)
		if err != nil {
			return err
		}
	}
	return nil
}

// afterTransitionCallback
func (sm *State52) afterTransitionCallback(t Transition, e *Event) error {
	if fn, ok := t.Callbacks["after"]; ok {
		err := fn(sm, e, &t)
		if err != nil {
			return err
		}
	}
	return nil
}

// successTransitionCallback
func (sm *State52) successTransitionCallback(t Transition, e *Event) error {
	if fn, ok := t.Callbacks["success"]; ok {
		err := fn(sm, e, &t)
		if err != nil {
			return err
		}
	}
	return nil
}

// afterEventCallback
func (sm *State52) afterEventCallback(e *Event) error {
	if fn, ok := e.Callbacks["after"]; ok {
		err := fn(sm, e)
		if err != nil {
			return err
		}
	}
	return nil
}

// ensureEventCallback
func (sm *State52) ensureEventCallback(e *Event) error {
	if fn, ok := e.Callbacks["ensure"]; ok {
		err := fn(sm, e)
		if err != nil {
			return err
		}
	}
	return nil
}

// afterAllEventsCallback
func (sm *State52) afterAllEventsCallback(e *Event) error {
	if fn, ok := sm.globalCallbacks["after_all_events"]; ok {
		err := fn(sm, e)
		if err != nil {
			return err
		}
	}
	return nil
}

// ensureAllEventsCallback
func (sm *State52) ensureAllEventsCallback(e *Event) error {
	if fn, ok := sm.globalCallbacks["ensure_all_events"]; ok {
		err := fn(sm, e)
		if err != nil {
			return err
		}
	}
	return nil
}

// PersistFailedError when the persistFn provided returns an error
type PersistFailedError struct {
	Message   error
	EventName string
}

func (e PersistFailedError) Error() string {
	return fmt.Sprintf("Perist failed for %s: %s", e.EventName, e.Message)
}

// CannotTransitionError will be returned when calling Event()
// with a CurrentState that cannot be transitioned from.
type CannotTransitionError struct {
	CurrentState string
	EventName    string
}

func (e CannotTransitionError) Error() string {
	return fmt.Sprintf("Cannot transition from %s when calling %s.", e.CurrentState, e.EventName)
}

// EventNotRegisteredError will be returned when calling Event()
// with an event name that is not registered.
type EventNotRegisteredError struct {
	EventName string
}

func (e EventNotRegisteredError) Error() string {
	return fmt.Sprintf("%s is not registered.", e.EventName)
}
