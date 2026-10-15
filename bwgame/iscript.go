package bwgame

func (f *StateFunctions) iscriptSetScript(image *Image, scriptID int32) {
	script, ok := f.globalSt.Iscript.Scripts[scriptID]
	if !ok {
		throwError("script %d does not exist", scriptID)
	}
	image.IscriptState.CurrentScript = script
}

func (f *StateFunctions) iscriptExecute(image *Image, state *IscriptState, noop bool, distanceMoved *Fp8, allowMainImageDestruction bool) bool {
	if state.Wait != 0 {
		state.Wait--
		return true
	}

	playFrame := func(frameIndex uint) {
		if image.FrameIndexBase == frameIndex {
			return
		}
		image.FrameIndexBase = frameIndex
		f.updateImageFrameIndex(image)
	}

	addImage := func(imageID ImageTypes, offset Xy, order int32) *Image {
		imageType := f.getImageType(imageID)
		scriptImage := image
		image := f.createImage(imageType, scriptImage.Sprite, offset, order, scriptImage)
		if image == nil {
			return nil
		}

		if image.Flags&ImageFlagHasDirectionalFrames != 0 {
			index := scriptImage.FrameIndexOffset
			if scriptImage.Flags&ImageFlagHorizontallyFlipped != 0 {
				index = 32 - scriptImage.FrameIndexOffset
			}
			f.setImageHeadingByIndex(image, index)
		}
		f.updateImageFrameIndex(image)
		if f.iscriptUnit != nil && (f.uBurrowed(f.iscriptUnit) || f.uInBunker(f.iscriptUnit)) {
			if !imageType.AlwaysVisible {
				f.hideImage(image)
			} else if image.Modifier == 0 {
				if scriptImage.Modifier >= 2 && scriptImage.Modifier <= 7 {
					f.setImageModifier(image, scriptImage.Modifier)
					image.ModifierData1 = scriptImage.ModifierData1
					image.ModifierData2 = scriptImage.ModifierData2
				}
			}
		}
		return image
	}

	useWeapon := func(u *Unit, weaponID WeaponTypes) {
		weaponType := f.getWeaponType(weaponID)
		if weaponType.Flingy.Get() == nil {
			return
		}
		f.createBullet(weaponType, u, u.OrderTarget.Pos, u.Owner, u.Heading)
		if weaponType.BulletCount == 2 {
			f.createBullet(weaponType, u, u.OrderTarget.Pos, u.Owner, u.Heading)
		}
	}

	attackWithWeapon := func(weapon *WeaponTypeT) {
		f.fireWeapon(f.iscriptUnit, weapon, -1)
		if weapon.BulletCount == 2 {
			f.fireWeapon(f.iscriptUnit, weapon, -1)
		}
	}

	attackWith := func(weaponN int32) {
		if f.iscriptUnit.OrderTarget.Unit == nil {
			return
		}
		var weapon *WeaponTypeT
		if weaponN == 1 {
			weapon = f.unitGroundWeapon(f.iscriptUnit)
		} else {
			weapon = f.unitAirWeapon(f.iscriptUnit)
		}
		if weapon == nil {
			throwError("attack_with %d: null weapon", weaponN)
		}
		attackWithWeapon(weapon)
	}

	attackWithForwardOffset := func(weaponN int32, forwardOffset int32) {
		if f.iscriptUnit.OrderTarget.Unit == nil {
			return
		}
		var weapon *WeaponTypeT
		if weaponN == 1 {
			weapon = f.unitGroundWeapon(f.iscriptUnit)
		} else {
			weapon = f.unitAirWeapon(f.iscriptUnit)
		}
		if weapon == nil {
			throwError("attack_with %d: null weapon", weaponN)
		}
		f.fireWeapon(f.iscriptUnit, weapon, forwardOffset)
		if weapon.BulletCount == 2 {
			f.fireWeapon(f.iscriptUnit, weapon, forwardOffset)
		}
	}

	programData := f.globalSt.Iscript.ProgramData
	p := state.ProgramCounter

	for {
		pc := p
		if pc == 0 {
			throwError("iscript: program counter is null")
		}
		opc := programData[p] - iscriptOpcodeBias
		p++
		var a, b, c int32
		switch opc {
		case IscriptOpcodesOpcPlayfram:
			a, p = programData[p], p+1
			if noop {
				break
			}
			playFrame(uint(a))
		case IscriptOpcodesOpcPlayframtile:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if uint(a)+f.gameSt.TilesetIndex < uint(len(image.Grp.Frames)) {
				playFrame(uint(a) + f.gameSt.TilesetIndex)
			}
		case IscriptOpcodesOpcSethorpos:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if image.Offset.X != a {
				image.Offset.X = a
				image.Flags |= ImageFlagRedraw
			}
		case IscriptOpcodesOpcSetvertpos:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit == nil || (!f.uRequiresDetector(f.iscriptUnit) && !f.uCloaked(f.iscriptUnit)) {
				if image.Offset.Y != a {
					image.Offset.Y = a
					image.Flags |= ImageFlagRedraw
				}
			}
		case IscriptOpcodesOpcSetpos:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			if noop {
				break
			}
			f.setImageOffset(image, Xy{X: a, Y: b})
		case IscriptOpcodesOpcWait:
			state.Wait = programData[p] - 1
			p++
			state.ProgramCounter = p
			return true
		case IscriptOpcodesOpcWaitrand:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			if noop {
				break
			}
			state.Wait = a + ((f.lcgRand(3) & 0xff) % (b - a + 1)) - 1
			state.ProgramCounter = p
			return true
		case IscriptOpcodesOpcGoto:
			p = uint(programData[p])
		case IscriptOpcodesOpcImgol, IscriptOpcodesOpcImgul:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			c, p = programData[p], p+1
			if noop {
				break
			}
			order := imageOrderBelow
			if opc == IscriptOpcodesOpcImgol {
				order = imageOrderAbove
			}
			addImage(ImageTypes(a), XyAdd(image.Offset, Xy{X: b, Y: c}), order)
		case IscriptOpcodesOpcImgolorig, IscriptOpcodesOpcSwitchul:
			a, p = programData[p], p+1
			if noop {
				break
			}
			order := imageOrderBelow
			if opc == IscriptOpcodesOpcImgolorig {
				order = imageOrderAbove
			}
			if newImage := addImage(ImageTypes(a), Xy{}, order); newImage != nil {
				if !f.iFlag(newImage, ImageFlagUsesSpecialOffset) {
					f.iSetFlag(newImage, ImageFlagUsesSpecialOffset)
					f.updateImageSpecialOffset(newImage)
				}
			}
		case IscriptOpcodesOpcImgoluselo:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			c, p = programData[p], p+1
			if noop {
				break
			}
			addImage(ImageTypes(a), f.getImageLoOffset(image, uint(b), uint(c), false), imageOrderAbove)

		case IscriptOpcodesOpcSprol:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			c, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptBullet != nil && f.iscriptBullet.BulletOwnerUnit != nil && f.unitIsGoliath(f.iscriptBullet.BulletOwnerUnit.UnitType) && f.playerHasUpgrade(f.iscriptBullet.BulletOwnerUnit.Owner, UpgradeTypesCharonBoosters) {
				f.createThingyAtImage(image, f.getSpriteType(SpriteTypesSPRITEIDHaloRocketsTrail), Xy{X: b, Y: c}, image.Sprite.ElevationLevel+1)
			} else {
				f.createThingyAtImage(image, f.getSpriteType(SpriteTypes(a)), Xy{X: b, Y: c}, image.Sprite.ElevationLevel+1)
			}

		case IscriptOpcodesOpcLowsprul:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			c, p = programData[p], p+1
			if noop {
				break
			}
			f.createThingyAtImage(image, f.getSpriteType(SpriteTypes(a)), Xy{X: b, Y: c}, 1)

		case IscriptOpcodesOpcSpruluselo:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			c, p = programData[p], p+1
			if noop {
				break
			}
			if sprite := f.getSpriteType(SpriteTypes(a)); sprite != nil {
				if f.iscriptUnit != nil && (f.uRequiresDetector(f.iscriptUnit) || f.uCloaked(f.iscriptUnit)) && !sprite.Image.Get().AlwaysVisible {
					break
				}
				t := f.createThingyAtImage(image, sprite, Xy{X: b, Y: c}, image.Sprite.ElevationLevel)
				if t != nil {
					f.setSpriteImagesHeadingByImageIndex(t.Sprite, image)
				}
			}
		case IscriptOpcodesOpcSprul:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			c, p = programData[p], p+1
			if noop {
				break
			}
			if sprite := f.getSpriteType(SpriteTypes(a)); sprite != nil {
				if f.iscriptUnit != nil && (f.uRequiresDetector(f.iscriptUnit) || f.uCloaked(f.iscriptUnit)) && !sprite.Image.Get().AlwaysVisible {
					break
				}
				t := f.createThingyAtImage(image, sprite, Xy{X: b, Y: c}, image.Sprite.ElevationLevel-1)
				if t != nil {
					f.setSpriteImagesHeadingByImageIndex(t.Sprite, image)
				}
			}
		case IscriptOpcodesOpcSproluselo:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			if noop {
				break
			}
			if sprite := f.getSpriteType(SpriteTypes(a)); sprite != nil {
				t := f.createThingyAtImage(image, sprite, f.getImageLoOffset(image, uint(b), 0, false), image.Sprite.ElevationLevel+1)
				if t != nil {
					f.setSpriteImagesHeadingByImageIndex(t.Sprite, image)
				}
			}
		case IscriptOpcodesOpcEnd:
			if noop {
				break
			}
			if image == image.Sprite.MainImage && !allowMainImageDestruction {
				throwError("iscript_execute: main image not allowed to be destroyed here")
			}
			state.ProgramCounter = 0
			f.destroyImage(image)
			return false
		case IscriptOpcodesOpcSetflipstate:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iFlag(image, ImageFlagHorizontallyFlipped) != (a != 0) {
				f.iSetFlagValue(image, ImageFlagHorizontallyFlipped, a != 0)
				f.setImageModifier(image, image.Modifier)
				if image.Flags&ImageFlagUsesSpecialOffset != 0 {
					f.updateImageSpecialOffset(image)
				}
			}
		case IscriptOpcodesOpcPlaysnd:
			a, p = programData[p], p+1
			if noop {
				break
			}
			f.playSoundAt(a, image.Sprite.Position, nil, false)
		case IscriptOpcodesOpcPlaysndrand:
			p = f.iscriptPlaySoundRandom(image, programData, p, noop)
		case IscriptOpcodesOpcPlaysndbtwn:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			if noop {
				break
			}
			f.playSoundAt(a+f.lcgRand(5)%(b-a+1), image.Sprite.Position, nil, false)
		case IscriptOpcodesOpcDomissiledmg, IscriptOpcodesOpcDogrddamage:
			if noop {
				break
			}
			if f.iscriptBullet != nil {
				f.bulletHit(f.iscriptBullet)
			}
		case IscriptOpcodesOpcAttackmelee:
			if !noop && f.iscriptUnit != nil {
				f.meleeDealDamage(f.iscriptUnit)
			}
			p = f.iscriptPlaySoundRandom(image, programData, p, noop)

		case IscriptOpcodesOpcFollowmaingraphic:
			if noop {
				break
			}
			if mainImage := image.Sprite.MainImage; mainImage != nil {
				frameIndex := mainImage.FrameIndex
				flipped := f.iFlag(mainImage, ImageFlagHorizontallyFlipped)
				if image.FrameIndex != frameIndex || f.iFlag(image, ImageFlagHorizontallyFlipped) != flipped {
					image.FrameIndexBase = mainImage.FrameIndexBase
					f.setImageFrameIndexOffset(image, mainImage.FrameIndexOffset, flipped)
				}
			}
		case IscriptOpcodesOpcRandcondjmp:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			if (f.lcgRand(7) & 0xff) <= a {
				p = uint(b)
			}

		case IscriptOpcodesOpcTurnccwise:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit != nil {
				f.setUnitHeading(f.iscriptUnit, f.iscriptUnit.Heading.Sub(eighthTurn().MulInt(a)))
			}
		case IscriptOpcodesOpcTurncwise:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit != nil {
				f.setUnitHeading(f.iscriptUnit, f.iscriptUnit.Heading.Add(eighthTurn().MulInt(a)))
			}
		case IscriptOpcodesOpcTurn1cwise:
			if noop {
				break
			}
			if f.iscriptUnit != nil && f.iscriptUnit.OrderTarget.Unit == nil {
				f.setUnitHeading(f.iscriptUnit, f.iscriptUnit.Heading.Add(eighthTurn()))
			}
		case IscriptOpcodesOpcTurnrand:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.lcgRand(6)%4 == 1 {
				if f.iscriptUnit != nil {
					f.setUnitHeading(f.iscriptUnit, f.iscriptUnit.Heading.Sub(eighthTurn().MulInt(a)))
				}
			} else {
				if f.iscriptUnit != nil {
					f.setUnitHeading(f.iscriptUnit, f.iscriptUnit.Heading.Add(eighthTurn().MulInt(a)))
				}
			}

		case IscriptOpcodesOpcSigorder:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptFlingy != nil {
				f.iscriptFlingy.OrderSignal |= a
			}
		case IscriptOpcodesOpcAttackwith:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit != nil {
				attackWith(a)
			}
		case IscriptOpcodesOpcAttack:
			if noop {
				break
			}
			if f.iscriptUnit != nil && f.iscriptUnit.OrderTarget.Unit != nil {
				if !f.uFlying(f.iscriptUnit.OrderTarget.Unit) {
					attackWith(1)
				} else {
					attackWith(2)
				}
			}
		case IscriptOpcodesOpcCastspell:
			if noop {
				break
			}
			if f.iscriptUnit != nil && f.iscriptUnit.OrderType.Weapon != WeaponTypesNone {
				if f.spellOrderValid(f.iscriptUnit) {
					attackWithWeapon(f.getWeaponType(f.iscriptUnit.OrderType.Weapon))
				}
			}
		case IscriptOpcodesOpcUseweapon:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit != nil {
				useWeapon(f.iscriptUnit, WeaponTypes(a))
			}
		case IscriptOpcodesOpcMove:
			a, p = programData[p], p+1
			if distanceMoved != nil {
				if f.iscriptUnit != nil {
					*distanceMoved = f.getModifiedUnitSpeed(f.iscriptUnit, Fp8Integer(a))
				}
			}
			if noop {
				break
			}
			if f.iscriptUnit != nil {
				f.setNextSpeed(&f.iscriptUnit.Flingy, f.getModifiedUnitSpeed(f.iscriptUnit, Fp8Integer(a)))
			}
		case IscriptOpcodesOpcGotorepeatattk:
			if noop {
				break
			}
			if f.iscriptUnit != nil {
				f.uUnsetMovementFlag(&f.iscriptUnit.Flingy, 8)
			}
		case IscriptOpcodesOpcEngframe:
			a, p = programData[p], p+1
			if noop {
				break
			}
			image.FrameIndexBase = uint(a)
			f.setImageFrameIndexOffset(image, image.Sprite.MainImage.FrameIndexOffset, f.iFlag(image.Sprite.MainImage, ImageFlagHorizontallyFlipped))
		case IscriptOpcodesOpcEngset:
			a, p = programData[p], p+1
			if noop {
				break
			}
			mainImage := image.Sprite.MainImage
			image.FrameIndexBase = mainImage.FrameIndexBase + (uint(len(mainImage.Grp.Frames))&0x7fff)*uint(a)
			f.setImageFrameIndexOffset(image, mainImage.FrameIndexOffset, f.iFlag(mainImage, ImageFlagHorizontallyFlipped))

		case IscriptOpcodesOpcNobrkcodestart:
			if noop {
				break
			}
			if f.iscriptUnit != nil {
				f.uSetStatusFlag(f.iscriptUnit, UnitStatusFlagIscriptNobrk)
				f.iscriptUnit.Sprite.Flags |= SpriteFlagIscriptNobrk
			}
		case IscriptOpcodesOpcNobrkcodeend:
			if noop {
				break
			}
			if f.iscriptUnit != nil {
				f.uUnsetStatusFlag(f.iscriptUnit, UnitStatusFlagIscriptNobrk)
				f.iscriptUnit.Sprite.Flags &^= SpriteFlagIscriptNobrk
				if !f.iscriptUnit.OrderQueue.Empty() && f.iscriptUnit.UserActionFlags&1 != 0 {
					f.iscriptRunToIdle(f.iscriptUnit)
					f.activateNextOrder(f.iscriptUnit)
				}
			}
		case IscriptOpcodesOpcIgnorerest:
			if noop {
				break
			}
			if f.iscriptUnit != nil && f.iscriptUnit.OrderTarget.Unit == nil {
				f.iscriptRunToIdle(f.iscriptUnit)
				break
			}
			state.Wait = 10
			state.ProgramCounter = p - 1
			return true
		case IscriptOpcodesOpcAttkshiftproj:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit != nil {
				attackWithForwardOffset(1, a)
			}
		case IscriptOpcodesOpcTmprmgraphicstart:
			if noop {
				break
			}
			f.hideImage(image)
		case IscriptOpcodesOpcTmprmgraphicend:
			if noop {
				break
			}
			f.showImage(image)

		case IscriptOpcodesOpcSetfldirect:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit != nil {
				f.setUnitHeading(f.iscriptUnit, eighthTurn().MulInt(a))
			}

		case IscriptOpcodesOpcSetflspeed:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit != nil {
				f.iscriptUnit.FlingyTopSpeed = Fp8FromRaw(a)
			}

		case IscriptOpcodesOpcCall:
			a, p = programData[p], p+1
			state.ReturnAddress = p
			p = uint(a)
		case IscriptOpcodesOpcReturn:
			p = state.ReturnAddress

		case IscriptOpcodesOpcCreategasoverlays:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit != nil && f.utResource(f.iscriptUnit.UnitType) {
				imageID := ImageTypesIMAGEIDVespeneGeyserSmoke1Overlay
				if f.iscriptUnit.Building.Resource.ResourceCount != 0 {
					imageID = ImageTypesIMAGEIDVespeneGeyserSmoke1
				}
				imageID = ImageTypes(uint(imageID) + uint(a))
				f.createImage(f.getImageType(imageID), image.Sprite, XyAdd(image.Offset, f.getImageLoOffset(image, 2, uint(a), false)), imageOrderAbove, image)
			}
		case IscriptOpcodesOpcPwrupcondjmp:
			a, p = programData[p], p+1
			if image.Sprite != nil && image.Sprite.MainImage != image {
				p = uint(a)
			}
		case IscriptOpcodesOpcTrgtrangecondjmp:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			if noop {
				continue
			}
			if f.iscriptUnit != nil && f.iscriptUnit.OrderTarget.Unit != nil {
				pos := f.getBulletAppearAtTargetPos(f.iscriptUnit, f.iscriptUnit.OrderTarget.Unit)
				if f.xyLengthFP8(XyFp8Sub(f.toXyFp8(pos), f.iscriptUnit.ExactPosition)).IntegerPart() <= a {
					p = uint(b)
				}
			}
		case IscriptOpcodesOpcTrgtarccondjmp:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			c, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit != nil && f.iscriptUnit.OrderTarget.Unit != nil {
				targetDirection := f.xyDirection(XySub(f.iscriptUnit.OrderTarget.Unit.Sprite.Position, f.iscriptUnit.Sprite.Position))
				if Fp8Extend(DirectionTFromRaw(int8(a)).Sub(targetDirection)).Abs().Less(Fp8FromRaw(b)) {
					p = uint(c)
				}
			}
		case IscriptOpcodesOpcCurdirectcondjmp:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			c, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit != nil && Fp8Extend(f.iscriptUnit.Heading.Sub(DirectionTFromRaw(int8(a)))).Abs().Less(Fp8FromRaw(b)) {
				p = uint(c)
			}
		case IscriptOpcodesOpcImgulnextid:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			if noop {
				break
			}
			addImage(ImageTypes(int32(image.ImageType.ID)+1), XyAdd(image.Offset, Xy{X: a, Y: b}), imageOrderBelow)

		case IscriptOpcodesOpcLiftoffcondjmp:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptUnit != nil && f.uFlying(f.iscriptUnit) {
				p = uint(a)
			}
		case IscriptOpcodesOpcWarpoverlay:
			a, p = programData[p], p+1
			if noop {
				break
			}
			image.ModifierData1 = a & 0xff
			image.ModifierData2 = (a >> 8) & 0xff
		case IscriptOpcodesOpcOrderdone:
			a, p = programData[p], p+1
			if noop {
				break
			}
			if f.iscriptFlingy != nil {
				f.iscriptFlingy.OrderSignal &^= a
			}
		case IscriptOpcodesOpcGrdsprol:
			a, p = programData[p], p+1
			b, p = programData[p], p+1
			c, p = programData[p], p+1
			if noop {
				break
			}
			fitPos := XyAdd(XyAdd(image.Sprite.Position, image.Offset), Xy{X: b, Y: c})
			if f.unitTypeCanFitAt(f.getUnitType(UnitTypesTerranMarine), fitPos) {
				f.createThingyAtImage(image, f.getSpriteType(SpriteTypes(a)), Xy{X: b, Y: c}, image.Sprite.ElevationLevel+1)
			}
		default:
			throwError("iscript: unhandled opcode %d", opc)
		}
	}
}

