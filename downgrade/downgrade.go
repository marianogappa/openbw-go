package downgrade

import (
	"github.com/icza/screp/repparser/repdecoder"
)

type Report struct {
	SourceFormat string
	Dropped      []DroppedCommand
	Changes      []string
	SCRSettings  SCRSettings
}

func (r Report) DroppedByID() map[byte]int {
	counts := map[byte]int{}
	for _, d := range r.Dropped {
		counts[d.ID] += d.Count
	}
	return counts
}

func (r Report) DroppedCount() int {
	total := 0
	for _, d := range r.Dropped {
		total += d.Count
	}
	return total
}

func formatName(format repdecoder.RepFormat) string {
	switch format {
	case repdecoder.RepFormatLegacy:
		return "legacy"
	case repdecoder.RepFormatModern:
		return "1.18-1.20"
	case repdecoder.RepFormatModern121:
		return "1.21+"
	}
	return "unknown"
}

func Downgrade(scr []byte) (legacy []byte, report Report, err error) {
	src, err := readSourceSections(scr)
	if err != nil {
		return nil, report, err
	}
	legacy, report, _, err = downgradeSections(src)
	return legacy, report, err
}

type Prepared struct {
	Loadable []byte
	Settings SCRSettings
	Report   Report
	Commands []SourceCommand
}

func Prepare(replay []byte) (*Prepared, error) {
	src, err := readSourceSections(replay)
	if err != nil {
		return nil, err
	}
	if src.format == repdecoder.RepFormatLegacy {
		settings := src.settings()
		return &Prepared{
			Loadable: replay,
			Settings: settings,
			Report:   Report{SourceFormat: formatName(src.format), SCRSettings: settings},
			Commands: indexLegacyActions(src.actions),
		}, nil
	}
	legacy, report, commands, err := downgradeSections(src)
	if err != nil {
		return nil, err
	}
	return &Prepared{Loadable: legacy, Settings: report.SCRSettings, Report: report, Commands: commands}, nil
}

func PrepareForOpenBW(rep []byte) (loadable []byte, settings SCRSettings, err error) {
	prepared, err := Prepare(rep)
	if err != nil {
		return nil, settings, err
	}
	return prepared.Loadable, prepared.Settings, nil
}

func downgradeSections(src sourceSections) (legacy []byte, report Report, commands []SourceCommand, err error) {
	report.SourceFormat = formatName(src.format)
	report.SCRSettings = src.settings()
	if src.format == repdecoder.RepFormatLegacy {
		return encodeLegacyReplay(src.gameInfo, src.actions, src.mapData), report, indexLegacyActions(src.actions), nil
	}
	translator := newCommandTranslator(slotPlayerIDs(src.gameInfo))
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
