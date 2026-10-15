package bwgame

import (
	"math"
)

func (f *StateFunctions) unitIsAtMoveTarget(u *Flingy) bool {
	return u.Sprite.Position == u.MoveTarget.Pos
}

func (f *StateFunctions) unitDead(u *Unit) bool {
	if u.OrderType == nil {
		return true
	}
	return u.OrderType.ID == OrdersDie && u.OrderState == 1
}

func (f *StateFunctions) unitDying(u *Unit) bool {
	if u.OrderType == nil {
		return true
	}
	return u.OrderType.ID == OrdersDie
}

func (f *StateFunctions) rectangleCanFitAt(pos Xy, width int32, height int32) bool {
	var area Rect
	area.From = XySub(pos, XyDivScalar(Xy{X: width, Y: width}, 2))
	area.To = XyAdd(area.From, Xy{X: width, Y: height})
	if !f.isInInnerMapBounds(area) {
		return false
	}
	if !f.isWalkable(pos) {
		return false
	}
	var inner [4]int32
	inner[0] = (height - 1) / 2
	inner[1] = -(width - (width / 2) - 1)
	inner[2] = -(height - (height / 2) - 1)
	inner[3] = (width - 1) / 2
	return f.canFitAt(pos, inner)
}

func (f *StateFunctions) unitTypeCanFitAt(unitType *UnitTypeT, pos Xy) bool {
	if !f.isUnitTypeInMapBounds(unitType, pos) {
		return false
	}
	if !f.isWalkable(pos) {
		return false
	}
	var inner [4]int32
	inner[0] = unitType.Dimensions.From.Y
	inner[1] = -unitType.Dimensions.To.X
	inner[2] = -unitType.Dimensions.To.Y
	inner[3] = unitType.Dimensions.From.X
	return f.canFitAt(pos, inner)
}

func (f *StateFunctions) canFitAt(pos Xy, inner [4]int32) bool {
	c0 := f.gameSt.Regions.Contours[0]
	for i := f.gameSt.Regions.contourLowerBound(0, pos.Y); i != 0; {
		i--
		if inner[0]+c0[i].V[0] < pos.Y {
			break
		}
		if inner[1]+c0[i].V[1] <= pos.X && inner[3]+c0[i].V[2] >= pos.X {
			return false
		}
	}
	c1 := f.gameSt.Regions.Contours[1]
	for i := f.gameSt.Regions.contourUpperBound(1, pos.X); i != len(c1); i++ {
		if inner[1]+c1[i].V[0] > pos.X {
			break
		}
		if inner[2]+c1[i].V[1] <= pos.Y && inner[0]+c1[i].V[2] >= pos.Y {
			return false
		}
	}
	c2 := f.gameSt.Regions.Contours[2]
	for i := f.gameSt.Regions.contourUpperBound(2, pos.Y); i != len(c2); i++ {
		if inner[2]+c2[i].V[0] > pos.Y {
			break
		}
		if inner[1]+c2[i].V[1] <= pos.X && inner[3]+c2[i].V[2] >= pos.X {
			return false
		}
	}
	c3 := f.gameSt.Regions.Contours[3]
	for i := f.gameSt.Regions.contourLowerBound(3, pos.X); i != 0; {
		i--
		if inner[3]+c3[i].V[0] < pos.X {
			break
		}
		if inner[2]+c3[i].V[1] <= pos.Y && inner[0]+c3[i].V[2] >= pos.Y {
			return false
		}
	}
	return true
}

func (f *StateFunctions) unitTargetIsEnemy(u *Unit, target *Unit) bool {
	nOwner := target.Owner
	if nOwner == 11 {
		nOwner = target.Sprite.Owner
	}
	return f.st.Alliances[u.Owner][nOwner] == 0
}

func (f *StateFunctions) getLargestBlockingUnit(u *Unit, bounds Rect) *Unit {
	var largestUnitArea int32
	var largestUnit *Unit
	for nu := range f.findUnits(bounds) {
		if f.unitCanCollideWith(u, nu) && f.unitFinderUnitInBounds(nu, bounds) {
			nBB := f.unitTypeBoundingBox(nu.UnitType)
			p := (nBB.To.X - nBB.From.X) * (nBB.To.Y - nBB.From.Y)
			if p > largestUnitArea {
				largestUnitArea = p
				largestUnit = nu
			}
		}
	}
	return largestUnit
}

func (f *StateFunctions) isBlocked(u *Unit, pos Xy) (bool, *Unit) {
	bounds := f.unitInnerBoundingBox(u, pos)
	if !f.isInInnerMapBounds(bounds) {
		return true, nil
	}
	largestUnit := f.getLargestBlockingUnit(u, bounds)
	if largestUnit == nil {
		return !f.unitTypeCanFitAt(u.UnitType, pos), nil
	}
	return true, largestUnit
}

func (f *StateFunctions) setFlingyMoveTarget(u *Flingy, targetPos Xy, targetUnit *Unit) {
	if u.MoveTarget.Pos == targetPos && (targetUnit == nil || u.MoveTarget.Unit == targetUnit) {
		return
	}
	u.MoveTarget.Pos = targetPos
	u.MoveTarget.Unit = targetUnit
	u.NextMovementWaypoint = targetPos
	f.uUnsetMovementFlag(u, 4)
	f.uSetMovementFlag(u, 1)
}

func (f *StateFunctions) setUnitMoveTarget(u *Unit, moveTarget Xy) {
	if u.MoveTarget.Pos == moveTarget {
		return
	}
	if u.Path != nil {
		u.Path.StateFlags |= 1
	}
	moveTarget = f.restrictMoveTargetToValidBounds(u.UnitType, moveTarget)
	f.setFlingyMoveTarget(&u.Flingy, moveTarget, nil)
	f.finishSetUnitMoveTarget(u)
}

func (f *StateFunctions) setUnitMoveTargetUnit(u *Unit, target *Unit) {
	targetPos := target.Sprite.Position
	if u.MoveTarget.Pos == targetPos && u.MoveTarget.Unit == target {
		return
	}
	if u.Path != nil {
		u.Path.StateFlags |= 1
	}
	f.setFlingyMoveTarget(&u.Flingy, targetPos, target)
	f.finishSetUnitMoveTarget(u)
}

func (f *StateFunctions) finishSetUnitMoveTarget(u *Unit) {
	if f.uImmovable(u) {
		f.uUnsetStatusFlag(u, UnitStatusFlagImmovable)
	}
	u.MoveTargetTimer = 15
	if !u.OrderQueue.Empty() && !u.OrderQueue.Front().OrderType.Unk7 {
		f.uSetMovementFlag(&u.Flingy, 0x20)
	} else {
		f.uUnsetMovementFlag(&u.Flingy, 0x20)
	}
}

func (f *StateFunctions) setCurrentVelocityDirection(u *Flingy, currentVelocityDirection DirectionT) {
	if u.CurrentVelocityDirection == currentVelocityDirection {
		return
	}
	u.CurrentVelocityDirection = currentVelocityDirection
	u.Velocity = f.directionXY(currentVelocityDirection, u.CurrentSpeed)
}

func (f *StateFunctions) setCurrentSpeed(u *Flingy, currentSpeed Fp8) {
	if u.CurrentSpeed == currentSpeed {
		return
	}
	u.CurrentSpeed = currentSpeed
	u.Velocity = f.directionXY(u.CurrentVelocityDirection, u.CurrentSpeed)
}

func (f *StateFunctions) setNextSpeed(u *Flingy, nextSpeed Fp8) {
	u.NextSpeed = nextSpeed
	f.setCurrentSpeed(u, nextSpeed)
}

func (f *StateFunctions) unitTurnRate(u *Flingy, desiredTurn DirectionT) DirectionT {
	uturnRate := u.FlingyTurnRate.AsUnsigned()
	if u.FlingyMovementType != 2 {
		uturnRate = uturnRate.DivInt(2)
	}
	turnRate := uturnRate.AsSigned()
	turn := Fp8Extend(desiredTurn)
	if turn.Greater(turnRate) {
		turn = turnRate
	} else if turn.Less(turnRate.Neg()) {
		turn = turnRate.Neg()
	}
	return DirectionTTruncate(turn)
}

func (f *StateFunctions) setDesiredVelocityDirection(u *Flingy, desiredVelocityDirection DirectionT) {
	u.DesiredVelocityDirection = desiredVelocityDirection
	if u.NextVelocityDirection != desiredVelocityDirection {
		turn := f.unitTurnRate(u, desiredVelocityDirection.Sub(u.NextVelocityDirection))
		f.setCurrentVelocityDirection(u, u.NextVelocityDirection.Add(turn))
	} else {
		f.setCurrentVelocityDirection(u, desiredVelocityDirection)
	}
}

func (f *StateFunctions) updateCurrentVelocityDirectionTowardsWaypoint(u *Flingy) {
	if u.Position != u.NextMovementWaypoint {
		f.setDesiredVelocityDirection(u, f.xyDirection(XySub(u.NextMovementWaypoint, u.Position)))
	} else if u.Position != u.NextTargetWaypoint {
		f.setDesiredVelocityDirection(u, f.xyDirection(XySub(u.NextTargetWaypoint, u.Position)))
	} else {
		f.setDesiredVelocityDirection(u, u.Heading)
	}
}

func (f *StateFunctions) updateUnitHeading(u *Flingy, velocityDirection DirectionT) {
	u.NextVelocityDirection = velocityDirection
	if !f.uMovementFlag(u, 2) || f.uMovementFlag(u, 1) {
		turn := u.DesiredVelocityDirection.Sub(u.Heading)
		if turn.Greater(DirectionTTruncate(u.FlingyTurnRate)) {
			turn = DirectionTTruncate(u.FlingyTurnRate)
		} else if turn.Less(DirectionTTruncate(u.FlingyTurnRate.Neg())) {
			turn = DirectionTTruncate(u.FlingyTurnRate).Neg()
		}
		u.Heading = u.Heading.Add(turn)
		if u.FlingyType.ID >= 0x8d && u.FlingyType.ID <= 0xab {
			u.FlingyTurnRate = u.FlingyTurnRate.Add(Fp8FromRaw(1))
		} else if u.FlingyType.ID >= 0xc9 && u.FlingyType.ID <= 0xce {
			u.FlingyTurnRate = u.FlingyTurnRate.Add(Fp8FromRaw(1))
		}
		if velocityDirection == u.DesiredVelocityDirection {
			if u.Heading == u.DesiredVelocityDirection {
				f.uUnsetMovementFlag(u, 1)
			}
		}
	}
	heading := u.Heading
	for image := range u.Sprite.Images.All() {
		f.setImageHeading(image, heading)
	}
}

