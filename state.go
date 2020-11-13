package wilddawg

import (
	"errors"
	"hash"

	"github.com/ugorji/go/codec"
)

type StateType int

const (
	LAZYDFA StateType = iota
)

var (
	ErrEdgeAlreadyUsed = errors.New("Edge already in use in deterministic state machine")
	ErrEdgeNotPresent  = errors.New("Edge does not exist")
)

/*
	A State is a state within a finite state automaton. It has a
	method "IsomorphismHash()" which must return a hash that
	identifies its outgoing edges and destination states without
	reliance on memory addresses. "MachineEdges()" returns an edge
	map that is based on Id values rather than memory addresses.
	The "Clone()" function returns a new State with the same
	outgoing edges and destinations.
*/
type StateId int

type State interface {
	GetId() StateId
	SetId(StateId) error
	IsTerminal() bool
	SetTerminal(bool) error
	AddEdge(interface{}, State) error
	RemoveEdge(interface{}, State) error
	FollowEdge(interface{}) []State
	FollowAllEdges() []State
	MachineEdges() map[interface{}]StateId
	IsomorphismHash() (uint32, error)
	Clone() State
	GetStateType() StateType
}

// This implementation lazily provides machine edge information. It is
// a state for a deterministic finite automaton.
type LazyDfaState struct {
	Id       StateId
	Terminal bool
	Edges    map[interface{}]State
	Encoding codec.Handle
	HashFunc hash.Hash32
	Type     StateType
}

func NewLazyDfaState(id StateId, encoding codec.Handle, hashFunc hash.Hash32) *LazyDfaState {
	return &LazyDfaState{
		Id:       id,
		Encoding: encoding,
		HashFunc: hashFunc,
		Type:     LAZYDFA,
	}
}

func (s *LazyDfaState) GetId() StateId {
	return s.Id
}

func (s *LazyDfaState) SetId(id StateId) error {
	s.Id = id
	return nil
}

func (s *LazyDfaState) IsTerminal() bool {
	return s.Terminal
}

func (s *LazyDfaState) SetTerminal(terminal bool) error {
	s.Terminal = terminal
	return nil
}

func (s *LazyDfaState) AddEdge(edgeTransition interface{}, destination State) error {
	if _, present := s.Edges[edgeTransition]; present {
		return ErrEdgeAlreadyUsed
	}
	s.Edges[edgeTransition] = destination
	return nil
}

func (s *LazyDfaState) RemoveEdge(edgeTransition interface{}, destination State) error {
	if edgeTo, present := s.Edges[edgeTransition]; !present {
		return ErrEdgeNotPresent
	} else if edgeTo != destination {
		return ErrEdgeNotPresent
	}
	delete(s.Edges, edgeTransition)
	return nil
}

func (s *LazyDfaState) FollowEdge(edgeTransition interface{}) []State {
	destinationStates := make([]State, 0)
	if destination, present := s.Edges[edgeTransition]; present {
		destinationStates = append(destinationStates, destination)
	}
	return destinationStates
}

func (s *LazyDfaState) FollowAllEdges() []State {
	uniqueDestinations := make(map[State]bool)
	for _, destination := range s.Edges {
		uniqueDestinations[destination] = true
	}

	destinationStates := make([]State, 0, len(uniqueDestinations))
	for destination, _ := range uniqueDestinations {
		destinationStates = append(destinationStates, destination)
	}
	return destinationStates
}

func (s *LazyDfaState) MachineEdges() map[interface{}]StateId {
	machineEdges := make(map[interface{}]StateId)
	for edge, dest := range s.Edges {
		machineEdges[edge] = dest.GetId()
	}
	return machineEdges
}

func (s *LazyDfaState) IsomorphismHash() (uint32, error) {
	encodedBytes := make([]byte, 0, 64)
	encoder := codec.NewEncoderBytes(&encodedBytes, s.Encoding)
	if err := encoder.Encode(s.MachineEdges()); err != nil {
		return 0, err
	}
	s.HashFunc.Reset()
	s.HashFunc.Write(encodedBytes)
	return s.HashFunc.Sum32(), nil
}

func (s *LazyDfaState) Clone() State {
	clone := &LazyDfaState{
		Id:       s.Id,
		Terminal: s.Terminal,
		Edges:    make(map[interface{}]State),
		Encoding: s.Encoding,
		HashFunc: s.HashFunc,
		Type:     s.Type,
	}
	for edge, destination := range s.Edges {
		clone.Edges[edge] = destination
	}
	return clone
}

func (s *LazyDfaState) GetStateType() StateType {
	return s.Type
}
