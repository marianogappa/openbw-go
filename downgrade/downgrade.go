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
	return downgradeSections(src)
}

func PrepareForOpenBW(rep []byte) (loadable []byte, settings SCRSettings, err error) {
	src, err := readSourceSections(rep)
	if err != nil {
		return nil, settings, err
	}
	if src.format == repdecoder.RepFormatLegacy {
		return rep, src.settings(), nil
	}
	legacy, report, err := downgradeSections(src)
	return legacy, report.SCRSettings, err
}

func downgradeSections(src sourceSections) (legacy []byte, report Report, err error) {
	report.SourceFormat = formatName(src.format)
	report.SCRSettings = src.settings()
	if src.format == repdecoder.RepFormatLegacy {
		return encodeLegacyReplay(src.gameInfo, src.actions, src.mapData), report, nil
	}
	translator := newCommandTranslator(slotPlayerIDs(src.gameInfo))
	actions, err := translator.translateActions(src.actions)
	if err != nil {
		return nil, report, err
	}
	report.Dropped = translator.dropped
	report.Changes = downgradePlayerColors(src.gameInfo, src.playerColors)
	mapData, mapChanges := downgradeMap(src.mapData)
	report.Changes = append(report.Changes, mapChanges...)
	return encodeLegacyReplay(src.gameInfo, actions, mapData), report, nil
}
