package bwgame

func (f *StateFunctions) pathfinderFindLongPath(pf *Pathfinder) bool {
	panic(notPorted("pathfinder_find_long_path"))
}

func binaryHeapUp[T any](open []T, element int, compare func(a, b T) bool) {
	panic(notPorted("binary_heap_up"))
}

func binaryHeapDown[T any](open []T, element int, compare func(a, b T) bool) {
	panic(notPorted("binary_heap_down"))
}

func (f *StateFunctions) pathfinderUnitCanCollideWith(u *Unit, target *Unit, considerCollisionWithUnit *Unit, considerCollisionWithMovingUnits bool) bool {
	panic(notPorted("pathfinder_unit_can_collide_with"))
}

func (f *StateFunctions) pathfinderUnitCanCollideWithPF(pf *Pathfinder, target *Unit) bool {
	panic(notPorted("pathfinder_unit_can_collide_with"))
}

func (f *StateFunctions) pathfinderFindShortPath(pf *Pathfinder, target Xy, targetRegion *Region) {
	panic(notPorted("pathfinder_find_short_path"))
}

func (f *StateFunctions) pathfinderAdjustTargetPos(unitInnerBb Rect, target Xy) (bool, Xy) {
	panic(notPorted("pathfinder_adjust_target_pos"))
}

func (f *StateFunctions) pathfinderAdjustDestination(sourceRegion *Region, destination Xy) Xy {
	panic(notPorted("pathfinder_adjust_destination"))
}

func (f *StateFunctions) pathfinderFindNextShortPath(pf *Pathfinder) bool {
	panic(notPorted("pathfinder_find_next_short_path"))
}

func (f *StateFunctions) pathfinderFindLongPathFromTo(pf *Pathfinder, from Xy, to Xy) bool {
	panic(notPorted("pathfinder_find_long_path"))
}

func (f *StateFunctions) pathfinderFind(pf *Pathfinder, shortPathOnly bool) bool {
	panic(notPorted("pathfinder_find"))
}

func (f *StateFunctions) newPath() *Path {
	panic(notPorted("new_path"))
}

func (f *StateFunctions) freeUnitPath(u *Unit) {
	panic(notPorted("free_path"))
}

func (f *StateFunctions) freePath(path *Path) {
	panic(notPorted("free_path"))
}

func (f *StateFunctions) pathProgress(u *Unit, to Xy, considerCollisionWithUnit *Unit, considerCollisionWithMovingUnits bool) bool {
	panic(notPorted("path_progress"))
}

func (f *StateFunctions) createSingleStepPath(source Xy, destination Xy) *Path {
	panic(notPorted("create_single_step_path"))
}

func (f *StateFunctions) unitPathTo(u *Unit, to Xy, considerCollisionWithUnit *Unit, considerCollisionWithMovingUnits bool) bool {
	panic(notPorted("unit_path_to"))
}

func (f *StateFunctions) unitUpdatePathMovementState(u *Unit, allowNewPath bool) bool {
	panic(notPorted("unit_update_path_movement_state"))
}
