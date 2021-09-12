package corners

import (
	"testing"
)

func isStartingPointGenerator(override MapOverride, startingPoint [2]int) bool {
	return override.Generator &&
		override.X == startingPoint[0] &&
		override.Y == startingPoint[1]
}

func TestStartingPointsHaveGenerators(t *testing.T) {
	startingPoints := [][2]int{
		{02, 01},
		{13, 13},
		{15, 03},
	}

	generatedMap := GenerateRandomMap(Options{
		startingPoints:     startingPoints,
		numberOfGenerators: 0,
		numberOfWalls:      0,
	})

	var missing [][2]int

	for _, startingPoint := range startingPoints {
		found := false

		for _, override := range generatedMap.Overrides {
			// discount generators created for starting points
			if isStartingPointGenerator(override, startingPoint) {
				found = true

				break
			}
		}

		if !found {
			missing = append(missing, startingPoint)
		}
	}

	if len(missing) > 0 {
		t.Errorf("%d starting points don't have a generator", len(missing))
	}
}

func expectExtrasCount(t *testing.T, expectedGenerators, expectedWalls int) {
	startingPoints := [][2]int{
		{02, 01},
		{13, 13},
		{15, 03},
	}

	generatedMap := GenerateRandomMap(Options{
		startingPoints:     startingPoints,
		numberOfGenerators: expectedGenerators,
		numberOfWalls:      expectedWalls,
	})

	extraGeneratorsCount := 0
	extraWallsCount := 0

	for _, override := range generatedMap.Overrides {
		extra := true

		for _, startingPoint := range startingPoints {
			// discount generators created for starting points
			if isStartingPointGenerator(override, startingPoint) {
				extra = false

				break
			}
		}

		if extra {
			if override.Generator {
				extraGeneratorsCount += 1
			} else {
				extraWallsCount += 1
			}
		}
	}

	if extraGeneratorsCount != expectedGenerators {
		t.Errorf(
			"Found %d extra generators but expected %d",
			extraGeneratorsCount,
			expectedGenerators,
		)
	}

	if extraWallsCount != expectedWalls {
		t.Errorf(
			"Found %d extra walls but expected %d",
			extraWallsCount,
			expectedWalls,
		)
	}
}

func TestNoExtraGeneratorsOrWalls(t *testing.T) {
	expectExtrasCount(t, 0, 0)
}

func TestGeneratesExtraGeneratorsButNoWalls(t *testing.T) {
	expectExtrasCount(t, 10, 10)
}

func TestGeneratesNoExtraGeneratorsButExtraWalls(t *testing.T) {
	expectExtrasCount(t, 0, 10)
}

func TestGeneratesExtraGeneratorsAndWalls(t *testing.T) {
	expectExtrasCount(t, 10, 10)
}
