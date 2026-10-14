package bwgame

func (f *StateFunctions) getSpriteTileLineIndex(y int32) uint {
	if y < 0 {
		return 0
	}
	r := uint(y) / 32
	if r >= f.gameSt.MapTileHeight {
		return f.gameSt.MapTileHeight - 1
	}
	return r
}

func (f *StateFunctions) addSpriteToTileLine(sprite *Sprite) {
	index := f.getSpriteTileLineIndex(sprite.Position.Y)
	bwInsertList(&f.st.SpritesOnTileLine[index], sprite)
}

func (f *StateFunctions) removeSpriteFromTileLine(sprite *Sprite) {
	index := f.getSpriteTileLineIndex(sprite.Position.Y)
	f.st.SpritesOnTileLine[index].Remove(sprite)
}

func (f *StateFunctions) moveSprite(sprite *Sprite, newPosition Xy) {
	if sprite.Position == newPosition {
		return
	}
	oldIndex := f.getSpriteTileLineIndex(sprite.Position.Y)
	newIndex := f.getSpriteTileLineIndex(newPosition.Y)
	sprite.Position = newPosition
	if oldIndex != newIndex {
		f.st.SpritesOnTileLine[oldIndex].Remove(sprite)
		bwInsertList(&f.st.SpritesOnTileLine[newIndex], sprite)
	}
}

func (f *StateFunctions) setSpriteVisibility(sprite *Sprite, visibilityFlags int32) {
	if sprite.VisibilityFlags == visibilityFlags {
		return
	}
	sprite.VisibilityFlags = visibilityFlags
}

func (f *StateFunctions) setImageOffset(image *Image, offset Xy) {
	if image.Offset == offset {
		return
	}
	image.Offset = offset
	image.Flags |= ImageFlagRedraw
}

func (f *StateFunctions) setImageModifier(image *Image, modifier int32) {
	image.Modifier = modifier
	if modifier == 17 {
		image.ModifierData1 = 48
		image.ModifierData2 = 2
	}
	image.Flags |= ImageFlagRedraw
}

func (f *StateFunctions) showImage(image *Image) {
	if ^image.Flags&ImageFlagHidden != 0 {
		return
	}
	image.Flags &^= ImageFlagHidden
	image.Flags |= ImageFlagRedraw
}

func (f *StateFunctions) hideImage(image *Image) {
	if image.Flags&ImageFlagHidden != 0 {
		return
	}
	image.Flags |= ImageFlagHidden
}

func (f *StateFunctions) updateImageSpecialOffset(image *Image) {
	f.setImageOffset(image, f.getImageLoOffset(image.Sprite.MainImage, 2, 0, false))
}

func (f *StateFunctions) updateImageFrameIndex(image *Image) {
	frameIndex := image.FrameIndexBase + image.FrameIndexOffset
	if image.FrameIndex != frameIndex {
		image.FrameIndex = frameIndex
		image.Flags |= ImageFlagRedraw
	}
}

func (f *StateFunctions) setImageFrameIndexOffset(image *Image, frameIndexOffset uint, flipped bool) {
	image.FrameIndexOffset = frameIndexOffset
	f.iSetFlagValue(image, ImageFlagHorizontallyFlipped, flipped)
	f.setImageModifier(image, image.Modifier)
	f.updateImageFrameIndex(image)
}

func (f *StateFunctions) setImageHeading(image *Image, heading DirectionT) {
	if image.Flags&ImageFlagUsesSpecialOffset != 0 {
		f.updateImageSpecialOffset(image)
	}
	if image.Flags&ImageFlagHasDirectionalFrames != 0 {
		frameIndexOffset := (f.directionIndex(heading) + 4) / 8
		flipped := false
		if frameIndexOffset > 16 {
			frameIndexOffset = 32 - frameIndexOffset
			flipped = true
		}
		if image.FrameIndexOffset != frameIndexOffset || f.iFlag(image, ImageFlagHorizontallyFlipped) != flipped {
			f.setImageFrameIndexOffset(image, frameIndexOffset, flipped)
		}
	}
}

func (f *StateFunctions) setImageHeadingByIndex(image *Image, frameIndexOffset uint) {
	if image.Flags&ImageFlagHasDirectionalFrames != 0 {
		flipped := false
		if frameIndexOffset > 16 {
			frameIndexOffset = 32 - frameIndexOffset
			flipped = true
		}
		if image.FrameIndexOffset != frameIndexOffset || f.iFlag(image, ImageFlagHorizontallyFlipped) != flipped {
			f.setImageFrameIndexOffset(image, frameIndexOffset, flipped)
			if image.Flags&ImageFlagUsesSpecialOffset != 0 {
				f.updateImageSpecialOffset(image)
			}
		}
	}
}

func (f *StateFunctions) setSpriteImagesHeadingByIndex(sprite *Sprite, frameIndexOffset uint) {
	for i := range sprite.Images.All() {
		f.setImageHeadingByIndex(i, frameIndexOffset)
	}
}

