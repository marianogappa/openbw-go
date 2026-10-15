package bwgame

type contourIndex struct {
	firstValue  int32
	lowerBounds []int32
	total       int
	built       bool
}

func newContourIndex(contours []Contour) contourIndex {
	index := contourIndex{total: len(contours), built: true}
	if len(contours) == 0 {
		return index
	}
	index.firstValue = contours[0].V[0]
	lastValue := contours[len(contours)-1].V[0]
	index.lowerBounds = make([]int32, lastValue-index.firstValue+1)
	next := 0
	for i := range index.lowerBounds {
		value := index.firstValue + int32(i)
		for next != len(contours) && contours[next].V[0] < value {
			next++
		}
		index.lowerBounds[i] = int32(next)
	}
	return index
}

func (index *contourIndex) lowerBound(value int64) int {
	offset := value - int64(index.firstValue)
	if offset <= 0 {
		return 0
	}
	if offset >= int64(len(index.lowerBounds)) {
		return index.total
	}
	return int(index.lowerBounds[offset])
}

func (r *Regions) contourIndexFor(dir int) *contourIndex {
	index := &r.contourIndexes[dir]
	if !index.built {
		*index = newContourIndex(r.Contours[dir])
	}
	return index
}

func (r *Regions) contourLowerBound(dir int, value int32) int {
	return r.contourIndexFor(dir).lowerBound(int64(value))
}

func (r *Regions) contourUpperBound(dir int, value int32) int {
	return r.contourIndexFor(dir).lowerBound(int64(value) + 1)
}
