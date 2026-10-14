package bwgame

type Fp8 struct {
	RawValue int32
}

func Fp8FromRaw(rawValue int32) Fp8 {
	return Fp8{RawValue: rawValue}
}

type Fp1 struct {
	RawValue int32
}

func Fp1FromRaw(rawValue int32) Fp1 {
	return Fp1{RawValue: rawValue}
}

type DirectionT struct {
	RawValue int8
}

func DirectionTFromRaw(rawValue int8) DirectionT {
	return DirectionT{RawValue: rawValue}
}

type XyT[T any] struct {
	X T
	Y T
}

type Xy = XyT[int32]

type RectT[T any] struct {
	From T
	To   T
}

type Rect = RectT[Xy]

type TypeContainer[T any] struct {
	Vec []T
}

type TypeID[T any, I ~int] struct {
	ID  I
	Ptr *T
}

func (t TypeID[T, I]) Get() *T {
	return t.Ptr
}

func (t TypeID[T, I]) Valid() bool {
	return t.Ptr != nil
}

type UnitTypeFlagsT uint32

const (
	UnitTypeTFlagBuilding              UnitTypeFlagsT = 1
	UnitTypeTFlagAddon                 UnitTypeFlagsT = 2
	UnitTypeTFlagFlyer                 UnitTypeFlagsT = 4
	UnitTypeTFlagWorker                UnitTypeFlagsT = 8
	UnitTypeTFlagTurret                UnitTypeFlagsT = 0x10
	UnitTypeTFlagFlyingBuilding        UnitTypeFlagsT = 0x20
	UnitTypeTFlagHero                  UnitTypeFlagsT = 0x40
	UnitTypeTFlagRegensHp              UnitTypeFlagsT = 0x80
	UnitTypeTFlag100                   UnitTypeFlagsT = 0x100
	UnitTypeTFlagTwoUnitsInOneEgg      UnitTypeFlagsT = 0x400
	UnitTypeTFlagPowerup               UnitTypeFlagsT = 0x800
	UnitTypeTFlagResourceDepot         UnitTypeFlagsT = 0x1000
	UnitTypeTFlagResource              UnitTypeFlagsT = 0x2000
	UnitTypeTFlagRobotic               UnitTypeFlagsT = 0x4000
	UnitTypeTFlagDetector              UnitTypeFlagsT = 0x8000
	UnitTypeTFlagOrganic               UnitTypeFlagsT = 0x10000
	UnitTypeTFlagRequiresCreep         UnitTypeFlagsT = 0x20000
	UnitTypeTFlagRequiresPsionicMatrix UnitTypeFlagsT = 0x80000
	UnitTypeTFlagCanBurrow             UnitTypeFlagsT = 0x100000
	UnitTypeTFlagHasEnergy             UnitTypeFlagsT = 0x200000
	UnitTypeTFlagInitiallyCloaked      UnitTypeFlagsT = 0x400000
	UnitTypeTFlagSpriteSizeMedium      UnitTypeFlagsT = 0x2000000
	UnitTypeTFlagSpriteSizeLarge       UnitTypeFlagsT = 0x4000000
	UnitTypeTFlagCanMove               UnitTypeFlagsT = 0x8000000
	UnitTypeTFlagCanTurn               UnitTypeFlagsT = 0x10000000
	UnitTypeTFlagInvincible            UnitTypeFlagsT = 0x20000000
	UnitTypeTFlagMechanical            UnitTypeFlagsT = 0x40000000
	UnitTypeTFlag80000000              UnitTypeFlagsT = 0x80000000
)

