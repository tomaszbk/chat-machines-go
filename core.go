package chatmachine

import (
	"strings"
)

// State is the interface for defining states in the state machine.
type State interface {
	OnEnter(session *SessionState)
	OnUpdate(session *SessionState)
	OnExit(session *SessionState)
}

// ChangeState is used to signal a state change via panic/recover.
type ChangeState struct{}

// SessionState holds the state of a chat session.
type SessionState struct {
	Input        string
	Output       string
	NextState    State
	Data         interface{}
	endSession   bool
	currentState State
}

// AddOutput appends output text to the session's output buffer.
func (s *SessionState) AddOutput(output string) {
	trimmed := strings.TrimSpace(output)
	s.Output += trimmed + "\n"
}

// ChangeState sets the next state and panics to signal a state change.
func (s *SessionState) ChangeState(state State) {
	s.NextState = state
	panic(ChangeState{})
}

// End marks the session for termination.
func (s *SessionState) End() {
	s.endSession = true
}

// ChatMachine is a generic state machine managing sessions.
type ChatMachine struct {
	sessions        map[string]*SessionState
	startState      State
	globalOnEnter   func(*SessionState)
	globalOnUpdate  func(*SessionState)
	globalOnExit    func(*SessionState)
}

// NewChatMachine creates a new ChatMachine with the given start state.
func NewChatMachine(start State) *ChatMachine {
	return &ChatMachine{
		sessions:       make(map[string]*SessionState),
		startState:     start,
		globalOnEnter:  func(*SessionState) {},
		globalOnUpdate: func(*SessionState) {},
		globalOnExit:   func(*SessionState) {},
	}
}

// GlobalOnEnter registers a hook to run on entering any state.
func (cm *ChatMachine) GlobalOnEnter(fn func(*SessionState)) {
	cm.globalOnEnter = fn
}

// GlobalOnUpdate registers a hook to run on state updates.
func (cm *ChatMachine) GlobalOnUpdate(fn func(*SessionState)) {
	cm.globalOnUpdate = fn
}

// GlobalOnExit registers a hook to run on exiting any state.
func (cm *ChatMachine) GlobalOnExit(fn func(*SessionState)) {
	cm.globalOnExit = fn
}

// getSession retrieves or creates a session by ID.
func (cm *ChatMachine) getSession(sessionID string) *SessionState {
	if s, ok := cm.sessions[sessionID]; ok {
		return s
	}
	s := &SessionState{}
	cm.sessions[sessionID] = s
	return s
}

// Run processes input for a session and returns the output.
func (cm *ChatMachine) Run(input, sessionID string) string {
	s := cm.getSession(sessionID)
	s.Input = input
	s.Output = ""
	var changed bool
	// Execute state logic and catch ChangeState signal
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(ChangeState); ok {
					changed = true
				} else {
					panic(r)
				}
			}
		}()
		if s.currentState == nil {
			// Initial state entry
			s.currentState = cm.startState
			cm.globalOnEnter(s)
			s.currentState.OnEnter(s)
		} else {
			// State update
			cm.globalOnUpdate(s)
			s.currentState.OnUpdate(s)
		}
	}()
	if changed {
		// Handle state transition
		s.currentState.OnExit(s)
		s.currentState = s.NextState
		s.NextState = nil
		cm.globalOnEnter(s)
		s.currentState.OnEnter(s)
	}
	output := s.Output
	if s.endSession {
		delete(cm.sessions, sessionID)
	}
	return output
}