func (f *StateFunctions) iscriptPlaySoundRandom(image *Image, programData []int32, p uint, noop bool) uint {
	n := programData[p]
	p++
	if !noop {
		index := f.lcgRand(4) % n
		f.playSoundAt(programData[p+uint(index)], image.Sprite.Position, nil, false)
	}
	return p + uint(n)
}

func eighthTurn() DirectionT {
	return DirectionTFromRaw(8)
}

func (f *StateFunctions) iscriptRunAnim(image *Image, newAnim int32) bool {
	oldAnim := image.IscriptState.Animation
	if newAnim == IscriptAnimsDeath && oldAnim == IscriptAnimsDeath {
		return true
	}
	if ^image.Flags&ImageFlagHasIscriptAnimations != 0 && newAnim != IscriptAnimsInit && newAnim != IscriptAnimsDeath {
		return true
	}
	if (newAnim == IscriptAnimsWalking || newAnim == IscriptAnimsIsWorking) && newAnim == oldAnim {
		return true
	}
	if newAnim == IscriptAnimsGndAttkRpt && oldAnim != IscriptAnimsGndAttkRpt {
		if oldAnim != IscriptAnimsGndAttkInit {
			newAnim = IscriptAnimsGndAttkInit
		}
	}
	if newAnim == IscriptAnimsAirAttkRpt && oldAnim != IscriptAnimsAirAttkRpt {
		if oldAnim != IscriptAnimsAirAttkInit {
			newAnim = IscriptAnimsAirAttkInit
		}
	}
	script := image.IscriptState.CurrentScript
	if script == nil {
		throwError("attempt to start animation without a script")
	}
	animsPc := script.AnimationPc
	if uint(newAnim) >= uint(len(animsPc)) {
		throwError("script %d does not have animation %d", script.ID, newAnim)
	}
	image.IscriptState.Animation = newAnim
	image.IscriptState.ProgramCounter = animsPc[newAnim]
	image.IscriptState.ReturnAddress = 0
	image.IscriptState.Wait = 0
	return f.iscriptExecute(image, &image.IscriptState, false, nil, false)
}

