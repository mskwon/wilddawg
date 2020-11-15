package wilddawg

import (
	"errors"
)

type RegisterType int

const (
	COLLISIONSAFEHASHMAP RegisterType = iota
)

var (
	ErrRegisterNilState  = errors.New("Nil state passed to register")
	ErrNonMinimalMachine = errors.New("Start state passed to register " +
		"is part of a non-minimal state machine")
)

/*
	The register keeps track of the equivalence classes of the machine. It
	should be able to initialize itself based on some start state of a minimized
	DAWG.
*/
type Register interface {
	GetEquivalenceClass(State) (State, error)
	Initialize(State) error
	Reset() error
	GetRegisterType() RegisterType
}

// This implementation of Register stores equivalence classes using maps of
// IsomorphismHashes to lists of State pointers. It allows for the possibility
// of hash collisions.
type CollisionSafeHashMapRegister struct {
	EquivalenceClassMap map[interface{}][]State
	Type                RegisterType
}

func NewCollisionSafeHashMapRegister() *CollisionSafeHashMapRegister {
	return &CollisionSafeHashMapRegister{
		EquivalenceClassMap: make(map[interface{}][]State),
		Type:                COLLISIONSAFEHASHMAP,
	}
}

func (r *CollisionSafeHashMapRegister) GetEquivalenceClass(queryState State) (
	State, error) {
	if queryState == nil {
		return nil, ErrRegisterNilState
	}
	if hash, err := queryState.IsomorphismHash(); err != nil {
		return nil, err
	} else if stateRef, present := r.EquivalenceClassMap[hash]; !present {
		r.EquivalenceClassMap[hash] = []State{queryState}
		return queryState, nil
	} else {
		queryMachineEdges := queryState.MachineEdges()
		for _, state := range stateRef {
			if sameMachineEdges(queryMachineEdges, state.MachineEdges()) {
				return state, nil
			}
		}
		r.EquivalenceClassMap[hash] = append(r.EquivalenceClassMap[hash],
			queryState)
		return queryState, nil
	}
}

func (r *CollisionSafeHashMapRegister) Reset() error {
	r.EquivalenceClassMap = make(map[interface{}][]State)
	return nil
}

func (r *CollisionSafeHashMapRegister) Initialize(startState State) error {
	if err := r.Reset(); err != nil {
		return err
	}
	if startState == nil {
		return ErrRegisterNilState
	}

	seenStates := map[StateId]bool{startState.GetId(): true}
	stack := []State{startState}
	for len(stack) != 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if ref, err := r.GetEquivalenceClass(curr); err != nil {
			return err
		} else if curr.GetId() != ref.GetId() {
			return ErrNonMinimalMachine
		}

		for _, next := range curr.FollowAllEdges() {
			nextId := next.GetId()
			if _, seen := seenStates[nextId]; !seen {
				stack = append(stack, next)
				seenStates[nextId] = true
			}
		}
	}

	return nil
}

func (r *CollisionSafeHashMapRegister) GetRegisterType() RegisterType {
	return r.Type
}