func (f *StateFunctions) updateCurrentSpeedTowardsWaypoint(u *Flingy) {
	switch u.FlingyMovementType {
	case 0:
		if f.unitIsAtMoveTarget(u) {
			if u.NextSpeed.Less(Fp8FromRaw(192)) {
				if !f.uMovementFlag(u, 0x20) && !f.uMovementFlag(u, 0x10) {
					f.uUnsetMovementFlag(u, 4)
					f.setCurrentSpeed(u, Fp8FromRaw(0))
					return
				}
			}
		}
	case 1:
		if f.unitIsAtMoveTarget(u) {
			f.uUnsetMovementFlag(u, 4)
			f.setCurrentSpeed(u, Fp8FromRaw(0))
			return
		}
	case 2:
		f.updateCurrentSpeedTowardsWaypointType2(u)
		return
	}
	f.setCurrentSpeed(u, u.NextSpeed)
	d := f.xyLength(XySub(u.NextMovementWaypoint, u.Position))
	accelerate := false
	if u.CurrentVelocityDirection == u.DesiredVelocityDirection {
		accelerate = true
	} else if d >= 32 {
		accelerate = true
	} else {
		turnRate := u.FlingyTurnRate
		diff := Fp8Extend(u.DesiredVelocityDirection.Sub(u.CurrentVelocityDirection)).Abs()
		val := uint32(Fp8DivideMultiply(diff.MulInt(2).Add(turnRate).Sub(Fp8FromRaw(1)), turnRate, u.NextSpeed).IntegerPart())
		if val*3/2 <= uint32(d) {
			accelerate = true
		}
	}
	if accelerate {
		if u.FlingyMovementType == 0 {
			if !f.uMovementFlag(u, 0x20) || u.NextMovementWaypoint != u.MoveTarget.Pos {
				if !f.uMovementFlag(u, 0x10) || f.uMovementFlag(u, 0x40) {
					remainingD := f.xyLengthFP8(XyFp8Sub(f.toXyFp8(u.NextMovementWaypoint), u.ExactPosition))
					if f.unitHaltDistance(u).GreaterEqual(remainingD) {
						accelerate = false
					}
				}
			}
		}
	}
	speed := u.CurrentSpeed
	if accelerate {
		speed = speed.Add(u.FlingyAcceleration)
	} else {
		speed = speed.Sub(u.FlingyAcceleration)
	}
	if speed.Less(Fp8FromRaw(0)) {
		speed = Fp8FromRaw(0)
	} else if speed.Greater(u.FlingyTopSpeed) {
		speed = u.FlingyTopSpeed
	}
	f.setCurrentSpeed(u, speed)
}

func (f *StateFunctions) updateCurrentSpeedTowardsWaypointType2(u *Flingy) {
	if f.unitIsAtMoveTarget(u) {
		f.uUnsetMovementFlag(u, 4)
		f.setCurrentSpeed(u, Fp8FromRaw(0))
		return
	}
	if !f.uMovementFlag(u, 1) {
		f.setCurrentSpeed(u, u.NextSpeed)
		return
	}
	headingError := Fp8Extend(u.Heading.Sub(u.DesiredVelocityDirection)).Abs()
	if headingError.GreaterEqual(Fp8FromRaw(32)) {
		if f.uMovementFlag(u, 2) {
			f.uUnsetMovementFlag(u, 2)
			f.uUnsetMovementFlag(u, 4)
			f.setCurrentSpeed(u, Fp8FromRaw(0))
		}
	} else {
		f.setCurrentSpeed(u, u.NextSpeed)
	}
}

func (f *StateFunctions) setMovementFlags(u *Flingy, ems *ExecuteMovementStruct) {
	ems.StartingMovement = false
	ems.StoppingMovement = false
	if !f.unitIsAtMoveTarget(u) {
		if !f.uMovementFlag(u, 2) {
			if u.FlingyMovementType != 2 || !f.uMovementFlag(u, 8) {
				f.uSetMovementFlag(u, 2)
			}
			if !f.uMovementFlag(u, 8) {
				ems.StartingMovement = true
			}
		}
	} else if f.uMovementFlag(u, 2) {
		f.uUnsetMovementFlag(u, 2)
		if !f.uMovementFlag(u, 8) {
			ems.StoppingMovement = true
		}
	}
}

func (f *StateFunctions) setMovementValues(u *Flingy, ems *ExecuteMovementStruct) {
	ems.Speed = u.CurrentSpeed
	if !f.uMovementFlag(u, 2) || u.CurrentSpeed == Fp8FromRaw(0) {
		ems.Position = u.Position
		ems.ExactPosition = u.ExactPosition
		return
	}
	remainingD := f.xyLengthFP8(XyFp8Sub(f.toXyFp8(u.NextMovementWaypoint), u.ExactPosition))
	if u.CurrentSpeed.GreaterEqual(remainingD) {
		ems.Position = u.NextMovementWaypoint
		ems.ExactPosition = f.toXyFp8(ems.Position)
		ems.Speed = remainingD
	} else {
		ems.ExactPosition = XyFp8Add(u.ExactPosition, u.Velocity)
		ems.Position = f.toXy(ems.ExactPosition)
	}
	if u.FlingyMovementType == 2 {
		f.setCurrentSpeed(u, Fp8FromRaw(0))
	}
}

func (f *StateFunctions) updateUnitMovementValues(u *Flingy, ems *ExecuteMovementStruct) {
	ems.PreMovementFlags = u.MovementFlags
	f.updateCurrentVelocityDirectionTowardsWaypoint(u)
	f.updateCurrentSpeedTowardsWaypoint(u)
	f.setMovementFlags(u, ems)
	f.setMovementValues(u, ems)
	ems.PostMovementFlags = u.MovementFlags
	u.MovementFlags = ems.PreMovementFlags
}

func isPosInsideRect(pos Xy, area Rect) bool {
	if pos.X >= area.From.X && pos.Y >= area.From.Y {
		if pos.X <= area.To.X && pos.Y <= area.To.Y {
			return true
		}
	}
	return false
}

func (f *StateFunctions) checkUnitMovementTerrainCollision(u *Unit, movement Xy) bool {
	newPos := XyAdd(u.Sprite.Position, movement)
	if isPosInsideRect(newPos, u.TerrainNoCollisionBounds) {
		return false
	}
	pos := u.Sprite.Position
	bounds := Rect{From: XySub(pos, Xy{X: 128, Y: 128}), To: XyAdd(pos, Xy{X: 128, Y: 128})}

	var inner [4]int32
	inner[0] = u.UnitType.Dimensions.From.Y
	inner[1] = -u.UnitType.Dimensions.To.X
	inner[2] = -u.UnitType.Dimensions.To.Y
	inner[3] = u.UnitType.Dimensions.From.X

	c0 := f.gameSt.Regions.Contours[0]
	c1 := f.gameSt.Regions.Contours[1]
	c2 := f.gameSt.Regions.Contours[2]
	c3 := f.gameSt.Regions.Contours[3]

	i0 := f.gameSt.Regions.contourUpperBound(0, pos.Y-inner[0])
	i1 := f.gameSt.Regions.contourLowerBound(1, pos.X-inner[1])
	i2 := f.gameSt.Regions.contourLowerBound(2, pos.Y-inner[2])
	i3 := f.gameSt.Regions.contourUpperBound(3, pos.X-inner[3])

	expand := func(c Contour) Contour {
		c.V[0] += inner[c.Dir]
		c.V[1] += inner[c.Flags&3]
		c.V[2] += inner[(c.Flags>>2)&3]
		return c
	}

	for {
		done := true
		if i0 != 0 {
			done = false
			i0--
			c := expand(c0[i0])
			if c.V[0] >= bounds.From.Y {
				if c.V[2] >= bounds.From.X && c.V[1] <= bounds.To.X {
					if pos.X < c.V[1] && pos.Y-c.V[0] < c.V[1]-pos.X {
						bounds.To.X = c.V[1] - 1
					} else if pos.X > c.V[2] && pos.Y-c.V[0] < pos.X-c.V[2] {
						bounds.From.X = c.V[2] + 1
					} else {
						bounds.From.Y = c.V[0] + 1
					}
				}
			} else {
				i0 = 0
			}
		}
		if i1 != len(c1) {
			done = false
			c := expand(c1[i1])
			i1++
			if c.V[0] <= bounds.To.X {
				if c.V[2] >= bounds.From.Y && c.V[1] <= bounds.To.Y {
					if pos.Y < c.V[1] && c.V[0]-pos.X < c.V[1]-pos.Y {
						bounds.To.Y = c.V[1] - 1
					} else if pos.Y > c.V[2] && c.V[0]-pos.X < pos.Y-c.V[2] {
						bounds.From.Y = c.V[2] + 1
					} else {
						bounds.To.X = c.V[0] - 1
					}
				}
			} else {
				i1 = len(c1)
			}
		}
		if i2 != len(c2) {
			done = false
			c := expand(c2[i2])
			i2++
			if c.V[0] <= bounds.To.Y {
				if c.V[2] >= bounds.From.X && c.V[1] <= bounds.To.X {
					if pos.X < c.V[1] && c.V[0]-pos.Y < c.V[1]-pos.X {
						bounds.To.X = c.V[1] - 1
					} else if pos.X > c.V[2] && c.V[0]-pos.Y < pos.X-c.V[2] {
						bounds.From.X = c.V[2] + 1
					} else {
						bounds.To.Y = c.V[0] - 1
					}
				}
			} else {
				i2 = len(c2)
			}
		}
		if i3 != 0 {
			done = false
			i3--
			c := expand(c3[i3])
			if c.V[0] >= bounds.From.X {
				if c.V[2] >= bounds.From.Y && c.V[1] <= bounds.To.Y {
					if pos.Y < c.V[1] && pos.X-c.V[0] < c.V[1]-pos.Y {
						bounds.To.Y = c.V[1] - 1
					} else if pos.Y > c.V[2] && pos.X-c.V[0] < pos.Y-c.V[2] {
						bounds.From.Y = c.V[2] + 1
					} else {
						bounds.From.X = c.V[0] + 1
					}
				}
			} else {
				i3 = 0
			}
		}
		if done {
			break
		}
	}

	u.TerrainNoCollisionBounds = bounds
	if isPosInsideRect(newPos, u.TerrainNoCollisionBounds) {
		return false
	}

	minPos := pos
	maxPos := newPos
	if maxPos.X < minPos.X {
		minPos.X, maxPos.X = maxPos.X, minPos.X
	}
	if maxPos.Y < minPos.Y {
		minPos.Y, maxPos.Y = maxPos.Y, minPos.Y
	}

	crosses := func(c Contour, minV int32, maxV int32) bool {
		if c.V[1] <= maxV && c.V[2] >= minV {
			if ^c.Flags&0x30 != 0 {
				return true
			}
			if c.V[1] <= minV {
				return true
			}
			if c.V[2] >= maxV {
				return true
			}
		}
		return false
	}

	if newPos.Y < pos.Y {
		for i := f.gameSt.Regions.contourUpperBound(0, pos.Y-inner[0]); i != 0; {
			i--
			c := expand(c0[i])
			if c.V[0] < newPos.Y {
				break
			}
			if crosses(c, minPos.X, maxPos.X) {
				return true
			}
		}
	}
	if newPos.X > pos.X {
		for i := f.gameSt.Regions.contourLowerBound(1, pos.X-inner[1]); i != len(c1); i++ {
			c := expand(c1[i])
			if c.V[0] > newPos.X {
				break
			}
			if crosses(c, minPos.Y, maxPos.Y) {
				return true
			}
		}
	}
	if newPos.Y > pos.Y {
		for i := f.gameSt.Regions.contourLowerBound(2, pos.Y-inner[2]); i != len(c2); i++ {
			c := expand(c2[i])
			if c.V[0] > newPos.Y {
				break
			}
			if crosses(c, minPos.X, maxPos.X) {
				return true
			}
		}
	}
	if newPos.X < pos.X {
		for i := f.gameSt.Regions.contourUpperBound(3, pos.X-inner[3]); i != 0; {
			i--
			c := expand(c3[i])
			if c.V[0] < newPos.X {
				break
			}
			if crosses(c, minPos.Y, maxPos.Y) {
				return true
			}
		}
	}
	return false
}

func (f *StateFunctions) checkUnitMovementTerrainCollisionEMS(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.uGroundUnit(u) {
		return f.checkUnitMovementTerrainCollision(u, XySub(ems.Position, u.Sprite.Position))
	}
	return !f.isInMapBounds(f.unitBoundingBox(u, ems.Position))
}

