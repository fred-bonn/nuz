package main

import "github.com/fred-bonn/nuz/nuzengine"

// event is one line of the NDJSON stream sent by /battles/start.
type event interface{ isEvent() }

// eventBase carries the "type" discriminator that the client switches on.
type eventBase struct {
	Type string `json:"type"`
}

func (eventBase) isEvent() {}

// promptEvent asks the client to choose one of Options.
type promptEvent struct {
	eventBase
	Text    string                   `json:"text,omitempty"`
	Options []string                 `json:"options,omitempty"`
	State   nuzengine.BattleSnapshot `json:"state"`
}

// hintEvent carries the 1-based number of the option currently rated best.
type hintEvent struct {
	eventBase
	Best int `json:"best"`
}

type endEvent struct {
	eventBase
	Won   bool                     `json:"won"`
	State nuzengine.BattleSnapshot `json:"state"`
}

// errorEvent is sent when the battle fails; timeoutEvent when the client stays silent too long.
type errorEvent struct {
	eventBase
	Error string `json:"error"`
}

type timeoutEvent struct {
	eventBase
	Error string `json:"error"`
}

func newPromptEvent(text string, options []string, state nuzengine.BattleSnapshot) promptEvent {
	return promptEvent{eventBase{"prompt"}, text, options, state}
}

func newHintEvent(best int) hintEvent { return hintEvent{eventBase{"hint"}, best} }

func newEndEvent(won bool, state nuzengine.BattleSnapshot) endEvent {
	return endEvent{eventBase{"end"}, won, state}
}

func newErrorEvent(err error) errorEvent { return errorEvent{eventBase{"error"}, err.Error()} }

func newTimeoutEvent() timeoutEvent {
	return timeoutEvent{eventBase{"timeout"}, "no input received in time"}
}