func (f *StateFunctions) setSpriteImagesHeadingByImageIndex(sprite *Sprite, image *Image) {
	index := image.FrameIndexOffset
	if f.iFlag(image, ImageFlagHorizontallyFlipped) {
		index = 32 - index
	}
	for i := range sprite.Images.All() {
		f.setImageHeadingByIndex(i, index)
	}
}

func (f *StateFunctions) getImageCenterMapPosition(image *Image) Xy {
	return XyAdd(image.Sprite.Position, image.Offset)
}

func (f *StateFunctions) getImageMapPosition(image *Image) Xy {
	mapPos := XyAdd(image.Sprite.Position, image.Offset)
	frame := &image.Grp.Frames[image.FrameIndex]
	if image.Flags&ImageFlagHorizontallyFlipped != 0 {
		mapPos.X += int32(image.Grp.Width/2 - (frame.Offset.X + frame.Size.X))
	} else {
		mapPos.X += int32(frame.Offset.X - image.Grp.Width/2)
	}
	if image.Flags&ImageFlagYFrozen != 0 {
		mapPos.Y = image.FrozenYValue
	} else {
		mapPos.Y += int32(frame.Offset.Y - image.Grp.Height/2)
	}
	return mapPos
}

func (f *StateFunctions) getImageLoOffset(image *Image, loIndex uint, offsetIndex uint, useFrameIndexOffset bool) Xy {
	frame := image.FrameIndex
	if useFrameIndexOffset {
		frame = image.FrameIndexOffset
	}
	imageID := uint(image.ImageType.ID)
	if imageID >= uint(len(f.globalSt.ImageLoOffsets)) {
		throwError("image_lo_offsets: invalid image id %d", imageID)
	}
	loOffsets := &f.globalSt.ImageLoOffsets[imageID]
	if loIndex >= uint(len(loOffsets)) {
		throwError("invalid lo index %d\n", loIndex)
	}
	frameOffsets := *loOffsets[loIndex]
	if frame >= uint(len(frameOffsets)) {
		throwError("image %d lo_index %d does not have offsets for frame %d (frame_offsets.size() is %d)", int32(image.ImageType.ID), loIndex, frame, len(frameOffsets))
	}
	if offsetIndex >= uint(len(frameOffsets[frame])) {
		throwError("image %d lo_index %d frame %d does not contain an offset at index %d", int32(image.ImageType.ID), loIndex, frame, offsetIndex)
	}
	r := frameOffsets[frame][offsetIndex]
	if image.Flags&ImageFlagHorizontallyFlipped != 0 {
		r.X = -r.X
	}
	return r
}

func (f *StateFunctions) newThingy() *Thingy {
	if !f.st.FreeThingies.Empty() {
		r := f.st.FreeThingies.Front()
		f.st.FreeThingies.PopFront()
		return r
	}
	if len(f.st.Thingies) >= 500 {
		return nil
	}
	t := &Thingy{}
	f.st.Thingies = append(f.st.Thingies, t)
	return t
}

func (f *StateFunctions) initializeThingy(t *Thingy, spriteType *SpriteTypeT, pos Xy, owner int32) bool {
	t.Hp = Fp8One()
	t.Sprite = f.createSprite(spriteType, pos, owner)
	return t.Sprite != nil
}

func (f *StateFunctions) createThingy(spriteType *SpriteTypeT, pos Xy, owner int32) *Thingy {
	t := f.newThingy()
	if t == nil {
		return nil
	}
	if !f.initializeThingy(t, spriteType, pos, owner) {
		f.st.FreeThingies.PushFront(t)
		return nil
	}
	f.st.ActiveThingiesSize++
	bwInsertList(&f.st.ActiveThingies, t)
	return t
}

func (f *StateFunctions) createThingyAtImage(parentImage *Image, spriteType *SpriteTypeT, offset Xy, elevationLevel int32) *Thingy {
	pos := XyAdd(XyAdd(parentImage.Sprite.Position, parentImage.Offset), offset)
	t := f.createThingy(spriteType, pos, 0)
	if t == nil {
		return nil
	}
	t.Sprite.ElevationLevel = elevationLevel
	if !f.usHidden(t) {
		f.setSpriteVisibility(t.Sprite, int32(f.tileVisibility(t.Sprite.Position)))
	}
	return t
}

func (f *StateFunctions) initializeImage(image *Image, imageType *ImageTypeT, sprite *Sprite, offset Xy) {
	image.ImageType = imageType
	image.Grp = f.globalSt.ImageGrp[uint(imageType.ID)]
	var flags int32
	if imageType.HasDirectionalFrames {
		flags |= ImageFlagHasDirectionalFrames
	}
	if imageType.IsClickable {
		flags |= ImageFlagClickable
	}
	image.Flags = flags
	image.FrameIndexBase = 0
	image.FrameIndexOffset = 0
	image.FrameIndex = 0
	image.Sprite = sprite
	image.Offset = offset
	image.ModifierData1 = 0
	image.ModifierData2 = 0
	image.IscriptState.CurrentScript = nil
	image.IscriptState.ProgramCounter = 0
	image.IscriptState.ReturnAddress = 0
	image.IscriptState.Animation = 0
	image.IscriptState.Wait = 0
	if imageType.Modifier == 14 {
		image.ModifierData1 = sprite.Owner
		image.ModifierData2 = 0
	}
}