func (f *StateFunctions) checkUnitMovementUnitCollision(u *Unit, ems *ExecuteMovementStruct) *Unit {
	if f.usHidden(&u.Thingy) {
		return nil
	}
	movement := XySub(ems.Position, u.Sprite.Position)

	bb := u.UnitFinderBoundingBox
	newBB := bb
	newBB.From = XyAdd(newBB.From, movement)
	newBB.To = XyAdd(newBB.To, movement)

	collides := func(n *Unit) bool {
		return f.unitCanCollideWith(u, n) && f.uGroundUnit(n)
	}

	if movement.X < 0 {
		arr := f.st.UnitFinderX
		for i := unitFinderUpperBound(arr, bb.From.X); i != 0; {
			i--
			if arr[i].Value < newBB.From.X {
				break
			}
			n := arr[i].U
			if n.UnitFinderBoundingBox.From.Y <= newBB.To.Y && n.UnitFinderBoundingBox.To.Y >= newBB.From.Y && collides(n) {
				return n
			}
		}
	} else if movement.X > 0 {
		arr := f.st.UnitFinderX
		for i := unitFinderLowerBound(arr, bb.To.X); i != len(arr); i++ {
			if arr[i].Value > newBB.To.X {
				break
			}
			n := arr[i].U
			if n.UnitFinderBoundingBox.From.Y <= newBB.To.Y && n.UnitFinderBoundingBox.To.Y >= newBB.From.Y && collides(n) {
				return n
			}
		}
	}
	if movement.Y < 0 {
		arr := f.st.UnitFinderY
		for i := unitFinderUpperBound(arr, bb.From.Y); i != 0; {
			i--
			if arr[i].Value < newBB.From.Y {
				break
			}
			n := arr[i].U
			if n.UnitFinderBoundingBox.From.X <= newBB.To.X && n.UnitFinderBoundingBox.To.X >= newBB.From.X && collides(n) {
				return n
			}
		}
	} else if movement.Y > 0 {
		arr := f.st.UnitFinderY
		for i := unitFinderLowerBound(arr, bb.To.Y); i != len(arr); i++ {
			if arr[i].Value > newBB.To.Y {
				break
			}
			n := arr[i].U
			if n.UnitFinderBoundingBox.From.X <= newBB.To.X && n.UnitFinderBoundingBox.To.X >= newBB.From.X && collides(n) {
				return n
			}
		}
	}
	return nil
}

func (f *StateFunctions) checkGroundUnitMovementUnitCollision(u *Unit, ems *ExecuteMovementStruct) *Unit {
	if f.uGroundUnit(u) {
		return f.checkUnitMovementUnitCollision(u, ems)
	}
	return nil
}

func (f *StateFunctions) updateRepulseDirection(u *Unit) {
	repulseDir := u.RepulseDirection
	if !f.unitIsAtMoveTarget(&u.Flingy) {
		targetDir := f.xyDirection(XySub(u.MoveTarget.Pos, u.Position))
		index := f.directionIndex(repulseDir.Sub(targetDir))
		adjust := 32
		if index >= 128 {
			adjust = -64
		}
		repulseDir = targetDir.Add(f.directionFromIndex(index%32 + uint(adjust)))
	}
	bb := f.unitSpriteBoundingBox(u)
	zero := DirectionTFromRaw(0)
	if bb.From.X < 32 {
		if repulseDir.Less(zero) {
			repulseDir = repulseDir.Neg()
		}
		u.RepulseFlags &= 0xf
	} else if uint(bb.To.X) > f.gameSt.MapWidth-32 {
		if repulseDir.Greater(zero) {
			repulseDir = repulseDir.Neg()
		}
		u.RepulseFlags &= 0xf
	}
	if bb.From.Y < 32 {
		if repulseDir.Add(DirectionTFromRaw(64)).GreaterEqual(zero) {
			repulseDir = repulseDir.Neg().Sub(DirectionTFromRaw(-128))
		}
		u.RepulseFlags &= 0xf
	} else if uint(bb.To.Y) > f.gameSt.MapHeight-32 {
		if repulseDir.Add(DirectionTFromRaw(64)).Less(zero) {
			repulseDir = repulseDir.Neg().Sub(DirectionTFromRaw(-128))
		}
		u.RepulseFlags &= 0xf
	}
	u.RepulseDirection = repulseDir
}

func repulseAdjustComponent(v Fp8, adjust Fp8) Fp8 {
	zero := Fp8FromRaw(0)
	if v.Greater(adjust.MulInt(4)) {
		return v.Sub(adjust)
	} else if v.Greater(zero) {
		return v.Add(adjust)
	} else if v.Less(adjust.Neg().MulInt(4)) {
		return v.Add(adjust)
	} else if v.Less(zero) {
		return v.Sub(adjust)
	}
	return v
}

func (f *StateFunctions) applyRepulseField(u *Unit, ems *ExecuteMovementStruct) bool {
	if !f.uCanMove(u) {
		return false
	}
	if f.utBuilding(u.UnitType) {
		return false
	}
	if f.unitIs(u.UnitType, UnitTypesProtossInterceptor) {
		return false
	}
	if f.uOrderNotInterruptible(u) {
		return false
	}
	if f.uIscriptNobrk(u) {
		return false
	}
	strength := int32(f.st.RepulseField[u.RepulseIndex])
	if strength <= 1 {
		u.RepulseFlags &^= 7
		return false
	}
	if u.RepulseFlags&7 != 0 {
		u.RepulseFlags--
	} else {
		u.RepulseFlags |= 7
		f.updateRepulseDirection(u)
	}
	repulse := XyFp8DivInt(f.directionXYOne(u.RepulseDirection), 2)
	if f.unitIsAtMoveTarget(&u.Flingy) {
		repulse = XyFp8MulInt(repulse, strength/4+1)
		repulse = XyFp8DivInt(repulse, 2)
	}
	index := uint(u.RepulseFlags >> 4)
	adjust := XyFp8{X: Fp8FromRaw(repulseAdjustTable[index*2]), Y: Fp8FromRaw(repulseAdjustTable[index*2+1])}
	repulse.X = repulseAdjustComponent(repulse.X, adjust.X)
	repulse.Y = repulseAdjustComponent(repulse.Y, adjust.Y)
	ems.ExactPosition = XyFp8Add(ems.ExactPosition, repulse)
	ems.Position = f.toXy(ems.ExactPosition)
	return true
}

func (f *StateFunctions) finishFlingyMovement(u *Flingy, ems *ExecuteMovementStruct) bool {
	moved := u.Position != ems.Position
	u.MovementFlags = ems.PostMovementFlags
	u.NextSpeed = u.CurrentSpeed
	u.Position = ems.Position
	u.ExactPosition = ems.ExactPosition
	f.moveSprite(u.Sprite, u.Position)
	f.updateUnitHeading(u, u.CurrentVelocityDirection)
	f.runMovementAnimation(u.Sprite, ems)
	return moved
}

func (f *StateFunctions) runMovementAnimation(sprite *Sprite, ems *ExecuteMovementStruct) {
	if ems.StoppingMovement {
		if !f.sFlag(sprite, SpriteFlagIscriptNobrk) {
			f.spriteRunAnim(sprite, IscriptAnimsWalkingToIdle)
		}
	} else if ems.StartingMovement {
		f.spriteRunAnim(sprite, IscriptAnimsWalking)
	}
}

func (f *StateFunctions) finishUnitMovement(u *Unit, ems *ExecuteMovementStruct) bool {
	prevPos := u.Position
	if !f.finishFlingyMovement(&u.Flingy, ems) {
		return false
	}
	if f.tileIndex(prevPos) != f.tileIndex(u.Position) {
		ems.RefreshVision = true
	}
	f.unitFinderReinsert(u)
	return true
}

func (f *StateFunctions) isMovingAlongPath(u *Unit) bool {
	if u.Path == nil {
		return false
	}
	if u.Path.Next == u.Sprite.Position {
		return false
	}
	if u.Path.Next == u.MoveTarget.Pos {
		return true
	}
	if u.FlingyMovementType != 0 {
		return true
	}
	remainingDistance := f.xyLength(XySub(u.Path.Next, u.Sprite.Position))
	if remainingDistance > 16 {
		return true
	}
	remainingTurn := Fp8Extend(u.DesiredVelocityDirection.Sub(u.NextVelocityDirection)).Abs()
	if remainingTurn.RawValue <= remainingDistance*2 {
		return true
	}
	if u.Path.CurrentShortPathIndex >= uint(len(u.Path.ShortPath))-1 {
		if f.getRegionAt(u.Sprite.Position) != u.Path.LongPath[u.Path.CurrentLongPathIndex] {
			return true
		}
	}
	return false
}

func (f *StateFunctions) setUnitImmovable(u *Unit) {
	f.setNextSpeed(&u.Flingy, Fp8FromRaw(0))
	f.stopUnit(u)
	f.setUnitMoveTarget(u, u.Sprite.Position)
	f.uSetStatusFlag(u, UnitStatusFlagImmovable)
	u.UserActionFlags &^= 2
}

func (f *StateFunctions) updateUnitPathingCollision(u *Unit) {
	if ^u.MovementFlags&2 != 0 {
		return
	}
	if u.PathingCollisionCounter <= 2 {
		return
	}
	if f.uOrderNotInterruptible(u) {
		return
	}
	if f.uIscriptNobrk(u) {
		return
	}
	u.MovementFlags &^= 2
	f.spriteRunAnim(u.Sprite, IscriptAnimsWalkingToIdle)
}

func (f *StateFunctions) cardinalDirectionFromTo(fromU *Unit, toU *Unit) int32 {
	aBB := f.unitSpriteInnerBoundingBox(toU)
	bBB := f.unitSpriteInnerBoundingBox(fromU)
	if aBB.To.X < bBB.From.X {
		return 3
	}
	if aBB.From.X > bBB.To.X {
		return 1
	}
	if aBB.To.Y < bBB.From.Y {
		return 0
	}
	if aBB.From.Y > bBB.To.Y {
		return 2
	}
	upDistance := absInt32(aBB.To.Y - bBB.From.Y)
	rightDistance := absInt32(aBB.From.X - bBB.To.X)
	downDistance := absInt32(aBB.From.Y - bBB.To.Y)
	leftDistance := absInt32(aBB.To.X - bBB.From.X)
	distances := [4]int32{rightDistance, leftDistance, downDistance, upDistance}
	r := [4]int32{1, 3, 2, 0}
	return r[getBestScore(distances[:], func(v int32) int32 { return v })]
}

