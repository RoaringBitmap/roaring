package roaring

import "testing"

//hack to return a container with just the 0 value in the container.
func ZeroContainer(test *testing.T) container {
	bytes := make([]byte, 4)
	//if you uncomment this, the test passes
	//test.Logf("produced: %+v", bytes)
	return &arrayContainer{byteSliceAsUint16Slice(bytes[2:])}
}

func containerGenerator(test *testing.T) (container, []container) {
	firstContainer := ZeroContainer(test)
	containerSlice := make([]container, 0)
	secondContainer := ZeroContainer(test)
	containerSlice = append(containerSlice, secondContainer)
	return firstContainer, containerSlice
}

func workingContainerGenerator(test *testing.T) (container, []container) {
	firstContainer := ZeroContainer(test)
	secondContainer := ZeroContainer(test)
	return firstContainer, []container{secondContainer}
}

func TestComparison(test *testing.T) {
	// the commented out line passes
	a, b := containerGenerator(test)
	//a, b := workingContainerGenerator(test)

	// if you comment out this line the test passes
	test.Logf("produced: %+v,%+v", a, b[0])
	if !a.equals(b[0]) {
		test.Errorf("unexpected:%+v,%+v", a, b[0])
	}
}