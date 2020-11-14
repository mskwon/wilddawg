package wilddawg

import (
	"errors"
	"hash"

	"github.com/ugorji/go/codec"
)

type StateType int

const (
	LAZYDFASTATEFUL StateType = iota
)

var (
	ErrEdgeAlreadyUsed   = errors.New("Edge already in use in deterministic state machine")
	ErrEdgeNotPresent    = errors.New("Edge does not exist")
	ErrAnnotationInvalid = errors.New("Invalid annotation")
	ErrNotImplemented    = errors.New("Not Implemented")
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
	AddAnnotation(interface{}) error
	RemoveAnnotation(interface{}) error
	GetAnnotations() ([]interface{}, error)
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
// a state for a deterministic finite automaton that also holds state
// information.
type LazyDfaStatefulState struct {
	Id                      StateId
	Terminal                bool
	Edges                   map[interface{}]State
	Encoding                codec.Handle
	HashFunc                hash.Hash32
	Annotations             map[interface{}]bool
	AddAnnotationHandler    func(interface{}) error
	RemoveAnnotationHandler func(interface{}) error
	GetAnnotationsHandler   func() interface{}
	Type                    StateType
}

func NewLazyDfaState(id StateId, encoding codec.Handle, hashFunc hash.Hash32) *LazyDfaStatefulState {
	return &LazyDfaStatefulState{
		Id:          id,
		Edges:       make(map[interface{}]State),
		Encoding:    encoding,
		HashFunc:    hashFunc,
		Type:        LAZYDFASTATEFUL,
		Annotations: make(map[interface{}]bool),
	}
}

func (s *LazyDfaStatefulState) GetId() StateId {
	return s.Id
}

func (s *LazyDfaStatefulState) SetId(id StateId) error {
	s.Id = id
	return nil
}

func (s *LazyDfaStatefulState) IsTerminal() bool {
	return s.Terminal
}

func (s *LazyDfaStatefulState) SetTerminal(terminal bool) error {
	s.Terminal = terminal
	return nil
}

func (s *LazyDfaStatefulState) AddAnnotation(annotation interface{}) error {
	s.Annotations[annotation] = true
	return nil
}

func (s *LazyDfaStatefulState) RemoveAnnotation(annotation interface{}) error {
	if _, present := s.Annotations[annotation]; !present {
		return ErrAnnotationInvalid
	}
	delete(s.Annotations, annotation)
	return nil
}

func (s *LazyDfaStatefulState) GetAnnotations() ([]interface{}, error) {
	annotationList := make([]interface{}, 0, len(s.Annotations))
	for annotation, _ := range s.Annotations {
		annotationList = append(annotationList, annotation)
	}
	return annotationList, nil
}

func (s *LazyDfaStatefulState) AddEdge(edgeTransition interface{}, destination State) error {
	if _, present := s.Edges[edgeTransition]; present {
		return ErrEdgeAlreadyUsed
	}
	s.Edges[edgeTransition] = destination
	return nil
}

func (s *LazyDfaStatefulState) RemoveEdge(edgeTransition interface{}, destination State) error {
	if edgeTo, present := s.Edges[edgeTransition]; !present {
		return ErrEdgeNotPresent
	} else if edgeTo != destination {
		return ErrEdgeNotPresent
	}
	delete(s.Edges, edgeTransition)
	return nil
}

func (s *LazyDfaStatefulState) FollowEdge(edgeTransition interface{}) []State {
	destinationStates := make([]State, 0)
	if destination, present := s.Edges[edgeTransition]; present {
		destinationStates = append(destinationStates, destination)
	}
	return destinationStates
}

func (s *LazyDfaStatefulState) FollowAllEdges() []State {
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

func (s *LazyDfaStatefulState) MachineEdges() map[interface{}]StateId {
	machineEdges := make(map[interface{}]StateId)
	for edge, dest := range s.Edges {
		machineEdges[edge] = dest.GetId()
	}
	return machineEdges
}

func (s *LazyDfaStatefulState) IsomorphismHash() (uint32, error) {
	encodedBytes := make([]byte, 0, 64)
	encoder := codec.NewEncoderBytes(&encodedBytes, s.Encoding)
	if err := encoder.Encode(s.MachineEdges()); err != nil {
		return 0, err
	}
	s.HashFunc.Reset()
	s.HashFunc.Write(encodedBytes)
	return s.HashFunc.Sum32(), nil
}

func (s *LazyDfaStatefulState) Clone() State {
	clone := &LazyDfaStatefulState{
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

func (s *LazyDfaStatefulState) GetStateType() StateType {
	return s.Type
}