func (f *StateFunctions) destroyImage(image *Image) {
	image.Grp = nil
	image.Sprite.Images.Remove(image)
	f.st.ImagesContainer.Push(image)
}

func (f *StateFunctions) createImage(imageType *ImageTypeT, sprite *Sprite, offset Xy, order int32, relimg *Image) *Image {
	if imageType == nil {
		throwError("attempt to create image of null type")
	}

	image := f.st.ImagesContainer.Top()
	if image == nil {
		return nil
	}
	f.st.ImagesContainer.Pop()

	images := &sprite.Images
	if images.Empty() {
		sprite.MainImage = image
		images.PushFront(image)
	} else if order == imageOrderTop {
		images.PushFront(image)
	} else if order == imageOrderBottom {
		images.PushBack(image)
	} else {
		if relimg == nil {
			relimg = sprite.MainImage
		}
		if order == imageOrderAbove {
			images.Insert(relimg, image)
		} else {
			images.Insert(images.Next(relimg), image)
		}
	}

	f.initializeImage(image, imageType, sprite, offset)
	f.setImageModifier(image, image.ImageType.Modifier)
	f.iSetFlagValue(image, ImageFlagHasIscriptAnimations, image.ImageType.HasIscriptAnimations)
	f.iscriptSetScript(image, image.ImageType.IscriptID)
	if !f.iscriptRunAnim(image, IscriptAnimsInit) {
		throwError("create_image: image destroyed")
	}
	return image
}

func (f *StateFunctions) destroySprite(sprite *Sprite) {
	for i := sprite.Images.Begin(); i != sprite.Images.End(); {
		image := i
		i = sprite.Images.Next(i)
		f.destroyImage(image)
	}
	f.removeSpriteFromTileLine(sprite)
	f.st.SpritesContainer.Push(sprite)
}

func (f *StateFunctions) createSprite(spriteType *SpriteTypeT, pos Xy, owner int32) *Sprite {
	if spriteType == nil {
		throwError("attempt to create sprite of null type")
	}

	sprite := f.st.SpritesContainer.Top()
	if sprite == nil {
		return nil
	}
	f.st.SpritesContainer.Pop()

	initializeSprite := func() bool {
		if uint(pos.X) >= f.gameSt.MapWidth || uint(pos.Y) >= f.gameSt.MapHeight {
			return false
		}
		sprite.Owner = owner
		sprite.SpriteType = spriteType
		sprite.Flags = 0
		sprite.Position = pos
		sprite.VisibilityFlags = ^int32(0)
		sprite.ElevationLevel = 4
		sprite.SelectionTimer = 0
		sprite.Images.Clear()
		if !spriteType.Visible {
			sprite.Flags |= SpriteFlagHidden
			f.setSpriteVisibility(sprite, 0)
		}
		if f.createImage(spriteType.Image.Get(), sprite, Xy{}, imageOrderAbove, nil) == nil {
			return false
		}
		sprite.Width = min(sprite.MainImage.Grp.Width, 0xff)
		sprite.Height = min(sprite.MainImage.Grp.Height, 0xff)
		return true
	}

	if !initializeSprite() {
		f.st.SpritesContainer.Push(sprite)
		return nil
	}
	f.addSpriteToTileLine(sprite)

	return sprite
}

func (f *StateFunctions) initializeFlingy(flingy *Flingy, flingyType *FlingyTypeT, pos Xy, owner int32, heading DirectionT) bool {
	flingy.FlingyType = flingyType
	flingy.MovementFlags = 0
	flingy.NextSpeed = Fp8Zero()
	flingy.FlingyTopSpeed = flingyType.TopSpeed
	flingy.FlingyAcceleration = flingyType.Acceleration
	flingy.FlingyTurnRate = flingyType.TurnRate
	flingy.FlingyMovementType = flingyType.MovementType

	flingy.Position = pos
	flingy.ExactPosition = f.toXyFp8(pos)

	f.setFlingyMoveTarget(flingy, pos, nil)
	f.setNextTargetWaypoint(flingy, pos)
	flingy.Heading = heading
	flingy.NextVelocityDirection = heading
	flingy.Hp = Fp8One()

	flingy.Sprite = f.createSprite(flingyType.Sprite.Get(), pos, owner)
	if flingy.Sprite == nil {
		return false
	}
	dir := flingy.Heading
	for i := range flingy.Sprite.Images.All() {
		f.setImageHeading(i, dir)
	}

	return true
}
