package state52_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/benhawker/state52"
)

func TestSingleEvent(t *testing.T) {
	events := state52.Events{
		{
			Name: "first_event",
			Transitions: state52.Transitions{
				{From: []string{"start"}, To: "succeeded_first"},
			},
		},
	}

	sm := state52.NewStateMachine(
		state52.SetInitial("start"),
		state52.SetEvents(events),
	)

	err := sm.Event("first_event")
	if err != nil {
		t.Errorf("expected error message to be: nil, got %s", err.Error())
	}

	if sm.CurrentState() != "succeeded_first" {
		t.Errorf("expected state to be 'succeeded_first', got %s", sm.CurrentState())
	}
}

func TestNoInitialState(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic due to no initialState, but no panic was thrown.")
		}
	}()

	state52.NewStateMachine()
}

func TestNoEvents(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic due to no events, but no panic was thrown.")
		}
	}()

	state52.NewStateMachine(
		state52.SetInitial("start"),
	)
}

func TestInvalidGlobalCallback(t *testing.T) {
	events := state52.Events{
		{
			Name: "first_event",
			Transitions: state52.Transitions{
				{From: []string{"start"}, To: "succeeded_first"},
			},
		},
	}

	globalCallbacks := state52.Callbacks{
		"not_a_global_callback": func(sm *state52.State52, e *state52.Event) error {
			return nil
		},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic due to an invalid global callback, but no panic was thrown.")
		}
	}()

	state52.NewStateMachine(
		state52.SetInitial("start"),
		state52.SetEvents(events),
		state52.SetGlobalCallbacks(globalCallbacks),
	)
}

func TestInvalidEventCallback(t *testing.T) {
	events := state52.Events{
		{
			Name: "first_event",
			Transitions: state52.Transitions{
				{From: []string{"start"}, To: "succeeded_first"},
			},
			Callbacks: state52.Callbacks{
				"not_an_event_callback": func(sm *state52.State52, e *state52.Event) error {
					return nil
				},
			},
		},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic due to no invalid event callback, but no panic was thrown.")
		}
	}()

	state52.NewStateMachine(
		state52.SetInitial("start"),
		state52.SetEvents(events),
	)
}

func TestInvalidTrantionCallback(t *testing.T) {
	events := state52.Events{
		{
			Name: "first_event",
			Transitions: state52.Transitions{
				{
					From: []string{"start"},
					To:   "succeeded_first",
					Callbacks: state52.TransitionCallbacks{
						"not_a_transition_callback": func(sm *state52.State52, e *state52.Event, t *state52.Transition) error {
							return nil
						},
					},
				},
			},
		},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic due to no invalid transition callback, but no panic was thrown.")
		}
	}()

	state52.NewStateMachine(
		state52.SetInitial("start"),
		state52.SetEvents(events),
	)
}

func TestAllCallbacks(t *testing.T) {
	beforeEvent := false
	afterEvent := false

	afterTransition := false
	successTransition := false

	beforeAllEvents := false
	afterAllEvents := false

	ensureEvent := false
	ensureAllEvents := false

	persistFnValue := false

	events := state52.Events{
		{
			Name: "first_event",
			Transitions: state52.Transitions{
				{From: []string{"start"}, To: "succeeded_first",
					Callbacks: state52.TransitionCallbacks{
						"after": func(sm *state52.State52, e *state52.Event, t *state52.Transition) error {
							afterTransition = true
							return nil
						},
						"success": func(sm *state52.State52, e *state52.Event, t *state52.Transition) error {
							successTransition = true
							return nil
						},
					},
				},
			},
			Callbacks: state52.Callbacks{
				"before": func(sm *state52.State52, e *state52.Event) error {
					beforeEvent = true
					return nil
				},
				"after": func(sm *state52.State52, e *state52.Event) error {
					afterEvent = true
					return nil
				},
				"ensure": func(sm *state52.State52, e *state52.Event) error {
					ensureEvent = true
					return nil
				},
			},
		},
		{
			Name: "second_event",
			Transitions: state52.Transitions{
				{From: []string{"succeeded_first"}, To: "succeeded_second"},
			},
		},
		{
			Name: "third_event",
			Transitions: state52.Transitions{
				{From: []string{"succeeded_second"}, To: "completed"},
			},
		},
	}

	globalCallbacks := state52.Callbacks{
		"before_all_events": func(sm *state52.State52, e *state52.Event) error {
			beforeAllEvents = true
			return nil
		},
		"after_all_events": func(sm *state52.State52, e *state52.Event) error {
			afterAllEvents = true
			return nil
		},
		"ensure_all_events": func(sm *state52.State52, e *state52.Event) error {
			ensureAllEvents = true
			return nil
		},
	}

	persistFn := func(newState string) error {
		persistFnValue = true
		return nil
	}

	sm := state52.NewStateMachine(
		state52.SetInitial("start"),
		state52.SetEvents(events),
		state52.SetGlobalCallbacks(globalCallbacks),
		state52.SetPersistFn(persistFn),
	)

	err := sm.Event("first_event")
	if err != nil {
		t.Errorf("expected error message to be: nil, got %s", err.Error())
	}

	if sm.CurrentState() != "succeeded_first" {
		t.Errorf("expected state to be 'succeeded_first', got %s", sm.CurrentState())
	}

	if !(beforeEvent && afterEvent) {
		t.Errorf("expected all callbacks to be called, got: beforeEvent: %t, afterEvent: %t",
			beforeEvent, afterEvent)
	}

	if !(beforeAllEvents && afterAllEvents) {
		t.Errorf("expected all callbacks to be called, got: beforeAllEvents: %t, afterAllEvents: %t",
			beforeAllEvents, afterAllEvents)
	}

	if !(afterTransition && successTransition) {
		t.Errorf("expected all callbacks to be called, got: afterTransition: %t, successTransition: %t",
			afterTransition, successTransition)
	}

	if !persistFnValue {
		t.Errorf("expected persistFn to be called: expected: false, got: %t", persistFnValue)
	}

	if !(ensureEvent) {
		t.Errorf("expected all callbacks to be called, got: ensureEvents: %t", ensureEvent)
	}

	if !(ensureAllEvents) {
		t.Errorf("expected all callbacks to be called, got: ensureAllEvents: %t", ensureAllEvents)
	}

	sm.Event("second_event")
	if sm.CurrentState() != "succeeded_second" {
		t.Errorf("expected state to be 'succeeded_second', got %s", sm.CurrentState())
	}

	sm.Event("third_event")
	if sm.CurrentState() != "completed" {
		t.Errorf("expected state to be 'completed', got %s", sm.CurrentState())
	}
}

