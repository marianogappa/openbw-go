package bwgame

func (f *StateFunctions) getSpriteTileLineIndex(y int32) uint {
	panic(notPorted("get_sprite_tile_line_index"))
}

func (f *StateFunctions) addSpriteToTileLine(sprite *Sprite) {
	panic(notPorted("add_sprite_to_tile_line"))
}

func (f *StateFunctions) removeSpriteFromTileLine(sprite *Sprite) {
	panic(notPorted("remove_sprite_from_tile_line"))
}

func (f *StateFunctions) moveSprite(sprite *Sprite, newPosition Xy) {
	panic(notPorted("move_sprite"))
}

func (f *StateFunctions) setSpriteVisibility(sprite *Sprite, visibilityFlags int32) {
	panic(notPorted("set_sprite_visibility"))
}

func (f *StateFunctions) setImageOffset(image *Image, offset Xy) {
	panic(notPorted("set_image_offset"))
}

func (f *StateFunctions) setImageModifier(image *Image, modifier int32) {
	panic(notPorted("set_image_modifier"))
}

func (f *StateFunctions) showImage(image *Image) {
	panic(notPorted("show_image"))
}

func (f *StateFunctions) hideImage(image *Image) {
	panic(notPorted("hide_image"))
}

func (f *StateFunctions) updateImageSpecialOffset(image *Image) {
	panic(notPorted("update_image_special_offset"))
}

func (f *StateFunctions) updateImageFrameIndex(image *Image) {
	panic(notPorted("update_image_frame_index"))
}

func (f *StateFunctions) setImageFrameIndexOffset(image *Image, frameIndexOffset uint, flipped bool) {
	panic(notPorted("set_image_frame_index_offset"))
}

func (f *StateFunctions) setImageHeading(image *Image, heading DirectionT) {
	panic(notPorted("set_image_heading"))
}

func (f *StateFunctions) setImageHeadingByIndex(image *Image, frameIndexOffset uint) {
	panic(notPorted("set_image_heading_by_index"))
}

func (f *StateFunctions) setSpriteImagesHeadingByIndex(sprite *Sprite, frameIndexOffset uint) {
	panic(notPorted("set_sprite_images_heading_by_index"))
}

func (f *StateFunctions) setSpriteImagesHeadingByImageIndex(sprite *Sprite, image *Image) {
	panic(notPorted("set_sprite_images_heading_by_image_index"))
}

func (f *StateFunctions) getImageCenterMapPosition(image *Image) Xy {
	panic(notPorted("get_image_center_map_position"))
}

func (f *StateFunctions) getImageMapPosition(image *Image) Xy {
	panic(notPorted("get_image_map_position"))
}

func (f *StateFunctions) getImageLoOffset(image *Image, loIndex uint, offsetIndex uint, useFrameIndexOffset bool) Xy {
	panic(notPorted("get_image_lo_offset"))
}

func (f *StateFunctions) iscriptSetScript(image *Image, scriptID int32) {
	panic(notPorted("iscript_set_script"))
}

func (f *StateFunctions) newThingy() *Thingy {
	panic(notPorted("new_thingy"))
}

func (f *StateFunctions) initializeThingy(t *Thingy, spriteType *SpriteTypeT, pos Xy, owner int32) bool {
	panic(notPorted("initialize_thingy"))
}

func (f *StateFunctions) createThingy(spriteType *SpriteTypeT, pos Xy, owner int32) *Thingy {
	panic(notPorted("create_thingy"))
}

func (f *StateFunctions) createThingyAtImage(parentImage *Image, spriteType *SpriteTypeT, offset Xy, elevationLevel int32) *Thingy {
	panic(notPorted("create_thingy_at_image"))
}

func (f *StateFunctions) iscriptExecute(image *Image, state *IscriptState, noop bool, distanceMoved *Fp8, allowMainImageDestruction bool) bool {
	panic(notPorted("iscript_execute"))
}

func (f *StateFunctions) iscriptRunAnim(image *Image, newAnim int32) bool {
	panic(notPorted("iscript_run_anim"))
}

func (f *StateFunctions) imageUpdateCloak(image *Image) {
	panic(notPorted("image_update_cloak"))
}

func (f *StateFunctions) imageUpdateDecloak(image *Image) {
	panic(notPorted("image_update_decloak"))
}

func (f *StateFunctions) imageUpdateWarpin(image *Image) {
	panic(notPorted("image_update_warpin"))
}

func (f *StateFunctions) iscriptExecuteSprite(sprite *Sprite) bool {
	panic(notPorted("iscript_execute_sprite"))
}

func (f *StateFunctions) spriteRunAnim(sprite *Sprite, anim int32) {
	panic(notPorted("sprite_run_anim"))
}

func (f *StateFunctions) initializeImage(image *Image, imageType *ImageTypeT, sprite *Sprite, offset Xy) {
	panic(notPorted("initialize_image"))
}

func (f *StateFunctions) destroyImage(image *Image) {
	panic(notPorted("destroy_image"))
}

func (f *StateFunctions) createImage(imageType *ImageTypeT, sprite *Sprite, offset Xy, order int32, relimg *Image) *Image {
	panic(notPorted("create_image"))
}

func (f *StateFunctions) destroySprite(sprite *Sprite) {
	panic(notPorted("destroy_sprite"))
}

func (f *StateFunctions) createSprite(spriteType *SpriteTypeT, pos Xy, owner int32) *Sprite {
	panic(notPorted("create_sprite"))
}

func (f *StateFunctions) initializeFlingy(flingy *Flingy, flingyType *FlingyTypeT, pos Xy, owner int32, heading DirectionT) bool {
	panic(notPorted("initialize_flingy"))
}