type UnitTypeT struct {
	ID UnitTypes

	Flingy                    TypeID[FlingyTypeT, FlingyTypes]
	TurretUnitType            TypeID[UnitTypeT, UnitTypes]
	Subunit2                  TypeID[UnitTypeT, UnitTypes]
	InfestationUnit           TypeID[UnitTypeT, UnitTypes]
	ConstructionAnimation     TypeID[ImageTypeT, ImageTypes]
	UnitDirection             int32
	HasShield                 bool
	ShieldPoints              int32
	Hitpoints                 Fp8
	ElevationLevel            int32
	Unknown1                  int32
	Sublabel                  int32
	ComputerAiIdle            TypeID[OrderTypeT, Orders]
	HumanAiIdle               TypeID[OrderTypeT, Orders]
	ReturnToIdle              TypeID[OrderTypeT, Orders]
	AttackUnit                TypeID[OrderTypeT, Orders]
	AttackMove                TypeID[OrderTypeT, Orders]
	GroundWeapon              TypeID[WeaponTypeT, WeaponTypes]
	MaxGroundHits             int32
	AirWeapon                 TypeID[WeaponTypeT, WeaponTypes]
	MaxAirHits                int32
	AiInternal                int32
	Flags                     UnitTypeFlagsT
	TargetAcquisitionRange    int32
	SightRange                int32
	ArmorUpgrade              TypeID[UpgradeTypeT, UpgradeTypes]
	UnitSize                  int32
	Armor                     int32
	RightClickAction          uint
	ReadySound                int32
	FirstWhatSound            int32
	LastWhatSound             int32
	FirstPissedSound          int32
	LastPissedSound           int32
	FirstYesSound             int32
	LastYesSound              int32
	PlacementSize             Xy
	AddonPosition             Xy
	Dimensions                Rect
	Portrait                  int32
	MineralCost               int32
	GasCost                   int32
	BuildTime                 int32
	Unknown2                  int32
	GroupFlags                int32
	SupplyProvided            Fp1
	SupplyRequired            Fp1
	SpaceRequired             uint
	SpaceProvided             uint
	BuildScore                int32
	DestroyScore              int32
	UnitMapStringIndex        uint
	IsBroodwar                bool
	StareditAvailabilityFlags int32
}

type UnitTypesT = TypeContainer[UnitTypeT]

const (
	WeaponTypeTBulletTypeFly = iota
	WeaponTypeTBulletTypeFollowTarget
	WeaponTypeTBulletTypeAppearAtTargetUnit
	WeaponTypeTBulletTypePersistAtTargetPos
	WeaponTypeTBulletTypeAppearAtTargetPos
	WeaponTypeTBulletTypeAppearAtSourceUnit
	WeaponTypeTBulletTypeSelfDestruct
	WeaponTypeTBulletTypeBounce
	WeaponTypeTBulletTypeAttackTargetPos
	WeaponTypeTBulletTypeExtendToMaxRange
)

const (
	WeaponTypeTHitTypeNone = iota
	WeaponTypeTHitTypeNormal
	WeaponTypeTHitTypeRadialSplash
	WeaponTypeTHitTypeEnemySplash
	WeaponTypeTHitTypeLockdown
	WeaponTypeTHitTypeNuclearMissile
	WeaponTypeTHitTypeParasite
	WeaponTypeTHitTypeBroodlings
	WeaponTypeTHitTypeEmpShockwave
	WeaponTypeTHitTypeIrradiate
	WeaponTypeTHitTypeEnsnare
	WeaponTypeTHitTypePlague
	WeaponTypeTHitTypeStasisField
	WeaponTypeTHitTypeDarkSwarm
	WeaponTypeTHitTypeConsume
	WeaponTypeTHitTypeYamatoGun
	WeaponTypeTHitTypeRestoration
	WeaponTypeTHitTypeDisruptionWeb
	WeaponTypeTHitTypeCorrosiveAcid
	WeaponTypeTHitTypeMindControl
	WeaponTypeTHitTypeFeedback
	WeaponTypeTHitTypeOpticalFlare
	WeaponTypeTHitTypeMaelstrom
	WeaponTypeTHitType23
	WeaponTypeTHitTypeAirSplash
)

const (
	WeaponTypeTDamageTypeNone = iota
	WeaponTypeTDamageTypeExplosive
	WeaponTypeTDamageTypeConcussive
	WeaponTypeTDamageTypeNormal
	WeaponTypeTDamageTypeIgnoreArmor
)

