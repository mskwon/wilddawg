package wilddawg

import (
	"hash/fnv"
	"testing"

	"github.com/ugorji/go/codec"
)

func TestLazyDfaStatefulStateId(t *testing.T) {
	var testState State = NewLazyDfaStatefulState(55, nil, nil)

	if stateId := testState.GetId(); stateId != 55 {
		t.Errorf("State Id: %d, want 55", stateId)
	}

	if err := testState.SetId(77); err != nil {
		t.Errorf("Error while trying to set Id to 77: %q", err)
	}

	if stateId := testState.GetId(); stateId != 77 {
		t.Errorf("State Id: %d after calling SetId(77), want 77", stateId)
	}
}

func TestLazyDfaStatefulStateTerminal(t *testing.T) {
	var testState State = NewLazyDfaStatefulState(55, nil, nil)

	if terminal := testState.IsTerminal(); terminal {
		t.Errorf("Terminal state: %t, want false after initialization",
			terminal)
	}

	if err := testState.SetTerminal(true); err != nil {
		t.Errorf("Error while trying to set terminal state to true: %q", err)
	}

	if terminal := testState.IsTerminal(); !terminal {
		t.Errorf("Terminal state: %t after calling SetTerminal(true), "+
			"want true", terminal)
	}

	if err := testState.SetTerminal(false); err != nil {
		t.Errorf("Error while trying to set terminal state to false: %q", err)
	}

	if terminal := testState.IsTerminal(); terminal {
		t.Errorf("Terminal state: %t after calling SetTerminal(false),"+
			" want false", terminal)
	}
}

func slicesSameValues(a []interface{}, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	count := make(map[interface{}]int)
	for _, el := range a {
		count[el] += 1
	}
	for _, el := range b {
		count[el] -= 1
	}
	for _, c := range count {
		if c != 0 {
			return false
		}
	}
	return true
}

func TestLazyDfaStatefulStateAnnotationsString(t *testing.T) {
	var testState State = NewLazyDfaStatefulState(55, nil, nil)

	if annotations, err := testState.GetAnnotations(); err != nil {
		t.Errorf("Error while getting annotations: %q", err)
	} else if len(annotations) != 0 {
		t.Errorf("GetAnnotations() returned a slice with %d elements, want "+
			"empty slice on initialization", len(annotations))
	}

	expected := make([]interface{}, 0)
	expected = append(expected, "a")
	if err := testState.AddAnnotation("a"); err != nil {
		t.Errorf("Error while adding annotation \"a\": %q", err)
	}

	if annotations, err := testState.GetAnnotations(); err != nil {
		t.Errorf("Error while getting annotations: %q", err)
	} else if !slicesSameValues(annotations, expected) {
		t.Errorf("GetAnnotations() returned %v, want %v", annotations, expected)
	}

	if err := testState.AddAnnotation("a"); err != nil {
		t.Errorf("Error while adding annotation \"a\": %q", err)
	}

	if annotations, err := testState.GetAnnotations(); err != nil {
		t.Errorf("Error while getting annotations: %q", err)
	} else if !slicesSameValues(annotations, expected) {
		t.Errorf("GetAnnotations() returned %v, want %v", annotations, expected)
	}

	expected = append(expected, "b")
	if err := testState.AddAnnotation("b"); err != nil {
		t.Errorf("Error while adding annotation \"b\": %q", err)
	}

	if annotations, err := testState.GetAnnotations(); err != nil {
		t.Errorf("Error while getting annotations: %q", err)
	} else if !slicesSameValues(annotations, expected) {
		t.Errorf("GetAnnotations() returned %v, want %v", annotations, expected)
	}

	expected = expected[1:]
	if err := testState.RemoveAnnotation("a"); err != nil {
		t.Errorf("Error while removing annotation \"a\": %q", err)
	}

	if annotations, err := testState.GetAnnotations(); err != nil {
		t.Errorf("Error while getting annotations: %q", err)
	} else if !slicesSameValues(annotations, expected) {
		t.Errorf("GetAnnotations() returned %v, want %v", annotations, expected)
	}

	if err := testState.RemoveAnnotation("x"); err != ErrAnnotationInvalid {
		t.Errorf("Removing invalid annotation, expected %q, got %q",
			ErrAnnotationInvalid, err)
	}
}

