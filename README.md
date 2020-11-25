# State52

State52 is a library for composing Finite State Machines in a Go program.

## Usage

Initialize your state machine by setting your **initialState** & **events**. 

Note that all state machines **must** set the initialState and at least 1 event.

```go
sm := NewSM(
    state52.SetInitial("start"),
    state52.SetEvents(
        state52.Events{
            {
                Name: "first_event",
                Transitions: state52.Transitions{
                    {from: []string{"start"}, to: "succeeded_first"},
                },
            },
            {
                Name: "second_event",
                Transitions: state52.Transitions{
                    {from: []string{"succeeded_first"}, to: "succeeded_second"},
                },
            },
        },
    ),
)

// You can then call each event using the name....
err := sm.Event("first_event")
if err != nil {
    // Something went wrong
}

// Your state machine (sm) has only 1 other publicly exposed method.
// This returns your current state.
state := sm.CurrentState()
```


When defining the state machine, you can optionally add **globalCallbacks** and a **persistFn**:
```go
sm := state52.NewStateMachine(
    state52.SetInitial("start"),
    state52.SetEvents(
        state52.Events{
            {
                Name: "first_event",
                Transitions: state52.Transitions{
                    {from: []string{"start"}, to: "succeeded_first"},
                },
            },
            {
                Name: "second_event",
                Transitions: state52.Transitions{
                    {from: []string{"succeeded_first"}, to: "succeeded_second"},
                },
            },
        },
    ),
    SetGlobalCallbacks(
        state52.Callbacks{
            "before_all_events": func(sm *state52.State52, e *state52.Event) error {
                // Do stuff
                return nil
            },
            "ensure_on_all_events": func(sm *state52.State52, e *state52.Event) error {
                // Do stuff
                return nil
            },
        },
    ),
    SetPersistFn(
        func(newState string) error {
            // Do stuff
            return nil
        }
    ),
)
```

`Event`(s), `Transition`(s) also have defined callbacks.

An event callback fn must have the following signature:
```go
type callback func(*State52, *Event) error
```

A transition callback must have the following signature:
```go
type tCallback func(*State52, *Event, *Transition) error
```

```go
sm := state52.NewStateMachine(
    state52.SetInitial("start"),
    state52.SetEvents(
        state52.Events{
            {
                Name: "first_event",
                Transitions: state52.Transitions{
                    {from: []string{"start"}, to: "succeeded_first",
                        callbacks: Transitionstate52.Callbacks{ // Transition level callbacks.
                            "after": func(sm *state52.State52, e *state52.Event, t *state52.Transition) error {
                                // Do stuff
                                return nil
                            },
                            "success": func(sm *state52.State52, e *state52.Event, t *state52.Transition) error {
                                // Do stuff
                                return nil
                            },
                        },
                    },
                },
                // Event level callbacks - they will be applied to every transition in "first_event"
                callbacks: state52.Callbacks{ 
                    "before": func(sm *state52.State52, e *state52.Event) error {
                        // Do stuff
                        return nil
                    },
                    "after": func(sm *state52.State52, e *state52.Event) error {
                        // Do stuff
                        return nil
                    },
                    "ensure": func(sm *state52.State52, e *state52.Event) error {
                        // Do stuff
                        return nil
                    },
                },
            },
        }        
    ),
)
```

They can also both have `guards`.

A guard is useful if you want to allow (a) particular transition(s) only if a condition is given.
You can set up guards for each transition, which will be run before executing the transition. All guards must return true for transition to proceed.

An **event-level guard** allows you to specify a guards that will be applied to all transitions within an event. Each guard is a fn that returns a boolean. The signature:
```go
type Guards []func() bool
```

```go
sm := state52.NewStateMachine(
    state52.SetInitial("start"),
    state52.SetEvents(
        state52.Events{
            {
                Name:   "first_event",
                Guards: state52.Guards{fnThatReturnsTrue, fnThatReturnsTrue},
                Transitions: state52.Transitions{
                    {from: []string{"start", "failed_first"}, to: "failed_first", Guards: state52.Guards{fnThatReturnsFalse}},
                    {from: []string{"start", "failed_first"}, to: "succeeded_first", Guards: state52.Guards{fnThatReturnsTrue}},
                },
            },
            {
                Name:   "second_event",
                Guards: state52.Guards{fnThatReturnsTrue},
                Transitions: state52.Transitions{
                    {from: []string{"succeeded_first", "failed_second"}, to: "failed_second", Guards: state52.Guards{fnThatReturnsFalse}},
                    {from: []string{"succeeded_first", "failed_second"}, to: "succeeded_second", Guards: state52.Guards{fnThatReturnsTrue}},
                },
            },
            {
                Name:   "third_event",
                Guards: state52.Guards{fnThatReturnsFalse},
                Transitions: state52.Transitions{
                    {from: []string{"succeeded_second"}, to: "declined", Guards: state52.Guards{fnThatReturnsFalse}},
                    {from: []string{"succeeded_second"}, to: "completed", Guards: state52.Guards{fnThatReturnsTrue}},
                },
            },
        }
    ),
)
```

You can trigger the next event as part of a callback like so:
```go
sm := state52.NewStateMachine(
    state52.SetInitial("start"),
    state52.SetEvents(
        state52.Events{
            {
                Name: "first_event",
                Transitions: state52.Transitions{
                    {from: []string{"start"}, to: "succeeded_first"},
                },
                callbacks: state52.Callbacks{
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
                    {from: []string{"succeeded_first"}, to: "succeeded_second"},
                },
                callbacks: state52.Callbacks{
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
                    {from: []string{"succeeded_second"}, to: "completed"},
                },
            },
        },
    ),
)
```

### Callbacks

The list belows shows the supported callbacks & to which concept they are associated with.
```
event       before_all_events
event       before
------      (Transition is selected)
transition  guards
transition  after
------      (New state set)
------      **`persistFn`** called
transition  success
event       after
event       after_all_events
event       ensure
event       ensure_on_all_events
```