func (f *StateFunctions) collisionGetSlideFreeDirection(u *Unit, collisionUnit *Unit, slideFreeDirection *DirectionT) bool {
	*slideFreeDirection = DirectionTFromRaw(-1)
	if f.usHidden(&collisionUnit.Thingy) {
		return false
	}
	targetDir := f.xyDirection(XySub(u.NextMovementWaypoint, u.Sprite.Position))
	targetDirQuadrant := f.directionIndex(targetDir) / 64
	dirErr := Fp8Extend(targetDir.Sub(u.CurrentVelocityDirection)).Abs()
	if dirErr.GreaterEqual(Fp8FromRaw(80)) {
		return false
	}
	leftX := collisionUnit.UnitFinderBoundingBox.From.X - u.UnitType.Dimensions.To.X - 1
	rightX := collisionUnit.UnitFinderBoundingBox.To.X + u.UnitType.Dimensions.From.X + 1
	upY := collisionUnit.UnitFinderBoundingBox.From.Y - u.UnitType.Dimensions.To.Y - 1
	downY := collisionUnit.UnitFinderBoundingBox.To.Y + u.UnitType.Dimensions.From.Y + 1
	dir0 := DirectionTFromRaw(0)
	dir64 := DirectionTFromRaw(64)
	dirNeg64 := DirectionTFromRaw(-64)
	dirNeg128 := DirectionTFromRaw(-128)
	var targetPos Xy
	switch f.cardinalDirectionFromTo(u, collisionUnit) {
	case 0:
		targetPos.Y = downY
		if targetDirQuadrant == 3 {
			*slideFreeDirection = dirNeg64
		} else if targetDirQuadrant == 0 {
			*slideFreeDirection = dir64
		}
	case 1:
		targetPos.X = leftX
		if targetDirQuadrant == 0 {
			*slideFreeDirection = dir0
		} else if targetDirQuadrant == 1 {
			*slideFreeDirection = dirNeg128
		}
	case 2:
		targetPos.Y = upY
		if targetDirQuadrant == 1 {
			*slideFreeDirection = dir64
		} else if targetDirQuadrant == 2 {
			*slideFreeDirection = dirNeg64
		}
	case 3:
		targetPos.X = rightX
		if targetDirQuadrant == 2 {
			*slideFreeDirection = dirNeg128
		} else if targetDirQuadrant == 3 {
			*slideFreeDirection = dir0
		}
	}
	for i := 0; i != 2; i++ {
		switch *slideFreeDirection {
		case dir0:
			targetPos.Y = upY
			if blocked, _ := f.isBlocked(u, targetPos); !blocked {
				return true
			}
			*slideFreeDirection = dirNeg128
		case dirNeg128:
			targetPos.Y = downY
			if blocked, _ := f.isBlocked(u, targetPos); !blocked {
				return true
			}
			*slideFreeDirection = dir0
		case dir64:
			targetPos.X = rightX
			if blocked, _ := f.isBlocked(u, targetPos); !blocked {
				return true
			}
			*slideFreeDirection = dirNeg64
		case dirNeg64:
			targetPos.X = leftX
			if blocked, _ := f.isBlocked(u, targetPos); !blocked {
				return true
			}
			*slideFreeDirection = dir64
		}
	}
	return false
}

func (f *StateFunctions) resetMovementState(u *Unit) {
	f.freeUnitPath(u)
	u.MovementState = MovementStatesUMInit
	if u.Sprite.ElevationLevel < 12 {
		u.PathingFlags |= 1
	} else {
		u.PathingFlags &^= 1
	}
}

func (f *StateFunctions) movementUMInit(u *Unit, ems *ExecuteMovementStruct) bool {
	u.PathingFlags &^= 1 | 2
	if u.Sprite.ElevationLevel < 12 {
		u.PathingFlags |= 1
	}
	u.TerrainNoCollisionBounds = Rect{}
	nextState := int32(MovementStatesUMLump)
	if !f.utTurret(u.UnitType) && f.uIscriptNobrk(u) {
		nextState = MovementStatesUMInitSeq
	} else if u.Sprite == nil || f.unitDead(u) {
		nextState = MovementStatesUMLump
	} else if f.uInBunker(u) {
		nextState = MovementStatesUMBunker
	} else if f.usHidden(&u.Thingy) {
		if f.uMovementFlag(&u.Flingy, 2) || !f.unitIsAtMoveTarget(&u.Flingy) {
			f.setUnitImmovable(u)
			f.updateUnitMovementValues(&u.Flingy, ems)
			f.finishUnitMovement(u, ems)
		}
		nextState = MovementStatesUMHidden
	} else if f.uBurrowed(u) {
		nextState = MovementStatesUMLump
	} else if f.uCanMove(u) {
		if u.PathingFlags&1 != 0 {
			nextState = MovementStatesUMAtRest
		} else {
			nextState = MovementStatesUMFlyer
		}
	} else if f.uCanTurn(u) {
		if f.utTurret(u.UnitType) {
			nextState = MovementStatesUMTurret
		} else {
			nextState = MovementStatesUMBldgTurret
		}
	} else if u.PathingFlags&1 != 0 && (f.uMovementFlag(&u.Flingy, 2) || !f.unitIsAtMoveTarget(&u.Flingy)) {
		nextState = MovementStatesUMLumpWannabe
	}
	u.MovementState = nextState
	return true
}

func (f *StateFunctions) stopBlockedMovement(u *Unit, ems *ExecuteMovementStruct) {
	f.setNextSpeed(&u.Flingy, Fp8FromRaw(0))
	u.MovementFlags = ems.PostMovementFlags
	f.updateUnitHeading(&u.Flingy, u.CurrentVelocityDirection)
	f.runMovementAnimation(u.Sprite, ems)
	f.setFlingyMoveTarget(&u.Flingy, u.Sprite.Position, nil)
	f.stopUnit(u)
}

func (f *StateFunctions) decayPathingCollisionCounter(u *Unit) {
	if u.PathingCollisionCounter != 0 {
		if u.PathingCollisionCounter > 2 {
			u.PathingCollisionCounter = 2
		} else {
			u.PathingCollisionCounter--
		}
	}
}

func (f *StateFunctions) movementUMAtRest(u *Unit, ems *ExecuteMovementStruct) bool {
	if !f.unitIsAtMoveTarget(&u.Flingy) {
		f.decayPathingCollisionCounter(u)
	} else {
		u.PathingCollisionCounter = 0
	}
	goToNextWaypoint := func() bool {
		if f.uMovementFlag(&u.Flingy, 4) {
			return true
		}
		if f.unitIsAtMoveTarget(&u.Flingy) {
			if f.uMovementFlag(&u.Flingy, 2) {
				return true
			}
			if u.Position != u.NextTargetWaypoint {
				dir := f.xyDirection(XySub(u.NextTargetWaypoint, u.Position))
				if u.Heading != dir {
					return true
				}
				if u.NextVelocityDirection != dir {
					return true
				}
			}
		}
		return false
	}
	goingToNextWaypoint := false
	if goToNextWaypoint() {
		goingToNextWaypoint = true
		f.updateUnitMovementValues(&u.Flingy, ems)
		if f.checkGroundUnitMovementUnitCollision(u, ems) != nil || f.checkUnitMovementTerrainCollisionEMS(u, ems) {
			f.stopBlockedMovement(u, ems)
		} else {
			f.finishUnitMovement(u, ems)
		}
	}
	if f.uCollision(u) && f.uGroundUnit(u) {
		u.MovementState = MovementStatesUMCheckIllegal
		return false
	}
	if !f.unitIsAtMoveTarget(&u.Flingy) && !f.uMovementFlag(&u.Flingy, 4) {
		u.MovementState = MovementStatesUMStartPath
		return true
	}
	if !goingToNextWaypoint {
		u.NextSpeed = Fp8FromRaw(0)
		if u.CurrentSpeed != Fp8FromRaw(0) {
			u.CurrentSpeed = Fp8FromRaw(0)
			u.Velocity = XyFp8{}
		}
		f.setNextTargetWaypoint(&u.Flingy, u.Sprite.Position)
		u.MovementState = MovementStatesUMDormant
	}
	return false
}

func (f *StateFunctions) getUniqueSidedPositionsWithinBounds(from *Xy, to *Xy, bounds Rect) bool {
	getFlags := func(pos Xy) int32 {
		var r int32
		if pos.Y < bounds.From.Y {
			r |= 8
		}
		if pos.Y > bounds.To.Y {
			r |= 4
		}
		if pos.X > bounds.To.X {
			r |= 2
		}
		if pos.X < bounds.From.X {
			r |= 1
		}
		return r
	}

	initialFrom := *from
	initialTo := *to

	fromFlags := getFlags(*from)
	toFlags := getFlags(*to)

	for {
		if fromFlags|toFlags == 0 {
			return true
		}
		if fromFlags&toFlags != 0 {
			*from = initialFrom
			*to = initialTo
			return false
		}
		visit := func(flags int32) Xy {
			var pos Xy
			if flags&8 != 0 {
				pos.X = from.X + (bounds.From.Y-from.Y)*(to.X-from.X)/(to.Y-from.Y)
				pos.Y = bounds.From.Y
			} else if flags&4 != 0 {
				pos.X = from.X + (bounds.To.Y-from.Y)*(to.X-from.X)/(to.Y-from.Y)
				pos.Y = bounds.To.Y
			} else if flags&2 != 0 {
				pos.X = bounds.To.X
				pos.Y = from.Y + (bounds.To.X-from.X)*(to.Y-from.Y)/(to.X-from.X)
			} else {
				pos.X = bounds.From.X
				pos.Y = from.Y + (bounds.From.X-from.X)*(to.Y-from.Y)/(to.X-from.X)
			}
			return pos
		}
		if fromFlags != 0 {
			*from = visit(fromFlags)
			fromFlags = getFlags(*from)
		} else {
			*to = visit(toFlags)
			toFlags = getFlags(*to)
		}
	}
}

func (f *StateFunctions) movementUMCheckIllegal(u *Unit, ems *ExecuteMovementStruct) bool {
	f.uUnsetStatusFlag(u, UnitStatusFlagCollision)
	checkIllegal := func() bool {
		if !f.uGroundUnit(u) {
			return false
		}
		blocked, blockingUnit := f.isBlocked(u, u.Sprite.Position)
		if !blocked {
			return false
		}
		if f.uOrderNotInterruptible(u) || f.uIscriptNobrk(u) || f.uMovementFlag(&u.Flingy, 8) {
			f.uSetStatusFlag(u, UnitStatusFlagCollision)
			return false
		}
		moveTo := u.Sprite.Position

		if f.unitTypeCanFitAt(u.UnitType, moveTo) {
			if blockingUnit != nil {
				f.checkUnitMovementTerrainCollision(u, Xy{})
				moveTo = f.checkIllegalFindMoveTo(u, blockingUnit)
			}
		} else {
			moveTo = f.checkIllegalMoveToWalkable(u, moveTo)
		}

		moveTo = f.restrictMoveTargetToValidBounds(u.UnitType, moveTo)
		if moveTo != u.Sprite.Position {
			f.freeUnitPath(u)
			if f.unitIsAtMoveTarget(&u.Flingy) || u.MovementFlags&4 != 0 {
				f.setUnitMoveTarget(u, moveTo)
			} else {
				f.setNextTargetWaypoint(&u.Flingy, moveTo)
				if u.NextMovementWaypoint != moveTo {
					u.NextMovementWaypoint = moveTo
					f.uSetMovementFlag(&u.Flingy, 1)
				}
			}
			u.Path = f.createSingleStepPath(u.Sprite.Position, moveTo)
			u.TerrainNoCollisionBounds = Rect{}
			return true
		}
		f.uSetStatusFlag(u, UnitStatusFlagCollision)
		return false
	}
	if checkIllegal() {
		u.PathingFlags |= 2
		u.MovementState = MovementStatesUMMoveToLegal
		return false
	}
	u.PathingFlags &^= 2 | 4
	if f.unitIsAtMoveTarget(&u.Flingy) || f.uMovementFlag(&u.Flingy, 4) {
		u.MovementState = MovementStatesUMAtRest
	} else {
		u.MovementState = MovementStatesUMAnotherPath
	}
	return true
}

