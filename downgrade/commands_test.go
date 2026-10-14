package downgrade

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func frameBlock(frame uint32, commands ...[]byte) []byte {
	block := bytes.Join(commands, nil)
	out := binary.LittleEndian.AppendUint32(nil, frame)
	out = append(out, byte(len(block)))
	return append(out, block...)
}

func testSlotIDs() [slotCount]uint32 {
	ids := [slotCount]uint32{}
	for i := range ids {
		ids[i] = 0xffffffff
	}
	ids[0], ids[1] = 0, 1
	return ids
}

func translateForTest(t *testing.T, actions []byte) ([]byte, []DroppedCommand) {
	t.Helper()
	translator := newCommandTranslator(testSlotIDs())
	out, err := translator.translateActions(actions)
	if err != nil {
		t.Fatal(err)
	}
	return out, translator.dropped
}

func TestTranslate121Commands(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want []byte
	}{
		{"right click", []byte{0, 0x60, 1, 0, 2, 0, 0xb4, 0x2c, 0, 0, 0xe4, 0, 1}, []byte{0, 0x14, 1, 0, 2, 0, 0x10, 0x0e, 0xe4, 0, 1}},
		{"targeted order", []byte{1, 0x61, 1, 0, 2, 0, 0x13, 0x0c, 1, 0, 0xe4, 0, 0x0e, 0}, []byte{1, 0x15, 1, 0, 2, 0, 0x6f, 0x45, 0xe4, 0, 0x0e, 0}},
		{"unload", []byte{0, 0x62, 0x01, 0x20, 0, 0}, []byte{0, 0x29, 0x01, 0x08}},
		{"select", []byte{0, 0x63, 2, 1, 0, 0, 0, 0, 0, 0, 0}, []byte{0, 0x09, 2, 1, 0, 0, 0}},
		{"select add", []byte{0, 0x64, 1, 0xa5, 0x06, 0, 0}, []byte{0, 0x0a, 1, 1, 0}},
		{"select remove", []byte{0, 0x65, 0}, []byte{0, 0x0b, 0}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, dropped := translateForTest(t, frameBlock(7, c.in))
			if want := frameBlock(7, c.want); !bytes.Equal(out, want) {
				t.Errorf("got % x, want % x", out, want)
			}
			if len(dropped) != 0 {
				t.Errorf("dropped %+v", dropped)
			}
		})
	}
}

func TestDropUnsupportedCommands(t *testing.T) {
	sync := []byte{0, 0x37, 1, 2, 3, 4, 5, 6}
	stop := []byte{1, 0x1a, 0}
	keepAlive := []byte{0, 0x05}
	out, dropped := translateForTest(t, frameBlock(3, sync, stop, keepAlive, keepAlive))
	if want := frameBlock(3, stop); !bytes.Equal(out, want) {
		t.Errorf("got % x, want % x", out, want)
	}
	want := []DroppedCommand{
		{ID: 0x37, Frame: 3, PlayerID: 0, Count: 1, Reason: ReasonUnsupported},
		{ID: 0x05, Frame: 3, PlayerID: 0, Count: 2, Reason: ReasonUnsupported},
	}
	assertDropped(t, dropped, want)
}

func TestDropUnknownCommandSkipsRestOfBlock(t *testing.T) {
	stop := []byte{1, 0x1a, 0}
	unknown := []byte{0, 0xd2, 9, 9, 9}
	out, dropped := translateForTest(t, append(frameBlock(4, stop, unknown, stop), frameBlock(5, stop)...))
	if want := append(frameBlock(4, stop), frameBlock(5, stop)...); !bytes.Equal(out, want) {
		t.Errorf("got % x, want % x", out, want)
	}
	assertDropped(t, dropped, []DroppedCommand{{ID: 0xd2, Frame: 4, PlayerID: 0, Count: 1, Reason: ReasonUnknownLength}})
}

func TestDropInvalidForOpenBW(t *testing.T) {
	observerChat := append([]byte{9, 0x5c}, make([]byte, 81)...)
	bigSelection := append([]byte{0, 0x63, 13}, make([]byte, 13*4)...)
	out, dropped := translateForTest(t, frameBlock(6, observerChat, bigSelection))
	if len(out) != 0 {
		t.Errorf("got % x, want an empty stream", out)
	}
	assertDropped(t, dropped, []DroppedCommand{
		{ID: 0x5c, Frame: 6, PlayerID: 9, Count: 1, Reason: ReasonUnknownPlayer},
		{ID: 0x63, Frame: 6, PlayerID: 0, Count: 1, Reason: ReasonTooManySelected},
	})
}

func TestTruncatedBlockIsAnError(t *testing.T) {
	broken := frameBlock(1, []byte{0, 0x1a, 0})[:6]
	if _, err := newCommandTranslator(testSlotIDs()).translateActions(broken); err == nil {
		t.Error("want an error")
	}
}

func assertDropped(t *testing.T, got, want []DroppedCommand) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("dropped %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("dropped[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}