func TestLazyDfaStatefulStateEdge(t *testing.T) {
	var testStateA State = NewLazyDfaStatefulState(1, nil, nil)
	var testStateB State = NewLazyDfaStatefulState(2, nil, nil)

	if err := testStateA.AddEdge("a", testStateB); err != nil {
		t.Errorf("Error while adding edge: %q", err)
	}
	if err := testStateA.AddEdge("b", testStateB); err != nil {
		t.Errorf("Error while adding edge: %q", err)
	}

	if dest := testStateA.FollowEdge("a"); len(dest) != 1 {
		t.Errorf("Destination state count %d, want 1", len(dest))
	}
	if dest := testStateA.FollowEdge("a"); dest[0] != testStateB {
		t.Errorf("Result state %v, wanted %v", dest[0], testStateB)
	}
	if dest := testStateA.FollowEdge("x"); len(dest) != 0 {
		t.Errorf("Destination state count %d, want 0", len(dest))
	}
	if dest := testStateA.FollowAllEdges(); len(dest) != 1 {
		t.Errorf("Destination state count %d (%v), want 1", len(dest), dest)
	}

	var testStateC State = NewLazyDfaStatefulState(3, nil, nil)
	if err := testStateA.AddEdge("a", testStateC); err != ErrEdgeAlreadyUsed {
		t.Errorf("Expected %q, got %q", ErrEdgeAlreadyUsed, err)
	}
	if err := testStateA.AddEdge("c", testStateC); err != nil {
		t.Errorf("Error while adding edge: %q", err)
	}

	if dest := testStateA.FollowEdge("a"); len(dest) != 1 {
		t.Errorf("Destination state count %d, want 1", len(dest))
	}
	if dest := testStateA.FollowEdge("a"); dest[0] != testStateB {
		t.Errorf("Result state %v, wanted %v", dest[0], testStateB)
	}
	if dest := testStateA.FollowEdge("c"); len(dest) != 1 {
		t.Errorf("Destination state count %d, want 1", len(dest))
	}
	if dest := testStateA.FollowEdge("c"); dest[0] != testStateC {
		t.Errorf("Result state %v, wanted %v", dest[0], testStateC)
	}

	if dest := testStateA.FollowAllEdges(); len(dest) != 2 {
		t.Errorf("Destination state count %d (%v), want 2", len(dest), dest)
	}

	if err := testStateA.RemoveEdge("d", nil); err != ErrEdgeNotPresent {
		t.Errorf("Expected %q, got %q", ErrEdgeNotPresent, err)
	}
	if err := testStateA.RemoveEdge("a", testStateC); err != ErrEdgeNotPresent {
		t.Errorf("Expected %q, got %q", ErrEdgeNotPresent, err)
	}
	if err := testStateA.RemoveEdge("a", testStateB); err != nil {
		t.Errorf("Error while removing edge: %q", err)
	}

	if dest := testStateA.FollowEdge("a"); len(dest) != 0 {
		t.Errorf("Destination state count %d, want 0", len(dest))
	}
	if dest := testStateA.FollowAllEdges(); len(dest) != 2 {
		t.Errorf("Destination state count %d (%v), want 2", len(dest), dest)
	}

	if err := testStateA.RemoveEdge("b", testStateB); err != nil {
		t.Errorf("Error while removing edge: %q", err)
	}
	if dest := testStateA.FollowAllEdges(); len(dest) != 1 {
		t.Errorf("Destination state count %d (%v), want 1", len(dest), dest)
	}
}

func sameMachineEdges(a map[interface{}]StateId,
	b map[interface{}]StateId) bool {
	if len(a) != len(b) {
		return false
	}
	for k, a_val := range a {
		if b_val, present := b[k]; !present {
			return false
		} else if a_val != b_val {
			return false
		}
	}
	return true
}

func TestLazyDfaStatefulStateMachineEdges(t *testing.T) {
	var testStateA State = NewLazyDfaStatefulState(1, nil, nil)

	if edges := testStateA.MachineEdges(); len(edges) != 0 {
		t.Errorf("Expected 0 machine edges, got %d", len(edges))
	}

	expected := make(map[interface{}]StateId)
	var testStateB State = NewLazyDfaStatefulState(2, nil, nil)

	expected["a"] = 2
	if err := testStateA.AddEdge("a", testStateB); err != nil {
		t.Errorf("Error while adding edge: %q", err)
	}
	if edges := testStateA.MachineEdges(); !sameMachineEdges(edges, expected) {
		t.Errorf("Expected %v, got %v", expected, edges)
	}

	expected["b"] = 2
	if err := testStateA.AddEdge("b", testStateB); err != nil {
		t.Errorf("Error while adding edge: %q", err)
	}
	if edges := testStateA.MachineEdges(); !sameMachineEdges(edges, expected) {
		t.Errorf("Expected %v, got %v", expected, edges)
	}

	var testStateC State = NewLazyDfaStatefulState(3, nil, nil)
	expected["c"] = 3
	if err := testStateA.AddEdge("c", testStateC); err != nil {
		t.Errorf("Error while adding edge: %q", err)
	}
	if edges := testStateA.MachineEdges(); !sameMachineEdges(edges, expected) {
		t.Errorf("Expected %v, got %v", expected, edges)
	}

	delete(expected, "b")
	if err := testStateA.RemoveEdge("b", testStateB); err != nil {
		t.Errorf("Error while removing edge: %q", err)
	}
	if edges := testStateA.MachineEdges(); !sameMachineEdges(edges, expected) {
		t.Errorf("Expected %v, got %v", expected, edges)
	}
}