func (f *StateFunctions) checkIllegalFindMoveTo(u *Unit, blockingUnit *Unit) Xy {
	moveTo := u.Sprite.Position
	otherUnitIsMoving := false
	if f.uCanMove(blockingUnit) && (!f.unitIsAtMoveTarget(&blockingUnit.Flingy) || blockingUnit.MovementState == MovementStatesUMCheckIllegal || blockingUnit.MovementState == MovementStatesUMMoveToLegal) {
		otherUnitIsMoving = true
		if f.lcgRandRange(50, 0, 31) < 24 {
			return moveTo
		}
	} else if bestPos, found := f.checkIllegalFindFreeRectPos(u, blockingUnit); found {
		return bestPos
	}
	dir := u.NextVelocityDirection.Add(DirectionTFromRaw(16).MulInt(f.lcgRandRange(50, -3, 3)))
	length := Fp8Integer(f.lcgRandRange(50, 2, 4) * 4)
	if !otherUnitIsMoving && f.uGroundedBuilding(blockingUnit) {
		length = length.MulInt(3)
	}
	moveTo = XyAdd(u.Sprite.Position, f.toXy(f.directionXY(dir, length)))
	if blocked, _ := f.isBlocked(u, moveTo); !blocked && f.isReachable(u.Sprite.Position, moveTo) {
		return moveTo
	}
	if f.xyLength(XySub(u.MoveTarget.Pos, u.Sprite.Position)) <= 32 || f.lcgRandRange(50, 0, 31) >= 24 {
		moveTo = u.Sprite.Position
		moveTo.X += 16 * f.lcgRandRange(50, -2, 2)
		moveTo.Y += 16 * f.lcgRandRange(50, -2, 2)
	} else {
		dir = f.xyDirection(XySub(u.MoveTarget.Pos, u.Sprite.Position)).Add(DirectionTFromRaw(16).MulInt(f.lcgRandRange(50, -1, 1)))
		length = Fp8Integer(8 + f.lcgRandRange(50, 0, 2)*4)
		if !otherUnitIsMoving && f.uGroundedBuilding(blockingUnit) {
			length = length.MulInt(3)
		}
		moveTo = XyAdd(u.Sprite.Position, f.toXy(f.directionXY(dir, length)))
	}
	if !f.isInInnerMapBounds(f.unitInnerBoundingBox(u, moveTo)) || !f.isReachable(u.Sprite.Position, moveTo) {
		return u.Sprite.Position
	}
	blocked, blockedBy := f.isBlocked(u, moveTo)
	if blocked {
		if blockedBy == nil {
			return u.Sprite.Position
		} else if otherUnitIsMoving && !f.uCanMove(blockedBy) {
			return u.Sprite.Position
		}
	}
	return moveTo
}

func (f *StateFunctions) checkIllegalFindFreeRectPos(u *Unit, blockingUnit *Unit) (Xy, bool) {
	superBB := func(a *Unit, b *Unit) Rect {
		r := f.unitSpriteInnerBoundingBox(a)
		r.From.X -= b.UnitType.Dimensions.To.X + 1
		r.From.Y -= b.UnitType.Dimensions.To.Y + 1
		r.To.X += b.UnitType.Dimensions.From.X + 1
		r.To.Y += b.UnitType.Dimensions.From.Y + 1
		return r
	}

	initialRect := superBB(blockingUnit, u)
	initialRects := [4]Rect{initialRect, initialRect, initialRect, initialRect}
	initialRects[0].To.Y = initialRect.From.Y
	initialRects[1].From.X = initialRect.To.X
	initialRects[2].From.Y = initialRect.To.Y
	initialRects[3].To.X = initialRect.From.X

	findRects := func(rects []Rect, bounds Rect) StaticVector[Rect, [4]Rect] {
		var r StaticVector[Rect, [4]Rect]
		for _, v := range rects {
			if f.getUniqueSidedPositionsWithinBounds(&v.From, &v.To, bounds) {
				r.PushBack(Rect{From: v.From, To: v.To})
			}
		}
		return r
	}

	vec := findRects(initialRects[:], u.TerrainNoCollisionBounds)

	findBB := initialRect
	findBB.From.X -= u.UnitType.Dimensions.From.X
	findBB.From.Y -= u.UnitType.Dimensions.From.Y
	findBB.To.X += u.UnitType.Dimensions.To.X
	findBB.To.Y += u.UnitType.Dimensions.To.Y
	for n := range f.findUnitsNoexpand(findBB) {
		if f.pathfinderUnitCanCollideWith(u, n, nil, true) {
			bounds := superBB(n, u)
			vec = findRects(vec.Values(), bounds)
		}
	}

	bestPos := u.Sprite.Position
	bestDistance := int32(9999)
	for _, v := range vec.Values() {
		pos := u.Sprite.Position
		if pos.X > v.To.X {
			pos.X = v.To.X
		} else if pos.X < v.From.X {
			pos.X = v.From.X
		}
		if pos.Y > v.To.Y {
			pos.Y = v.To.Y
		} else if pos.Y < v.From.Y {
			pos.Y = v.From.Y
		}
		distance := f.xyLength(XySub(pos, u.Sprite.Position))
		if distance < bestDistance {
			bestDistance = distance
			bestPos = pos
		}
	}
	if bestDistance != 9999 {
		if f.unitTypeCanFitAt(u.UnitType, bestPos) {
			if blocked, _ := f.isBlocked(u, bestPos); !blocked && f.isReachable(u.Sprite.Position, bestPos) {
				return bestPos, true
			}
		}
	}
	return Xy{}, false
}

func (f *StateFunctions) checkIllegalMoveToWalkable(u *Unit, moveTo Xy) Xy {
	r := f.getRegionAt(u.Sprite.Position)
	if r.Walkable() {
		dir := f.xyDirection(XySub(f.toXy(r.Center), u.Sprite.Position)).Add(DirectionTFromRaw(16).MulInt(f.lcgRandRange(50, -2, 2)))
		length := Fp8Integer(f.lcgRandRange(50, 2, 4) * 8)
		centerLength := Fp8Integer(f.xyLength(XySub(f.toXy(r.Center), u.Sprite.Position)))
		if length.Greater(centerLength.Sub(Fp8FromRaw(1))) {
			length = centerLength.Sub(Fp8FromRaw(1))
		}
		moveTo = XyAdd(u.Sprite.Position, f.toXy(f.directionXY(dir, length)))
		if !f.isReachable(u.Sprite.Position, moveTo) {
			moveTo = u.Sprite.Position
		}
		return moveTo
	}
	bestDistance := int32(math.MaxInt32)
	var bestRegion *Region
	var lastCenterPos Xy
	for _, n := range r.WalkableNeighbors {
		lastCenterPos = f.toXy(n.Center)
		distance := f.xyLength(XySub(f.toXy(n.Center), u.Sprite.Position))
		if distance < bestDistance {
			bestDistance = distance
			bestRegion = n
		}
	}
	if bestRegion != nil {
		maybeOutside := false
		if u.Sprite.Position.X+32 < bestRegion.Area.From.X {
			maybeOutside = true
		}
		if u.Sprite.Position.X-32 >= bestRegion.Area.To.X {
			maybeOutside = true
		}
		if u.Sprite.Position.Y+32 < bestRegion.Area.From.Y {
			maybeOutside = true
		}
		if u.Sprite.Position.Y-32 >= bestRegion.Area.To.Y {
			maybeOutside = true
		}
		if maybeOutside {
			dir := f.xyDirection(XySub(lastCenterPos, u.Sprite.Position))
			return XyAdd(u.Sprite.Position, f.toXy(f.directionXYInt(dir, 64)))
		}
		moveTo = f.pathfinderAdjustDestination(bestRegion, u.Sprite.Position)
		if moveTo == u.Sprite.Position {
			moveTo = XySub(u.Sprite.Position, Xy{X: 0, Y: 64})
		}
		return moveTo
	}
	if len(r.NonWalkableNeighbors) != 0 {
		randomIndex := uint(f.lcgRandRange(50, 0, int32(len(r.NonWalkableNeighbors))-1))
		n := r.NonWalkableNeighbors[randomIndex]
		length := f.xyLength(XySub(f.toXy(n.Center), u.Sprite.Position))
		if length > 64 {
			length = 64
		}
		dir := f.xyDirection(XySub(f.toXy(n.Center), u.Sprite.Position))
		moveTo = XyAdd(u.Sprite.Position, f.toXy(f.directionXYInt(dir, length)))
	}
	return moveTo
}

func (f *StateFunctions) movementUMDormant(u *Unit, ems *ExecuteMovementStruct) bool {
	rest := false
	if f.uCollision(u) && f.uGroundUnit(u) {
		rest = true
	}
	if !f.unitIsAtMoveTarget(&u.Flingy) {
		rest = true
	}
	if u.Position != u.NextTargetWaypoint {
		rest = true
	}
	if rest {
		u.MovementState = MovementStatesUMAtRest
		return true
	}
	return false
}

func (f *StateFunctions) movementUMTurret(u *Unit, ems *ExecuteMovementStruct) bool {
	ems.RefreshVision = false
	f.setUnitMoveTarget(u, u.Sprite.Position)
	dirError := u.DesiredVelocityDirection.Sub(u.Heading)
	if dirError != DirectionTFromRaw(-128) {
		if dirError.GreaterEqual(DirectionTFromRaw(-10)) && dirError.LessEqual(DirectionTFromRaw(10)) {
			f.uUnsetMovementFlag(&u.Flingy, 1)
		}
	}
	if f.uStatusFlag(u, 0x2000000) {
		f.setMovementFlags(&u.Flingy, ems)
	} else {
		f.updateCurrentVelocityDirectionTowardsWaypoint(&u.Flingy)
		f.setMovementFlags(&u.Flingy, ems)
		f.updateUnitHeading(&u.Flingy, u.CurrentVelocityDirection)
	}
	return false
}

func (f *StateFunctions) movementUMFlyer(u *Unit, ems *ExecuteMovementStruct) bool {
	if u.Sprite.Position != u.MoveTarget.Pos {
		moveTarget := f.restrictUnitPosToBounds(u.MoveTarget.Pos, u.UnitType, f.mapBounds())
		if moveTarget != u.MoveTarget.Pos {
			f.setFlingyMoveTarget(&u.Flingy, moveTarget, nil)
			f.uSetStatusFlag(u, UnitStatusFlagImmovable)
		}
	}
	f.updateUnitMovementValues(&u.Flingy, ems)
	beingRepulsed := f.applyRepulseField(u, ems)
	ems.Position = f.restrictUnitPosToBounds(ems.Position, u.UnitType, f.mapBounds())
	f.finishUnitMovement(u, ems)
	if f.uCanMove(u) && !f.utBuilding(u.UnitType) && u.UnitType.ID != UnitTypesProtossInterceptor {
		index := f.repulseIndex(u.Position)
		if index != u.RepulseIndex {
			f.decrementRepulseField(u)
			f.incrementRepulseField(u)
		}
		if beingRepulsed {
			if max(absInt32(u.MoveTarget.Pos.X-u.Position.X), absInt32(u.MoveTarget.Pos.Y-u.Position.Y)) < 24 {
				u.MoveTarget.Pos = u.Position
				u.NextMovementWaypoint = u.Position
			}
		}
	}
	return false
}

func (f *StateFunctions) movementUMAnotherPath(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.unitPathTo(u, u.MoveTarget.Pos, f.st.ConsiderCollisionWithUnitBug, false) {
		u.MovementState = MovementStatesUMFollowPath
	} else {
		u.MovementState = MovementStatesUMFailedPath
	}
	return false
}

func (f *StateFunctions) movementUMStartPath(u *Unit, ems *ExecuteMovementStruct) bool {
	if !f.unitPathTo(u, u.MoveTarget.Pos, f.st.ConsiderCollisionWithUnitBug, false) {
		u.MovementState = MovementStatesUMFailedPath
		return false
	}
	nextMovementState := int32(MovementStatesUMFollowPath)
	if u.UserActionFlags&2 != 0 {
		u.UserActionFlags &^= 2
		if u.FlingyMovementType == 2 {
			u.Path.Delay = f.lcgRandRange(51, 0, 2)
			nextMovementState = MovementStatesUMUIOrderDelay
		}
	}
	u.MovementState = nextMovementState
	return true
}

