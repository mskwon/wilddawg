package wilddawg

import (
	"errors"
	"hash"

	"github.com/ugorji/go/codec"
)

var (
	ErrInvalidStateType = errors.New("Invalid StateType")
)

/*
	A StateFactory handles initialization and Id handling of States.
*/
type StateFactory interface {
	GetIdCounter() StateId
	SetIdCounter(StateId) error
	GetDefaultStateType() StateType
	SetDefaultStateType(StateType) error
	NewState() (State, error)
	CloneState(State) (State, error)
}

// This implementation is a state factory that can initialize States that need
// an encoding and hashing function.
type EncodeHashStateFactory struct {
	IdCounter        StateId
	Encoding         codec.Handle
	HashFunc         hash.Hash32
	DefaultStateType StateType
}

func NewEncodeHashStateFactory(encoding codec.Handle, hashFunc hash.Hash32,
	defaultStateType StateType) (*EncodeHashStateFactory, error) {
	switch defaultStateType {
	case LAZYDFAANNOTATED:
		break
	default:
		return nil, ErrInvalidStateType
	}

	newFactory := &EncodeHashStateFactory{
		IdCounter:        0,
		Encoding:         encoding,
		HashFunc:         hashFunc,
		DefaultStateType: defaultStateType,
	}
	return newFactory, nil
}

func (f *EncodeHashStateFactory) GetIdCounter() StateId {
	return f.IdCounter
}

func (f *EncodeHashStateFactory) SetIdCounter(countPos StateId) error {
	f.IdCounter = countPos
	return nil
}

func (f *EncodeHashStateFactory) GetDefaultStateType() StateType {
	return f.DefaultStateType
}

func (f *EncodeHashStateFactory) SetDefaultStateType(newType StateType) error {
	switch newType {
	case LAZYDFAANNOTATED:
		f.DefaultStateType = newType
	default:
		return ErrInvalidStateType
	}
	return nil
}

func (f *EncodeHashStateFactory) NewState() (State, error) {
	var newState State

	switch f.DefaultStateType {
	case LAZYDFAANNOTATED:
		newState = NewLazyDfaAnnotatedState(f.IdCounter, f.Encoding, f.HashFunc)
	default:
		return nil, ErrInvalidStateType
	}
	f.IdCounter += 1

	return newState, nil
}

func (f *EncodeHashStateFactory) CloneState(orig State) (State, error) {
	clone := orig.Clone()

	if err := clone.SetId(f.IdCounter); err != nil {
		return nil, err
	}
	f.IdCounter += 1

	return clone, nil
}