type WeaponTypeT struct {
	ID WeaponTypes

	Label               int32
	Flingy              TypeID[FlingyTypeT, FlingyTypes]
	Unused              int32
	TargetFlags         int32
	MinRange            int32
	MaxRange            int32
	DamageUpgrade       TypeID[UpgradeTypeT, UpgradeTypes]
	DamageType          int32
	BulletType          int32
	Lifetime            int32
	HitType             int32
	InnerSplashRadius   int32
	MediumSplashRadius  int32
	OuterSplashRadius   int32
	DamageAmount        int32
	DamageBonus         int32
	Cooldown            int32
	BulletCount         int32
	AttackAngle         Fp8
	BulletHeadingOffset DirectionT
	ForwardOffset       int32
	UpwardOffset        int32
	TargetErrorMessage  int32
	Icon                int32
}

type WeaponTypesT = TypeContainer[WeaponTypeT]

type UpgradeTypeT struct {
	ID UpgradeTypes

	MineralCostBase   int32
	MineralCostFactor int32
	GasCostBase       int32
	GasCostFactor     int32
	TimeCostBase      int32
	TimeCostFactor    int32
	Unknown           int32
	Icon              int32
	Label             int32
	Race              int32
	MaxLevel          int32
	IsBroodwar        bool
}

type UpgradeTypesT = TypeContainer[UpgradeTypeT]

type TechTypeT struct {
	ID TechTypes

	MineralCost  int32
	GasCost      int32
	ResearchTime int32
	EnergyCost   int32
	Unknown      int32
	Icon         int32
	Label        int32
	Race         int32
	Flags        int32
}

type TechTypesT = TypeContainer[TechTypeT]

type FlingyTypeT struct {
	ID FlingyTypes

	Sprite       TypeID[SpriteTypeT, SpriteTypes]
	TopSpeed     Fp8
	Acceleration Fp8
	HaltDistance Fp8
	TurnRate     Fp8
	Unused       int32
	MovementType int32
}

type FlingyTypesT = TypeContainer[FlingyTypeT]

type SpriteTypeT struct {
	ID SpriteTypes

	Image               TypeID[ImageTypeT, ImageTypes]
	HealthBarSize       int32
	Unk0                int32
	Visible             bool
	SelectionCircle     int32
	SelectionCircleVpos int32
}

type SpriteTypesT = TypeContainer[SpriteTypeT]

type ImageTypeT struct {
	ID ImageTypes

	GrpFilenameIndex         int32
	HasDirectionalFrames     bool
	IsClickable              bool
	HasIscriptAnimations     bool
	AlwaysVisible            bool
	Modifier                 int32
	ColorShift               int32
	IscriptID                int32
	ShieldFilenameIndex      int32
	AttackFilenameIndex      int32
	DamageFilenameIndex      int32
	SpecialFilenameIndex     int32
	LandingDustFilenameIndex int32
	LiftOffFilenameIndex     int32
}

type ImageTypesT = TypeContainer[ImageTypeT]

type OrderTypeT struct {
	ID Orders

	Label            int32
	TargetsEnemies   bool
	Background       int32
	Unused3          int32
	ValidForTurret   bool
	Unused5          int32
	CanBeInterrupted bool
	Unk7             bool
	CanBeQueued      bool
	Unk9             int32
	CanBeObstructed  bool
	Unk11            int32
	Unused12         int32
	Weapon           WeaponTypes
	TechType         TechTypes
	Animation        int32
	Highlight        int32
	DepIndex         int32
	TargetOrder      Orders
}

type OrderTypesT = TypeContainer[OrderTypeT]