func (f *StateFunctions) movementUMFollowPath(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.unitUpdatePathMovementState(u, true) {
		return true
	}
	if !f.unitPathTo(u, u.MoveTarget.Pos, f.st.ConsiderCollisionWithUnitBug, false) {
		u.MovementState = MovementStatesUMAnotherPath
		return true
	}
	speed := u.NextSpeed
	f.updateUnitMovementValues(&u.Flingy, ems)
	if u.FlingyMovementType == 0 {
		speed = u.CurrentSpeed
	}
	collisionUnit := f.checkGroundUnitMovementUnitCollision(u, ems)
	if collisionUnit != nil {
		if f.followPathMustFixCollision(u, ems) {
			u.Path.LastCollisionUnit = f.getUnitID(collisionUnit)
			u.Path.LastCollisionSpeed = speed
			u.MovementState = MovementStatesUMFixCollision
			return true
		}
	} else if f.checkUnitMovementTerrainCollisionEMS(u, ems) {
		u.Path.LastCollisionSpeed = speed
		u.MovementState = MovementStatesUMFixTerrain
		return true
	}
	if u.Sprite.Position != ems.Position {
		f.decayPathingCollisionCounter(u)
	}
	f.finishUnitMovement(u, ems)
	if f.unitIsAtMoveTarget(&u.Flingy) {
		u.MovementState = MovementStatesUMAtMoveTarget
	} else if u.Path != nil {
		if u.Path.Delay != 0 {
			u.Path.Delay--
		} else {
			u.Path.Delay = 30
			u.MovementState = MovementStatesUMScoutPath
		}
	}
	return false
}

func (f *StateFunctions) followPathMustFixCollision(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.checkUnitMovementTerrainCollisionEMS(u, ems) {
		return true
	}
	if u.Velocity.X.Less(Fp8FromRaw(32)) && u.Velocity.Y.Less(Fp8FromRaw(32)) {
		return true
	}
	if ems.Speed != u.CurrentSpeed {
		return true
	}
	pos := ems.Position
	exactPos := ems.ExactPosition
	ems.ExactPosition = XyFp8Add(u.ExactPosition, XyFp8DivInt(u.Velocity, 2))
	ems.Position = f.toXy(ems.ExactPosition)
	if f.checkUnitMovementUnitCollision(u, ems) != nil {
		ems.ExactPosition = XyFp8Add(u.ExactPosition, XyFp8DivInt(u.Velocity, 4))
		ems.Position = f.toXy(ems.ExactPosition)
		if f.checkUnitMovementUnitCollision(u, ems) != nil {
			ems.Position = pos
			ems.ExactPosition = exactPos
			return true
		}
	}
	return false
}

func (f *StateFunctions) movementUMScoutPath(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.unitUpdatePathMovementState(u, true) {
		return true
	}
	u.MovementState = MovementStatesUMFollowPath
	return true
}

func (f *StateFunctions) freePathAndClear(u *Unit) {
	if u.Path != nil {
		f.freePath(u.Path)
		u.Path = nil
	}
}

func (f *StateFunctions) movementUMAtMoveTarget(u *Unit, ems *ExecuteMovementStruct) bool {
	f.freePathAndClear(u)
	if u.NextMovementWaypoint != u.MoveTarget.Pos {
		u.NextMovementWaypoint = u.MoveTarget.Pos
	}
	if !f.uGroundUnit(u) || f.uMovementFlag(&u.Flingy, 4) {
		u.MovementState = MovementStatesUMAtRest
	} else {
		u.MovementState = MovementStatesUMCheckIllegal
	}
	return true
}

func (f *StateFunctions) movementUMNewMoveTarget(u *Unit, ems *ExecuteMovementStruct) bool {
	u.Path.StateFlags &^= 1
	if !f.unitUpdatePathMovementState(u, true) {
		f.freePathAndClear(u)
		u.MovementState = MovementStatesUMRepath
	}
	return true
}

func (f *StateFunctions) movementUMRepath(u *Unit, ems *ExecuteMovementStruct) bool {
	collisionUnit := f.st.ConsiderCollisionWithUnitBug
	if u.Path != nil {
		collisionUnit = f.getUnitByID(u.Path.LastCollisionUnit)
		f.freePath(u.Path)
		u.Path = nil
	}
	f.st.ConsiderCollisionWithUnitBug = nil
	if f.unitPathTo(u, u.MoveTarget.Pos, collisionUnit, false) {
		u.MovementState = MovementStatesUMFollowPath
	} else {
		u.MovementState = MovementStatesUMFailedPath
	}
	return true
}

func (f *StateFunctions) movementUMUIOrderDelay(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.unitUpdatePathMovementState(u, true) {
		return true
	}
	if u.Path.Delay == 0 {
		u.MovementState = MovementStatesUMFollowPath
		return true
	}
	u.Path.Delay--
	return false
}

func (f *StateFunctions) incrementPathingCollisionCounter(u *Unit) {
	if u.PathingCollisionCounter < 255 {
		u.PathingCollisionCounter++
	}
}

func (f *StateFunctions) movementUMFixTerrain(u *Unit, ems *ExecuteMovementStruct) bool {
	f.incrementPathingCollisionCounter(u)
	if f.unitUpdatePathMovementState(u, true) {
		return true
	}
	f.setNextSpeed(&u.Flingy, u.Path.LastCollisionSpeed)
	f.updateUnitMovementValues(&u.Flingy, ems)
	if !f.checkUnitMovementTerrainCollisionEMS(u, ems) && f.checkUnitMovementUnitCollision(u, ems) == nil {
		f.finishUnitMovement(u, ems)
		u.MovementFlags |= 1
		u.MovementState = MovementStatesUMFollowPath
		return false
	}
	f.updateUnitPathingCollision(u)
	if u.NextVelocityDirection != f.xyDirection(XySub(u.NextMovementWaypoint, u.Sprite.Position)) {
		f.updateUnitHeading(&u.Flingy, u.CurrentVelocityDirection)
		return false
	}
	slideFreeDirection := DirectionTFromRaw(-1)
	movement := XySub(ems.Position, u.Sprite.Position)
	desiredQuadrant := f.directionIndex(u.CurrentVelocityDirection) / 64
	if ems.Position.X != u.Sprite.Position.X && (ems.Position.Y == u.Sprite.Position.Y || f.checkUnitMovementTerrainCollision(u, Xy{X: movement.X, Y: 0})) {
		if movement.X >= 0 {
			if desiredQuadrant == 0 {
				slideFreeDirection = DirectionTFromRaw(0)
			}
			if desiredQuadrant == 1 {
				slideFreeDirection = DirectionTFromRaw(-128)
			}
		} else {
			if desiredQuadrant == 3 {
				slideFreeDirection = DirectionTFromRaw(0)
			}
			if desiredQuadrant == 2 {
				slideFreeDirection = DirectionTFromRaw(-128)
			}
		}
	} else {
		if movement.Y < 0 {
			if desiredQuadrant == 0 {
				slideFreeDirection = DirectionTFromRaw(64)
			}
			if desiredQuadrant == 3 {
				slideFreeDirection = DirectionTFromRaw(-64)
			}
		} else {
			if desiredQuadrant == 1 {
				slideFreeDirection = DirectionTFromRaw(64)
			}
			if desiredQuadrant == 2 {
				slideFreeDirection = DirectionTFromRaw(-64)
			}
		}
	}
	u.Path.SlideFreeDirection = slideFreeDirection
	u.MovementState = MovementStatesUMTerrainSlide
	return false
}

func (f *StateFunctions) movementUMFailedPath(u *Unit, ems *ExecuteMovementStruct) bool {
	u.PathingCollisionCounter = 10
	f.updateUnitPathingCollision(u)
	u.MovementState = MovementStatesUMRetryPath
	return false
}

func (f *StateFunctions) movementUMTerrainSlide(u *Unit, ems *ExecuteMovementStruct) bool {
	f.incrementPathingCollisionCounter(u)
	if f.unitUpdatePathMovementState(u, true) {
		return true
	}
	move := func(direction DirectionT) {
		ems.PreMovementFlags = u.MovementFlags
		f.setCurrentVelocityDirection(&u.Flingy, direction)
		f.updateCurrentSpeedTowardsWaypoint(&u.Flingy)
		f.setMovementFlags(&u.Flingy, ems)
		f.setMovementValues(&u.Flingy, ems)
		ems.PostMovementFlags = u.MovementFlags
		u.MovementFlags = ems.PreMovementFlags
	}
	u.MovementFlags &^= 1
	nextVelocityDirection := u.NextVelocityDirection
	move(u.Path.SlideFreeDirection)
	u.MovementFlags |= 1
	u.NextVelocityDirection = nextVelocityDirection
	f.setDesiredVelocityDirection(&u.Flingy, u.Path.SlideFreeDirection)
	if f.checkUnitMovementUnitCollision(u, ems) != nil || f.checkUnitMovementTerrainCollisionEMS(u, ems) {
		u.MovementState = MovementStatesUMForceMoveFree
		return false
	}
	f.finishUnitMovement(u, ems)
	nextSpeed := u.NextSpeed
	f.setNextSpeed(&u.Flingy, u.Path.LastCollisionSpeed)
	move(f.xyDirection(XySub(u.NextMovementWaypoint, u.Sprite.Position)))
	f.setNextSpeed(&u.Flingy, nextSpeed)
	if !f.checkUnitMovementTerrainCollisionEMS(u, ems) {
		u.MovementFlags |= 1
		u.MovementState = MovementStatesUMFollowPath
	}
	return false
}

func (f *StateFunctions) movementUMTurnAndStart(u *Unit, ems *ExecuteMovementStruct) bool {
	collisionUnit := f.st.ConsiderCollisionWithUnitBug
	if u.Path != nil {
		collisionUnit = f.getUnitByID(u.Path.LastCollisionUnit)
		f.freePath(u.Path)
		u.Path = nil
		f.st.ConsiderCollisionWithUnitBug = collisionUnit
	}
	if f.unitIsAtMoveTarget(&u.Flingy) || u.MovementFlags&4 != 0 {
		u.MovementState = MovementStatesUMAtMoveTarget
		return true
	}
	f.st.ConsiderCollisionWithUnitBug = nil
	if f.unitPathTo(u, u.MoveTarget.Pos, collisionUnit, false) {
		u.MovementState = MovementStatesUMFaceTarget
		return false
	}
	u.MovementState = MovementStatesUMFailedPath
	return true
}

func (f *StateFunctions) movementUMFaceTarget(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.unitUpdatePathMovementState(u, true) {
		return true
	}
	if f.unitIsAtMoveTarget(&u.Flingy) || u.MovementFlags&4 != 0 {
		u.MovementState = MovementStatesUMAtMoveTarget
		return true
	}
	if u.NextVelocityDirection == f.xyDirection(XySub(u.NextMovementWaypoint, u.Sprite.Position)) {
		if u.Path.StateFlags&1 != 0 {
			u.MovementState = MovementStatesUMNewMoveTarget
		} else {
			u.MovementState = MovementStatesUMFollowPath
		}
		return true
	}
	f.updateUnitMovementValues(&u.Flingy, ems)
	f.updateUnitHeading(&u.Flingy, u.CurrentVelocityDirection)
	return false
}

