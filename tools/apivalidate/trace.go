package main

import (
	"fmt"
	"os"

	openbw "github.com/marianogappa/openbw-go"
)

func traceUnit(data *openbw.Data, path string, unit openbw.UnitID) error {
	replay, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	g, err := openbw.Open(replay, openbw.Options{Data: data, Kinds: openbw.AllKinds &^ openbw.KindTick})
	if err != nil {
		return err
	}
	for e := range g.Events() {
		if mentions(e, unit) {
			fmt.Printf("%T %+v\n", e, e)
		}
	}
	return g.Err()
}

func mentions(e openbw.Event, unit openbw.UnitID) bool {
	switch e := e.(type) {
	case *openbw.UnitCreated:
		return e.Unit == unit || e.Parent == unit
	case *openbw.UnitCompleted:
		return e.Unit == unit
	case *openbw.UnitMorphed:
		return e.Unit == unit
	case *openbw.MorphStarted:
		return e.Unit == unit || e.Morpher == unit
	case *openbw.UnitDestroyed:
		return e.Unit == unit || e.Killer == unit || e.Container == unit
	case *openbw.UnitOwnerChanged:
		return e.Unit == unit
	case *openbw.ProductionCancelled:
		return e.Unit == unit || e.Producer == unit
	}
	return false
}