const (
	IscriptOpcodesOpcPlayfram = iota
	IscriptOpcodesOpcPlayframtile
	IscriptOpcodesOpcSethorpos
	IscriptOpcodesOpcSetvertpos
	IscriptOpcodesOpcSetpos
	IscriptOpcodesOpcWait
	IscriptOpcodesOpcWaitrand
	IscriptOpcodesOpcGoto
	IscriptOpcodesOpcImgol
	IscriptOpcodesOpcImgul
	IscriptOpcodesOpcImgolorig
	IscriptOpcodesOpcSwitchul
	IscriptOpcodesOpc0c
	IscriptOpcodesOpcImgoluselo
	IscriptOpcodesOpcImguluselo
	IscriptOpcodesOpcSprol
	IscriptOpcodesOpcHighsprol
	IscriptOpcodesOpcLowsprul
	IscriptOpcodesOpcUflunstable
	IscriptOpcodesOpcSpruluselo
	IscriptOpcodesOpcSprul
	IscriptOpcodesOpcSproluselo
	IscriptOpcodesOpcEnd
	IscriptOpcodesOpcSetflipstate
	IscriptOpcodesOpcPlaysnd
	IscriptOpcodesOpcPlaysndrand
	IscriptOpcodesOpcPlaysndbtwn
	IscriptOpcodesOpcDomissiledmg
	IscriptOpcodesOpcAttackmelee
	IscriptOpcodesOpcFollowmaingraphic
	IscriptOpcodesOpcRandcondjmp
	IscriptOpcodesOpcTurnccwise
	IscriptOpcodesOpcTurncwise
	IscriptOpcodesOpcTurn1cwise
	IscriptOpcodesOpcTurnrand
	IscriptOpcodesOpcSetspawnframe
	IscriptOpcodesOpcSigorder
	IscriptOpcodesOpcAttackwith
	IscriptOpcodesOpcAttack
	IscriptOpcodesOpcCastspell
	IscriptOpcodesOpcUseweapon
	IscriptOpcodesOpcMove
	IscriptOpcodesOpcGotorepeatattk
	IscriptOpcodesOpcEngframe
	IscriptOpcodesOpcEngset
	IscriptOpcodesOpc2d
	IscriptOpcodesOpcNobrkcodestart
	IscriptOpcodesOpcNobrkcodeend
	IscriptOpcodesOpcIgnorerest
	IscriptOpcodesOpcAttkshiftproj
	IscriptOpcodesOpcTmprmgraphicstart
	IscriptOpcodesOpcTmprmgraphicend
	IscriptOpcodesOpcSetfldirect
	IscriptOpcodesOpcCall
	IscriptOpcodesOpcReturn
	IscriptOpcodesOpcSetflspeed
	IscriptOpcodesOpcCreategasoverlays
	IscriptOpcodesOpcPwrupcondjmp
	IscriptOpcodesOpcTrgtrangecondjmp
	IscriptOpcodesOpcTrgtarccondjmp
	IscriptOpcodesOpcCurdirectcondjmp
	IscriptOpcodesOpcImgulnextid
	IscriptOpcodesOpc3e
	IscriptOpcodesOpcLiftoffcondjmp
	IscriptOpcodesOpcWarpoverlay
	IscriptOpcodesOpcOrderdone
	IscriptOpcodesOpcGrdsprol
	IscriptOpcodesOpc43
	IscriptOpcodesOpcDogrddamage
)

const (
	IscriptAnimsInit = iota
	IscriptAnimsDeath
	IscriptAnimsGndAttkInit
	IscriptAnimsAirAttkInit
	IscriptAnimsUnused1
	IscriptAnimsGndAttkRpt
	IscriptAnimsAirAttkRpt
	IscriptAnimsCastSpell
	IscriptAnimsGndAttkToIdle
	IscriptAnimsAirAttkToIdle
	IscriptAnimsUnused2
	IscriptAnimsWalking
	IscriptAnimsWalkingToIdle
	IscriptAnimsSpecialState1
	IscriptAnimsSpecialState2
	IscriptAnimsAlmostBuilt
	IscriptAnimsBuilt
	IscriptAnimsLanding
	IscriptAnimsLiftOff
	IscriptAnimsIsWorking
	IscriptAnimsWorkingToIdle
	IscriptAnimsWarpIn
	IscriptAnimsUnused3
	IscriptAnimsStarEditInit
	IscriptAnimsDisable
	IscriptAnimsBurrow
	IscriptAnimsUnBurrow
	IscriptAnimsEnable
)

type IscriptTScript struct {
	ID          int32
	AnimationPc []uint
}

type IscriptT struct {
	Scripts     map[int32]*IscriptTScript
	ProgramData []int32
}

type GrpTFrameT struct {
	Offset         XyT[uint]
	Size           XyT[uint]
	LineDataOffset []uint
	DataContainer  []uint8
}

type GrpT struct {
	Width  uint
	Height uint
	Frames []GrpTFrameT
}