func (f *StateFunctions) imageUpdateCloak(image *Image) {
	if image.ModifierData2 != 0 {
		image.ModifierData2--
		return
	}
	image.ModifierData1++
	image.ModifierData2 = 3
	if image.ModifierData1 >= 8 {
		f.setImageModifier(image, image.Modifier+1)
		if f.iscriptUnit != nil {
			f.uSetStatusFlag(f.iscriptUnit, UnitStatusFlagCloaked)
			f.uSetStatusFlag(f.iscriptUnit, UnitStatusFlagRequiresDetector)
		}
	}
}

func (f *StateFunctions) imageUpdateDecloak(image *Image) {
	if image.ModifierData2 != 0 {
		image.ModifierData2--
		return
	}
	image.ModifierData2 = 3
	if image.ModifierData1 != 0 {
		image.ModifierData1--
		return
	}
	f.setImageModifier(image, 0)
	if f.iscriptUnit != nil {
		f.uUnsetStatusFlag(f.iscriptUnit, UnitStatusFlagCloaked)
		f.uUnsetStatusFlag(f.iscriptUnit, UnitStatusFlagRequiresDetector)
	}
}

func (f *StateFunctions) imageUpdateWarpin(image *Image) {
	if image.ModifierData2 != 0 {
		image.ModifierData2--
		return
	}
	image.ModifierData2 = 2
	if image.ModifierData1 < 63 {
		image.ModifierData1++
		return
	}
	f.iscriptRunAnim(image, IscriptAnimsDeath)
	if f.iscriptFlingy != nil {
		f.iscriptFlingy.OrderSignal |= 1
	}
}

func (f *StateFunctions) iscriptExecuteSprite(sprite *Sprite) bool {
	for i := sprite.Images.Begin(); i != sprite.Images.End(); {
		image := i
		i = sprite.Images.Next(i)
		if image.Modifier == 2 || image.Modifier == 5 {
			f.imageUpdateCloak(image)
		} else if image.Modifier == 4 || image.Modifier == 7 {
			f.imageUpdateDecloak(image)
		} else if image.Modifier == 17 {
			f.imageUpdateWarpin(image)
		}
		isMainImage := image == sprite.MainImage
		destroyed := !f.iscriptExecute(image, &image.IscriptState, false, nil, true)
		if isMainImage && destroyed && !sprite.Images.Empty() {
			sprite.MainImage = sprite.Images.Front()
		}
	}
	if !sprite.Images.Empty() {
		return true
	}

	f.removeSpriteFromTileLine(sprite)
	f.st.SpritesContainer.Push(sprite)

	return false
}

func (f *StateFunctions) spriteRunAnim(sprite *Sprite, anim int32) {
	for i := sprite.Images.Begin(); i != sprite.Images.End(); {
		image := i
		i = sprite.Images.Next(i)
		f.iscriptRunAnim(image, anim)
	}
}