func (f *StateFunctions) movementUMRetryPath(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.unitUpdatePathMovementState(u, false) {
		return true
	}
	if u.PathingCollisionCounter < 20 {
		u.PathingCollisionCounter++
		return false
	}
	if f.unitPathTo(u, u.MoveTarget.Pos, f.st.ConsiderCollisionWithUnitBug, false) {
		u.MovementState = MovementStatesUMFollowPath
		return true
	}
	f.setUnitImmovable(u)
	u.MovementState = MovementStatesUMAtMoveTarget
	return false
}

func (f *StateFunctions) movementUMInitSeq(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.uStatusFlag(u, UnitStatusFlagIscriptNobrk) {
		return false
	}
	u.MovementState = MovementStatesUMInit
	return true
}

func (f *StateFunctions) movementUMForceMoveFree(u *Unit, ems *ExecuteMovementStruct) bool {
	f.incrementPathingCollisionCounter(u)
	if f.unitIsAtMoveTarget(&u.Flingy) || f.uMovementFlag(&u.Flingy, 4) {
		u.MovementState = MovementStatesUMAtMoveTarget
		return false
	}
	f.setNextSpeed(&u.Flingy, Fp8FromRaw(0))
	f.freeUnitPath(u)
	u.MovementState = MovementStatesUMTurnAndStart
	return true
}

func (f *StateFunctions) movementUMFixCollision(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.unitUpdatePathMovementState(u, false) {
		return true
	}
	f.incrementPathingCollisionCounter(u)
	collisionUnit := f.getUnitByID(u.Path.LastCollisionUnit)
	if collisionUnit == nil || f.usHidden(&collisionUnit.Thingy) || !f.unitCanCollideWith(u, collisionUnit) {
		u.MovementState = MovementStatesUMFollowPath
		return false
	}
	var slideFreeDirection DirectionT
	state := f.fixCollisionState(u, collisionUnit, &slideFreeDirection)
	if f.st.CurrentFrame-u.Path.CreationFrame >= 7 || state < 3 || state > 5 {
		switch state {
		case 1:
			f.setUnitImmovable(u)
			f.uUnsetStatusFlag(u, UnitStatusFlagImmovable)
			u.MovementState = MovementStatesUMAtMoveTarget
		case 2:
			f.setUnitImmovable(u)
			u.MovementState = MovementStatesUMAtMoveTarget
		case 3:
			u.MovementState = MovementStatesUMRepath
		case 4:
			u.MovementState = MovementStatesUMRepathMovers
		case 5:
			u.MovementState = MovementStatesUMTurnAndStart
			f.setNextSpeed(&u.Flingy, Fp8FromRaw(0))
		case 6:
			u.MovementState = MovementStatesUMWaitFree
		case 7:
			u.Path.SlideFreeDirection = slideFreeDirection
			u.MovementState = MovementStatesUMSlidePrep
		case 8:
			u.MovementState = MovementStatesUMGetFree
		}
	} else {
		f.updateUnitPathingCollision(u)
	}
	return false
}

func (f *StateFunctions) fixCollisionState(u *Unit, collisionUnit *Unit, slideFreeDirection *DirectionT) int32 {
	if u.MoveTarget.Unit != nil && f.isIntersecting(f.unitSpriteBoundingBox(f.unitMainUnit(u)), f.unitSpriteBoundingBox(u.MoveTarget.Unit)) {
		return 1
	}
	state := int32(1)
	if f.isIntersecting(f.unitSpriteInnerBoundingBox(collisionUnit), f.unitInnerBoundingBox(u, u.MoveTarget.Pos)) {
		if collisionUnit.PathingCollisionCounter >= 30 {
			state = f.fixCollisionRepathState(u)
		} else if f.isIntersecting(f.unitInnerBoundingBox(collisionUnit, collisionUnit.MoveTarget.Pos), f.unitInnerBoundingBox(u, u.MoveTarget.Pos)) {
			state = f.fixCollisionRepathState(u)
		}
	} else {
		if !f.isIntersecting(f.unitSpriteInnerBoundingBox(collisionUnit), f.unitInnerBoundingBox(u, u.NextMovementWaypoint)) {
			if f.uMovementFlag(&collisionUnit.Flingy, 2) {
				state = f.fixCollisionMovingUnitState(u, collisionUnit)
			}
			if state == 1 {
				if f.collisionGetSlideFreeDirection(u, collisionUnit, slideFreeDirection) {
					state = 7
				} else {
					state = 5
				}
			}
		}
		if state == 1 {
			if collisionUnit.PathingCollisionCounter >= 30 {
				state = 3
			} else if f.isIntersecting(f.unitInnerBoundingBox(collisionUnit, collisionUnit.MoveTarget.Pos), f.unitInnerBoundingBox(u, u.NextMovementWaypoint)) {
				state = 3
			}
		}
	}
	if state == 1 {
		collisionDir := f.xyDirection(XySub(collisionUnit.Sprite.Position, u.Sprite.Position))
		cmpDir := DirectionTFromRaw(0)
		index := f.directionIndex(collisionUnit.NextVelocityDirection)
		if index < 32 {
			cmpDir = DirectionTFromRaw(0)
		} else if index < 96 {
			cmpDir = DirectionTFromRaw(64)
		} else if index < 160 {
			cmpDir = DirectionTFromRaw(-128)
		} else if index < 224 {
			cmpDir = DirectionTFromRaw(-64)
		}
		if Fp8Extend(cmpDir.Sub(collisionDir)).Abs().Greater(Fp8FromRaw(64)) {
			state = 4
		} else {
			state = 6
		}
	}
	return state
}

func (f *StateFunctions) fixCollisionRepathState(u *Unit) int32 {
	if u.MoveTarget.Unit != nil {
		return 3
	}
	return 2
}

func (f *StateFunctions) fixCollisionMovingUnitState(u *Unit, collisionUnit *Unit) int32 {
	index := f.directionIndex(collisionUnit.NextVelocityDirection)
	switch f.cardinalDirectionFromTo(u, collisionUnit) {
	case 0:
		if index <= 64 || index >= 192 {
			return 6
		}
	case 1:
		if index <= 128 {
			return 6
		}
	case 2:
		if index >= 64 && index <= 192 {
			return 6
		}
	case 3:
		if index == 0 || index >= 128 {
			return 6
		}
	}
	if collisionUnit.PathingCollisionCounter <= 2 {
		if !f.unitIsAtMoveTarget(&collisionUnit.Flingy) {
			if collisionUnit.MovementState == MovementStatesUMFollowPath || collisionUnit.MovementState == MovementStatesUMScoutPath {
				return 6
			} else if collisionUnit.MovementState != MovementStatesUMWaitFree {
				return 4
			}
		}
	}
	return 1
}

func (f *StateFunctions) movementUMWaitFree(u *Unit, ems *ExecuteMovementStruct) bool {
	f.incrementPathingCollisionCounter(u)
	if f.unitUpdatePathMovementState(u, true) {
		return true
	}
	f.setNextSpeed(&u.Flingy, u.Path.LastCollisionSpeed)
	f.updateUnitMovementValues(&u.Flingy, ems)
	if f.checkUnitMovementUnitCollision(u, ems) != nil {
		f.updateUnitPathingCollision(u)
		f.setNextSpeed(&u.Flingy, Fp8FromRaw(0))
		if u.NextVelocityDirection != f.xyDirection(XySub(u.NextMovementWaypoint, u.Sprite.Position)) {
			f.updateUnitHeading(&u.Flingy, u.CurrentVelocityDirection)
		}
		if u.PathingCollisionCounter >= 25 {
			if f.lcgRand(51) < 0x7fff/2 {
				u.MovementState = MovementStatesUMForceMoveFree
			} else {
				u.MovementState = MovementStatesUMRepathMovers
			}
		}
	} else if f.checkUnitMovementTerrainCollisionEMS(u, ems) {
		u.MovementState = MovementStatesUMRepath
	} else {
		f.finishUnitMovement(u, ems)
		u.MovementState = MovementStatesUMFollowPath
	}
	return false
}

func (f *StateFunctions) movementUMSlidePrep(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.unitUpdatePathMovementState(u, true) {
		return true
	}
	if u.FlingyMovementType == 0 {
		if u.Path.LastCollisionSpeed.GreaterEqual(Fp8Integer(2)) {
			u.Path.LastCollisionSpeed = u.Path.LastCollisionSpeed.DivInt(2)
		} else if u.Path.LastCollisionSpeed.Greater(Fp8Integer(1)) {
			u.Path.LastCollisionSpeed = Fp8Integer(1)
		}
	}
	f.setCurrentVelocityDirection(&u.Flingy, u.NextVelocityDirection)
	f.setNextSpeed(&u.Flingy, u.Path.LastCollisionSpeed)
	u.MovementState = MovementStatesUMSlideFree
	return true
}

func (f *StateFunctions) movementUMSlideFree(u *Unit, ems *ExecuteMovementStruct) bool {
	f.incrementPathingCollisionCounter(u)
	if f.unitUpdatePathMovementState(u, true) {
		return true
	}
	u.MovementFlags |= 1
	f.setDesiredVelocityDirection(&u.Flingy, u.Path.SlideFreeDirection)
	f.updateUnitHeading(&u.Flingy, u.CurrentVelocityDirection)
	if u.FlingyMovementType == 2 {
		u.CurrentSpeed = u.FlingyTopSpeed
	} else {
		f.updateCurrentSpeedTowardsWaypoint(&u.Flingy)
	}

	move := func() {
		ems.PreMovementFlags = u.MovementFlags
		f.setMovementFlags(&u.Flingy, ems)
		f.setMovementValues(&u.Flingy, ems)
		ems.PostMovementFlags = u.MovementFlags
		u.MovementFlags = ems.PreMovementFlags
	}

	originalSpeed := u.CurrentSpeed
	originalVelocityDirection := u.NextVelocityDirection
	collisionUnit := f.getUnitByID(u.Path.LastCollisionUnit)
	var nextCollisionUnit *Unit
	forceMoveFree := false
	for speed := originalSpeed; speed.Greater(Fp8FromRaw(0)); speed = speed.Sub(Fp8Integer(1)) {
		f.uUnsetMovementFlag(&u.Flingy, 1)
		f.setCurrentVelocityDirection(&u.Flingy, u.Path.SlideFreeDirection)
		f.setNextSpeed(&u.Flingy, minFp8(speed, Fp8Integer(1)))
		move()
		if f.checkUnitMovementUnitCollision(u, ems) != nil || f.checkUnitMovementTerrainCollisionEMS(u, ems) {
			forceMoveFree = true
			break
		}
		f.finishUnitMovement(u, ems)
		u.NextVelocityDirection = originalVelocityDirection
		f.updateCurrentVelocityDirectionTowardsWaypoint(&u.Flingy)
		f.setNextSpeed(&u.Flingy, originalSpeed)
		move()
		nextCollisionUnit = f.checkUnitMovementUnitCollision(u, ems)
		if nextCollisionUnit != collisionUnit {
			break
		}
	}
	u.NextVelocityDirection = originalVelocityDirection
	f.setCurrentVelocityDirection(&u.Flingy, u.NextVelocityDirection)
	f.setNextSpeed(&u.Flingy, originalSpeed)
	if forceMoveFree {
		u.MovementState = MovementStatesUMForceMoveFree
		return false
	}
	if nextCollisionUnit == nil {
		f.uSetMovementFlag(&u.Flingy, 1)
		if f.st.CurrentFrame-u.Path.CreationFrame < 150 {
			u.MovementState = MovementStatesUMFollowPath
		} else {
			u.MovementState = MovementStatesUMTurnAndStart
		}
	}
	return false
}

func minFp8(a Fp8, b Fp8) Fp8 {
	if b.Less(a) {
		return b
	}
	return a
}

