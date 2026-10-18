package downgrade

// Report tells what the downgrade did with a replay.
type Report struct {
	// SourceFormat is the input format; see SourceFormat.
	SourceFormat string
	Dropped      []DroppedCommand
	// Changes describe the edits to the game info and the map, one per line.
	Changes     []string
	SCRSettings SCRSettings
}

// DroppedByID counts the dropped commands by command ID.
func (r Report) DroppedByID() map[byte]int {
	counts := map[byte]int{}
	for _, d := range r.Dropped {
		counts[d.ID] += d.Count
	}
	return counts
}

// DroppedCount is the number of dropped commands.
func (r Report) DroppedCount() int {
	total := 0
	for _, d := range r.Dropped {
		total += d.Count
	}
	return total
}

// Downgrade converts a replay (SC:R or legacy) into the bytes of a legacy replay.
func Downgrade(scr []byte) (legacy []byte, report Report, err error) {
	src, err := readSourceSections(scr)
	if err != nil {
		return nil, report, err
	}
	legacy, report, _, err = downgradeSections(src)
	return legacy, report, err
}

// Prepared is a replay ready for the engine.
type Prepared struct {
	// Loadable is a legacy replay: the input itself when it is legacy.
	Loadable []byte
	Settings SCRSettings
	Report   Report
	// Commands are the commands of the input replay, in stream order.
	Commands []SourceCommand
}

// Prepare downgrades a replay for the engine and indexes its commands.
func Prepare(replay []byte) (*Prepared, error) {
	src, err := readSourceSections(replay)
	if err != nil {
		return nil, err
	}
	if src.format == formatLegacy {
		settings := src.settings()
		return &Prepared{
			Loadable: replay,
			Settings: settings,
			Report:   Report{SourceFormat: src.format.String(), SCRSettings: settings},
			Commands: indexLegacyActions(src.actions),
		}, nil
	}
	legacy, report, commands, err := downgradeSections(src)
	if err != nil {
		return nil, err
	}
	return &Prepared{Loadable: legacy, Settings: report.SCRSettings, Report: report, Commands: commands}, nil
}

// PrepareForOpenBW is Prepare that returns only the loadable replay and the settings.
func PrepareForOpenBW(rep []byte) (loadable []byte, settings SCRSettings, err error) {
	prepared, err := Prepare(rep)
	if err != nil {
		return nil, settings, err
	}
	return prepared.Loadable, prepared.Settings, nil
}

func downgradeSections(src sourceSections) (legacy []byte, report Report, commands []SourceCommand, err error) {
	report.SourceFormat = src.format.String()
	report.SCRSettings = src.settings()
	if src.format == formatLegacy {
		return encodeLegacyReplay(src.gameInfo, src.actions, src.mapData), report, indexLegacyActions(src.actions), nil
	}
	translator := newCommandTranslator(slotPlayerIDs(src.gameInfo), report.SCRSettings.Limits.Units)
	actions, err := translator.translateActions(src.actions)
	if err != nil {
		return nil, report, nil, err
	}
	report.Dropped = translator.dropped
	report.Changes = downgradePlayerColors(src.gameInfo, src.playerColors)
	mapData, mapChanges := downgradeMap(src.mapData)
	report.Changes = append(report.Changes, mapChanges...)
	return encodeLegacyReplay(src.gameInfo, actions, mapData), report, translator.commands, nil
}
