package bwgame

const (
	fnv1a64OffsetBasis uint64 = 0xcbf29ce484222325
	fnv1a64Prime       uint64 = 0x100000001b3
)

type StateHasher struct {
	Value uint64
}

func NewStateHasher() StateHasher {
	return StateHasher{Value: fnv1a64OffsetBasis}
}

func (h *StateHasher) AddInt32(v int32) {
	bits := uint32(v)
	for i := 0; i != 4; i++ {
		h.Value ^= uint64((bits >> (8 * i)) & 0xff)
		h.Value *= fnv1a64Prime
	}
}

func UnitOrderID(u *Unit) int32 {
	if u.OrderType == nil {
		return -1
	}
	return int32(u.OrderType.ID)
}

func (h *StateHasher) addUnit(u *Unit) {
	h.AddInt32(int32(u.Index))
	h.AddInt32(int32(u.UnitType.ID))
	h.AddInt32(u.Owner)
	h.AddInt32(u.ExactPosition.X.RawValue)
	h.AddInt32(u.ExactPosition.Y.RawValue)
	h.AddInt32(u.Hp.RawValue)
	h.AddInt32(u.ShieldPoints.RawValue)
	h.AddInt32(u.Energy.RawValue)
	h.AddInt32(UnitOrderID(u))
	h.AddInt32(u.StatusFlags)
}

func (h *StateHasher) addUnitList(units *UnitList) {
	count := int32(0)
	for range units.All() {
		count++
	}
	h.AddInt32(count)
	for u := range units.All() {
		h.addUnit(u)
	}
}

func (h *StateHasher) addPlayers(st *State) {
	for i := 0; i != 12; i++ {
		h.AddInt32(st.CurrentMinerals[i])
		h.AddInt32(st.CurrentGas[i])
		for race := 0; race != 3; race++ {
			h.AddInt32(st.SupplyUsed[i][race].RawValue)
		}
		for race := 0; race != 3; race++ {
			h.AddInt32(st.SupplyAvailable[i][race].RawValue)
		}
	}
}

func HashedUnitLists(st *State) []*UnitList {
	return []*UnitList{&st.VisibleUnits, &st.HiddenUnits, &st.MapRevealerUnits}
}

func StateHash(st *State) uint64 {
	h := NewStateHasher()
	h.AddInt32(st.CurrentFrame)
	h.AddInt32(int32(st.LcgRandState))
	h.addPlayers(st)
	for _, units := range HashedUnitLists(st) {
		h.addUnitList(units)
	}
	return h.Value
}