func TestEventGuards(t *testing.T) {
	events := state52.Events{
		{
			Name:   "first_event",
			Guards: state52.Guards{fnThatReturnsTrue, fnThatReturnsTrue},
			Transitions: state52.Transitions{
				{From: []string{"start", "another_value"}, To: "special_case", Guards: state52.Guards{fnThatReturnsFalse, fnThatReturnsTrue}},
				{From: []string{"start"}, To: "succeeded_first", Guards: state52.Guards{fnThatReturnsTrue}},
			},
		},
		{
			Name:   "second_event",
			Guards: state52.Guards{fnThatReturnsTrue},
			Transitions: state52.Transitions{
				{From: []string{"succeeded_first"}, To: "failed_second", Guards: state52.Guards{fnThatReturnsFalse}},
				{From: []string{"succeeded_first"}, To: "succeeded_second", Guards: state52.Guards{fnThatReturnsTrue}},
			},
		},
		{
			Name:   "third_event",
			Guards: state52.Guards{fnThatReturnsFalse},
			Transitions: state52.Transitions{
				{From: []string{"succeeded_second"}, To: "declined", Guards: state52.Guards{fnThatReturnsFalse}},
				{From: []string{"succeeded_second"}, To: "completed", Guards: state52.Guards{fnThatReturnsTrue}},
			},
		},
	}

	sm := state52.NewStateMachine(
		state52.SetInitial("start"),
		state52.SetEvents(events),
	)

	err := sm.Event("first_event")
	if err != nil {
		t.Errorf("expected error message to be: nil, got %s", err.Error())
	}

	if sm.CurrentState() != "succeeded_first" {
		t.Errorf("expected state to be 'succeeded_first', got %s", sm.CurrentState())
	}

	err = sm.Event("second_event")
	if err != nil {
		t.Errorf("expected error message to be: nil, got %s", err.Error())
	}

	if sm.CurrentState() != "succeeded_second" {
		t.Errorf("expected state to be 'succeeded_second', got %s", sm.CurrentState())
	}

	err = sm.Event("third_event")
	expectedErrorMessage := "Cannot transition from succeeded_second when calling third_event."
	if err.Error() != expectedErrorMessage {
		t.Errorf("expected error message to be: %s, got %s", expectedErrorMessage, err.Error())
	}

	if sm.CurrentState() != "succeeded_second" {
		t.Errorf("expected state to be 'succeeded_second', got %s", sm.CurrentState())
	}
}

func TestUnregisteredEventError(t *testing.T) {
	sm := state52.NewStateMachine(
		state52.SetInitial("start"),
		state52.SetEvents(
			state52.Events{
				{
					Name: "first_event",
					Transitions: state52.Transitions{
						{From: []string{"start"}, To: "succeeded_first"},
					},
				},
			},
		),
	)

	err := sm.Event("not_defined_event")
	expectedErrorMessage := "not_defined_event is not registered."
	if err.Error() != expectedErrorMessage {
		t.Errorf("expected error message to be: %s, got %s", expectedErrorMessage, err.Error())
	}

	if sm.CurrentState() != "start" {
		t.Errorf("expected state to be 'start', got %s", sm.CurrentState())
	}
}

