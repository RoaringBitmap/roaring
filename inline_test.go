package roaring

import "testing"

//hack to return a container with just the 0 value in the container.
func ZeroContainer(test *testing.T) container {
	bytes := make([]byte, 4)
	return &arrayContainer{byteSliceAsUint16Slice(bytes[2:])}
}

func inlinedFunction(test *testing.T) (container, []container) {
	populationContainer := ZeroContainer(test)
	observationContainers := make([]container, 0)
	container:= ZeroContainer(test)
	observationContainers = append(observationContainers, container)
	return populationContainer, observationContainers
}

func TestHundredTimes(test *testing.T) {
	for i := 0; i < 100; i++ {
		testComparison(test)
	}
}

func testComparison(test *testing.T) {
	a, b := inlinedFunction(test)
	//this line is apparently required?
	test.Logf("produced: %+v,%+v", a, b[0])
	if !a.equals(b[0]) {
		test.Errorf("unexpected:%+v,%+v", a, b[0])
	}
}