func (f *StateFunctions) movementUMRepathMovers(u *Unit, ems *ExecuteMovementStruct) bool {
	f.freeUnitPath(u)
	moveTarget := u.MoveTarget
	if !f.unitPathTo(u, moveTarget.Pos, f.st.ConsiderCollisionWithUnitBug, true) || f.uImmovable(u) {
		u.MoveTarget = moveTarget
		u.MovementState = MovementStatesUMTurnAndStart
		return false
	}
	u.MovementState = MovementStatesUMFaceTarget
	return true
}

func (f *StateFunctions) movementUMGetFree(u *Unit, ems *ExecuteMovementStruct) bool {
	f.incrementPathingCollisionCounter(u)
	if f.unitUpdatePathMovementState(u, true) {
		return true
	}
	f.setNextSpeed(&u.Flingy, u.Path.LastCollisionSpeed)
	f.updateUnitMovementValues(&u.Flingy, ems)
	if f.checkUnitMovementUnitCollision(u, ems) != nil {
		if u.NextVelocityDirection == f.xyDirection(XySub(u.NextMovementWaypoint, u.Sprite.Position)) {
			f.freePath(u.Path)
			u.Path = nil
		}
	} else if f.checkUnitMovementTerrainCollisionEMS(u, ems) {
		u.MovementState = MovementStatesUMRepath
	} else {
		f.finishUnitMovement(u, ems)
		u.MovementState = MovementStatesUMFollowPath
	}
	return false
}

func (f *StateFunctions) movementUMLump(u *Unit, ems *ExecuteMovementStruct) bool {
	return false
}

func (f *StateFunctions) movementUMMoveToLegal(u *Unit, ems *ExecuteMovementStruct) bool {
	if u.Path == nil {
		u.MovementState = MovementStatesUMCheckIllegal
		return true
	}

	move := func() bool {
		if !f.isMovingAlongPath(u) {
			return true
		}
		f.updateUnitMovementValues(&u.Flingy, ems)
		if f.isUnitTypeInMapBounds(u.UnitType, ems.Position) {
			f.finishUnitMovement(u, ems)
			return false
		}
		if u.NextVelocityDirection == f.xyDirection(XySub(u.NextMovementWaypoint, u.Sprite.Position)) {
			return true
		}
		f.updateUnitMovementValues(&u.Flingy, ems)
		f.updateUnitHeading(&u.Flingy, u.CurrentVelocityDirection)
		return false
	}

	if u.PathingFlags&4 != 0 {
		u.PathingFlags &^= 4
		f.uSetStatusFlag(u, UnitStatusFlagCollision)
		f.freePathAndClear(u)
		u.MovementState = MovementStatesUMAtRest
		return true
	}
	if u.Path.StateFlags&1 != 0 || u.Path.Next != u.NextMovementWaypoint {
		u.NextMovementWaypoint = u.Path.Next
	}
	if move() {
		f.freePathAndClear(u)
		f.checkUnitCollision(u)
		f.uSetStatusFlag(u, UnitStatusFlagCollision)
		u.MovementState = MovementStatesUMCheckIllegal
	}
	return false
}

func (f *StateFunctions) movementUMHidden(u *Unit, ems *ExecuteMovementStruct) bool {
	ems.RefreshVision = false
	return false
}

func (f *StateFunctions) movementUMBunker(u *Unit, ems *ExecuteMovementStruct) bool {
	ems.RefreshVision = false
	f.updateCurrentVelocityDirectionTowardsWaypoint(&u.Flingy)
	f.setMovementFlags(&u.Flingy, ems)
	f.updateUnitHeading(&u.Flingy, u.CurrentVelocityDirection)
	return false
}

func (f *StateFunctions) movementUMLumpWannabe(u *Unit, ems *ExecuteMovementStruct) bool {
	if f.unitIsAtMoveTarget(&u.Flingy) && !f.uMovementFlag(&u.Flingy, 2) {
		u.MovementState = MovementStatesUMLump
		return true
	}
	f.updateUnitMovementValues(&u.Flingy, ems)
	if f.checkGroundUnitMovementUnitCollision(u, ems) != nil || f.checkUnitMovementTerrainCollisionEMS(u, ems) {
		f.stopBlockedMovement(u, ems)
	} else {
		f.finishUnitMovement(u, ems)
	}
	return false
}

func (f *StateFunctions) movementUMBldgTurret(u *Unit, ems *ExecuteMovementStruct) bool {
	f.updateCurrentVelocityDirectionTowardsWaypoint(&u.Flingy)
	f.setMovementFlags(&u.Flingy, ems)
	f.updateUnitHeading(&u.Flingy, u.CurrentVelocityDirection)
	return false
}

func (f *StateFunctions) runMovementState(u *Unit, ems *ExecuteMovementStruct) bool {
	switch u.MovementState {
	case MovementStatesUMInit:
		return f.movementUMInit(u, ems)
	case MovementStatesUMInitSeq:
		return f.movementUMInitSeq(u, ems)
	case MovementStatesUMLump:
		return f.movementUMLump(u, ems)
	case MovementStatesUMTurret:
		return f.movementUMTurret(u, ems)
	case MovementStatesUMBunker:
		return f.movementUMBunker(u, ems)
	case MovementStatesUMBldgTurret:
		return f.movementUMBldgTurret(u, ems)
	case MovementStatesUMHidden:
		return f.movementUMHidden(u, ems)
	case MovementStatesUMFlyer:
		return f.movementUMFlyer(u, ems)
	case MovementStatesUMAtRest:
		return f.movementUMAtRest(u, ems)
	case MovementStatesUMDormant:
		return f.movementUMDormant(u, ems)
	case MovementStatesUMAtMoveTarget:
		return f.movementUMAtMoveTarget(u, ems)
	case MovementStatesUMCheckIllegal:
		return f.movementUMCheckIllegal(u, ems)
	case MovementStatesUMMoveToLegal:
		return f.movementUMMoveToLegal(u, ems)
	case MovementStatesUMLumpWannabe:
		return f.movementUMLumpWannabe(u, ems)
	case MovementStatesUMFailedPath:
		return f.movementUMFailedPath(u, ems)
	case MovementStatesUMRetryPath:
		return f.movementUMRetryPath(u, ems)
	case MovementStatesUMStartPath:
		return f.movementUMStartPath(u, ems)
	case MovementStatesUMUIOrderDelay:
		return f.movementUMUIOrderDelay(u, ems)
	case MovementStatesUMTurnAndStart:
		return f.movementUMTurnAndStart(u, ems)
	case MovementStatesUMFaceTarget:
		return f.movementUMFaceTarget(u, ems)
	case MovementStatesUMNewMoveTarget:
		return f.movementUMNewMoveTarget(u, ems)
	case MovementStatesUMAnotherPath:
		return f.movementUMAnotherPath(u, ems)
	case MovementStatesUMRepath:
		return f.movementUMRepath(u, ems)
	case MovementStatesUMRepathMovers:
		return f.movementUMRepathMovers(u, ems)
	case MovementStatesUMFollowPath:
		return f.movementUMFollowPath(u, ems)
	case MovementStatesUMScoutPath:
		return f.movementUMScoutPath(u, ems)
	case MovementStatesUMFixCollision:
		return f.movementUMFixCollision(u, ems)
	case MovementStatesUMWaitFree:
		return f.movementUMWaitFree(u, ems)
	case MovementStatesUMGetFree:
		return f.movementUMGetFree(u, ems)
	case MovementStatesUMSlidePrep:
		return f.movementUMSlidePrep(u, ems)
	case MovementStatesUMSlideFree:
		return f.movementUMSlideFree(u, ems)
	case MovementStatesUMForceMoveFree:
		return f.movementUMForceMoveFree(u, ems)
	case MovementStatesUMFixTerrain:
		return f.movementUMFixTerrain(u, ems)
	case MovementStatesUMTerrainSlide:
		return f.movementUMTerrainSlide(u, ems)
	}
	throwError("fixme: movement state %d\n", u.MovementState)
	return false
}

func (f *StateFunctions) executeMovement(u *Unit) bool {
	var ems ExecuteMovementStruct
	ems.RefreshVision = f.updateTiles
	for f.runMovementState(u, &ems) {
	}
	return ems.RefreshVision
}

func movementModifier(stimTimer int32, upgraded bool, ensnareTimer int32) int32 {
	var mod int32
	if stimTimer != 0 {
		mod++
	}
	if upgraded {
		mod++
	}
	if ensnareTimer != 0 {
		mod--
	}
	return mod
}

func (f *StateFunctions) getModifiedUnitSpeed(u *Unit, baseSpeed Fp8) Fp8 {
	speed := baseSpeed.AsUnsigned()
	mod := movementModifier(u.StimTimer, f.uSpeedUpgrade(u), u.EnsnareTimer)
	if mod < 0 {
		speed = speed.DivInt(2)
	}
	if mod > 0 {
		if f.unitIsScout(u.UnitType) {
			speed = Ufp8Integer(6).Add(Ufp8Integer(1).Sub(Ufp8Integer(1).DivInt(3)))
		} else {
			speed = speed.Add(speed.DivInt(2))
			minSpeed := Ufp8Integer(3).Add(Ufp8Integer(1).DivInt(3))
			if speed.Less(minSpeed) {
				speed = minSpeed
			}
		}
	}
	return speed.AsSigned()
}

func (f *StateFunctions) getModifiedUnitAcceleration(u *Unit, baseAcceleration Fp8) Fp8 {
	acceleration := baseAcceleration.AsUnsigned()
	mod := movementModifier(u.StimTimer, f.uSpeedUpgrade(u), u.EnsnareTimer)
	if mod < 0 {
		acceleration = acceleration.Sub(acceleration.DivInt(4))
	}
	if mod > 0 {
		acceleration = acceleration.MulInt(2)
	}
	return acceleration.AsSigned()
}

func (f *StateFunctions) getModifiedUnitTurnRate(u *Unit, baseTurnRate Fp8) Fp8 {
	turnRate := baseTurnRate.AsUnsigned()
	mod := movementModifier(u.StimTimer, f.uSpeedUpgrade(u), u.EnsnareTimer)
	if mod < 0 {
		turnRate = turnRate.Sub(turnRate.DivInt(4))
	}
	if mod > 0 {
		turnRate = turnRate.MulInt(2)
	}
	return turnRate.AsSigned()
}

func (f *StateFunctions) getModifiedWeaponCooldown(u *Unit, weapon *WeaponTypeT) int32 {
	cooldown := weapon.Cooldown
	if u.AcidSporeCount != 0 {
		cooldown += max(cooldown/8, 3) * u.AcidSporeCount
	}
	mod := movementModifier(u.StimTimer, f.uCooldownUpgrade(u), u.EnsnareTimer)
	if mod > 0 {
		cooldown /= 2
	}
	if mod < 0 {
		cooldown += cooldown / 4
	}
	if cooldown > 250 {
		cooldown = 250
	}
	if cooldown < 5 {
		cooldown = 5
	}
	return cooldown
}

func (f *StateFunctions) unitHaltDistance(u *Flingy) Fp8 {
	speed := u.NextSpeed.AsUnsigned()
	if speed == Ufp8Zero() {
		return Fp8FromRaw(0)
	}
	if u.FlingyMovementType != 0 {
		return Fp8FromRaw(0)
	}
	if speed == u.FlingyType.TopSpeed.AsUnsigned() && u.FlingyAcceleration == u.FlingyType.Acceleration {
		return u.FlingyType.HaltDistance
	}
	return Ufp8MultiplyDivide(speed, speed, u.FlingyAcceleration.AsUnsigned()).DivInt(2).AsSigned()
}