func TestTransitionGuards(t *testing.T) {
	sm := state52.NewStateMachine(
		state52.SetInitial("start"),
		state52.SetEvents(
			state52.Events{
				{
					Name: "first_event",
					Transitions: state52.Transitions{
						{From: []string{"start", "another_value"}, To: "special_case", Guards: state52.Guards{fnThatReturnsFalse, fnThatReturnsTrue}},
						{From: []string{"start"}, To: "succeeded_first", Guards: state52.Guards{fnThatReturnsTrue}},
					},
				},
				{
					Name: "second_event",
					Transitions: state52.Transitions{
						{From: []string{"succeeded_first"}, To: "failed_second", Guards: state52.Guards{fnThatReturnsFalse}},
						{From: []string{"succeeded_first"}, To: "succeeded_second", Guards: state52.Guards{fnThatReturnsTrue}},
					},
				},
				{
					Name: "third_event",
					Transitions: state52.Transitions{
						{From: []string{"succeeded_second"}, To: "declined", Guards: state52.Guards{fnThatReturnsFalse}},
						{From: []string{"succeeded_second"}, To: "completed", Guards: state52.Guards{fnThatReturnsTrue}},
					},
				},
			},
		),
	)

	err := sm.Event("first_event")
	if err != nil {
		t.Errorf("expected error message to be: nil, got %s", err.Error())
	}
	if sm.CurrentState() != "succeeded_first" {
		t.Errorf("expected state to be 'succeeded_first', got %s", sm.CurrentState())
	}

	err = sm.Event("second_event")
	if err != nil {
		t.Errorf("expected error message to be: nil, got %s", err.Error())
	}
	if sm.CurrentState() != "succeeded_second" {
		t.Errorf("expected state to be 'succeeded_second', got %s", sm.CurrentState())
	}

	err = sm.Event("third_event")
	if err != nil {
		t.Errorf("expected error message to be: nil, got %s", err.Error())
	}
	if sm.CurrentState() != "completed" {
		t.Errorf("expected state to be 'completed', got %s", sm.CurrentState())
	}
}

func TestCallingEventFnWithinCallback(t *testing.T) {
	sm := state52.NewStateMachine(
		state52.SetInitial("start"),
		state52.SetEvents(
			state52.Events{
				{
					Name: "first_event",
					Transitions: state52.Transitions{
						{From: []string{"start"}, To: "succeeded_first"},
					},
					Callbacks: state52.Callbacks{
						"after": func(sm *state52.State52, e *state52.Event) error {
							err := sm.Event("second_event")
							if err != nil {
								return err
							}
							return nil
						},
					},
				},
				{
					Name: "second_event",
					Transitions: state52.Transitions{
						{From: []string{"succeeded_first", "failed_second"}, To: "succeeded_second"},
					},
					Callbacks: state52.Callbacks{
						"after": func(sm *state52.State52, e *state52.Event) error {
							err := sm.Event("third_event")
							if err != nil {
								return err
							}
							return nil
						},
					},
				},
				{
					Name: "third_event",
					Transitions: state52.Transitions{
						{From: []string{"succeeded_second"}, To: "completed"},
					},
				},
			},
		),
	)

	err := sm.Event("first_event")
	if err != nil {
		t.Errorf("expected error message to be: nil, got %s", err.Error())
	}
	if sm.CurrentState() != "completed" {
		t.Errorf("expected state to be 'completed', got %s", sm.CurrentState())
	}
}

func TestEventNotRegisteredError(t *testing.T) {
	eventName := "not_an_event"

	e := state52.EventNotRegisteredError{eventName}
	if e.Error() != fmt.Sprintf("%s is not registered.", e.EventName) {
		t.Errorf("Expected %s, Got: %s", fmt.Sprintf("%s is not registered.", e.EventName), e.Error())
	}
}

func TestCannotTransitionError(t *testing.T) {
	eventName := "not_an_event"
	currentState := "initial"

	e := state52.CannotTransitionError{currentState, eventName}
	if e.Error() != fmt.Sprintf("Cannot transition from %s when calling %s.", e.CurrentState, e.EventName) {
		t.Errorf("Expected %s, Got: %s", fmt.Sprintf("Cannot transition from %s when calling %s.", e.CurrentState, e.EventName), e.Error())
	}
}

func TestPersistFailedError(t *testing.T) {
	eventName := "event"
	message := errors.New("something broke")

	e := state52.PersistFailedError{message, eventName}
	if e.Error() != fmt.Sprintf("Perist failed for %s: %s", e.EventName, e.Message) {
		t.Errorf("Expected %s, Got: %s", fmt.Sprintf("Perist failed for %s: %s", e.EventName, e.Message), e.Error())
	}
}

func fnThatReturnsTrue() bool {
	return true
}

func fnThatReturnsFalse() bool {
	return false
}
