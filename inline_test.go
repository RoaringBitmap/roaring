package roaring

import "testing"

//hack to return a container with just the 0 value in the container.
func UnitZeroContainer(test *testing.T) container {
	bytes := make([]byte, 4)
	return &arrayContainer{byteSliceAsUint16Slice(bytes[2:])}
}

func inlinedFunction(test *testing.T) (container, []container) {
	populationContainer := UnitZeroContainer(test)
	observationContainers := make([]container, 0)
	container:= UnitZeroContainer(test)
	observationContainers = append(observationContainers, container)
	return populationContainer, observationContainers
}

func TestHundredTimes(test *testing.T) {
	for i := 0; i < 100; i++ {
		testUnits(test)
	}
}

func testUnits(test *testing.T) {
	pop, obs := inlinedFunction(test)
	//this line is apparently required?
	test.Logf("produced: %+v,%+v", pop, obs[0])
	if !pop.equals(obs[0]) {
		test.Errorf("unexpected:%+v,%+v", pop, obs[0])
	}
}