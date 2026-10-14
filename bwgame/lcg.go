package bwgame

func (f *StateFunctions) lcgRand(source int32) int32 {
	st := f.st
	st.RandomCounts[source]++
	st.TotalRandomCounts++
	st.LcgRandState = st.LcgRandState*22695477 + 1
	return int32((st.LcgRandState >> 16) & 0x7fff)
}

func (f *StateFunctions) lcgRandRange(source int32, from int32, to int32) int32 {
	return from + ((f.lcgRand(source) * (to - from + 1)) >> 15)
}