func TestLazyDfaStatefulStateIsomorphismHash(t *testing.T) {
	hashFunc := func(data map[interface{}]StateId) uint32 {
		codecHandle := new(codec.BincHandle)
		codecHandle.Canonical = true
		encodedBytes := make([]byte, 0, 64)
		encoder := codec.NewEncoderBytes(&encodedBytes, codecHandle)
		if err := encoder.Encode(data); err != nil {
			t.Errorf("Error while running validation encoding func: %q", err)
		}
		fnv := fnv.New32()
		if _, err := fnv.Write(encodedBytes); err != nil {
			t.Errorf("Error while running validation hash func: %q", err)
		}
		return fnv.Sum32()
	}
	expected := make(map[interface{}]StateId)

	sharedCodecHandle := new(codec.BincHandle)
	sharedCodecHandle.Canonical = true
	sharedHashFunc := fnv.New32()

	var testStateA State = NewLazyDfaStatefulState(1, sharedCodecHandle,
		sharedHashFunc)
	if hash, err := testStateA.IsomorphismHash(); err != nil {
		t.Errorf("Error while obtaining IsomorphismHash: %q", err)
	} else if expectedHash := hashFunc(expected); hash != expectedHash {
		t.Errorf("Expected hash %d, got %d", expectedHash, hash)
	}

	var testStateB State = NewLazyDfaStatefulState(2, sharedCodecHandle,
		sharedHashFunc)
	expected["a"] = 2
	if err := testStateA.AddEdge("a", testStateB); err != nil {
		t.Errorf("Error while adding edge: %q", err)
	}
	if hash, err := testStateA.IsomorphismHash(); err != nil {
		t.Errorf("Error while obtaining IsomorphismHash: %q", err)
	} else if expectedHash := hashFunc(expected); hash != expectedHash {
		t.Errorf("Expected hash %d, got %d", expectedHash, hash)
	}

	expected["b"] = 2
	if err := testStateA.AddEdge("b", testStateB); err != nil {
		t.Errorf("Error while adding edge: %q", err)
	}
	if hash, err := testStateA.IsomorphismHash(); err != nil {
		t.Errorf("Error while obtaining IsomorphismHash: %q", err)
	} else if expectedHash := hashFunc(expected); hash != expectedHash {
		t.Errorf("Expected hash %d, got %d", expectedHash, hash)
	}

	expected["c"] = 1
	if err := testStateA.AddEdge("c", testStateA); err != nil {
		t.Errorf("Error while adding edge: %q", err)
	}
	if hash, err := testStateA.IsomorphismHash(); err != nil {
		t.Errorf("Error while obtaining IsomorphismHash: %q", err)
	} else if expectedHash := hashFunc(expected); hash != expectedHash {
		t.Errorf("Expected hash %d, got %d", expectedHash, hash)
	}

	delete(expected, "a")
	delete(expected, "b")
	delete(expected, "c")
	if hash, err := testStateB.IsomorphismHash(); err != nil {
		t.Errorf("Error while obtaining IsomorphismHash: %q", err)
	} else if expectedHash := hashFunc(expected); hash != expectedHash {
		t.Errorf("Expected hash %d, got %d", expectedHash, hash)
	}

	var testStateC State = NewLazyDfaStatefulState(3, nil, nil)
	if _, err := testStateC.IsomorphismHash(); err != ErrNilEncoder {
		t.Errorf("Expected %q, got %q", ErrNilEncoder, err)
	}

	var testStateD State = NewLazyDfaStatefulState(4, sharedCodecHandle, nil)
	if _, err := testStateD.IsomorphismHash(); err != ErrNilHashFunc {
		t.Errorf("Expected %q, got %q", ErrNilHashFunc, err)
	}
}

func TestLazyDfaStatefulStateType(t *testing.T) {
	var testStateA State = NewLazyDfaStatefulState(1, nil, nil)

	if stateType := testStateA.GetStateType(); stateType != LAZYDFASTATEFUL {
		t.Errorf("Expected StateType %d, got %d", LAZYDFASTATEFUL, stateType)
	}
}
