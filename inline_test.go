package roaring

import "testing"

//hack to return a container with just the 0 value in the container.
func UnitZeroContainer(test *testing.T) container {
	bytes := make([]byte, 4)
	return ContainerFromBytes(bytes)
}

func testSingleUnitContainerSegmentWithEvents(test *testing.T) {
	pop, obs := inlinedFunction(test)
	test.Logf("produced: %+v\n%+v",
		pop,
		obs)
	union := pop.or(obs[0])
	if pop.getCardinality()  != union.getCardinality() {
		test.Errorf("unexpected:%+v,%+v,%+v",
			pop,
			obs,
			union)
	}
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
	test.Logf("produced: %+v\n%+v",
		pop,
		obs)
	union := pop.or(obs[0])
	if pop.getCardinality()  != union.getCardinality() {
		test.Errorf("unexpected:%+v,%+v,%+v",
			pop,
			obs,
			union)
	}
}