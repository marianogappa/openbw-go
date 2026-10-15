// Package pkware decompresses the PKWARE DCL segments of legacy StarCraft replays.
package pkware

import "fmt"

type formatError string

func (e formatError) Error() string { return string(e) }

func fail(format string, args ...any) {
	panic(formatError(fmt.Sprintf(format, args...)))
}

type bitReader struct {
	input []byte
	pos   int
	data  uint64
	bitsN uint
}

func (br *bitReader) take(n int) uint64 {
	if len(br.input)-br.pos < n {
		fail("attempt to read past end")
	}
	var v uint64
	for i := range n {
		v |= uint64(br.input[br.pos+i]) << (8 * i)
	}
	br.pos += n
	return v
}

func (br *bitReader) next() {
	left := len(br.input) - br.pos
	if left >= 4 {
		br.data |= br.take(4) << br.bitsN
		br.bitsN += 32
		return
	}
	switch left {
	case 0, 1:
		br.data |= br.take(1) << br.bitsN
		br.bitsN += 8
	case 2:
		br.data |= br.take(2) << br.bitsN
		br.bitsN += 16
	case 3:
		br.data |= br.take(2) << br.bitsN
		br.bitsN += 16
		br.data |= br.take(1) << br.bitsN
		br.bitsN += 8
	}
}

func (br *bitReader) getBits(bits uint) uint32 {
	if br.bitsN < bits {
		br.next()
	}
	r := uint32(br.data & ((uint64(1) << bits) - 1))
	br.data >>= bits
	br.bitsN -= bits
	return r
}

func (br *bitReader) getU8() uint8 {
	return uint8(br.getBits(8))
}

func (br *bitReader) getBit() bool {
	return br.getBits(1) != 0
}

func pickByBit(br *bitReader, ifSet, ifClear uint) uint {
	if br.getBit() {
		return ifSet
	}
	return ifClear
}

var decompressLength6BitsTable = [64]uint{
	3: 22, 7: 23, 11: 24, 15: 25, 19: 26, 23: 27, 27: 28, 31: 29,
	35: 30, 39: 31, 43: 32, 47: 33, 51: 34, 55: 35, 59: 36, 63: 37,
}

func decompressGetLength6Bits(br *bitReader) uint {
	v := uint(br.getBits(6))
	switch v % 4 {
	case 3:
		return decompressLength6BitsTable[v]
	case 0:
		step := v / 4
		if step%2 == 0 {
			return 262 + step/2 + 8*uint(br.getBits(5))
		}
		return 134 + step/2 + 8*uint(br.getBits(4))
	case 1:
		k := v / 4
		return pickByBit(br, 54+k, 38+k)
	default:
		return 70 + v/4 + 16*uint(br.getBits(2))
	}
}

func decompressGetLength(br *bitReader) uint {
	switch br.getBits(2) {
	case 3:
		return 1
	case 0:
		switch br.getBits(2) {
		case 3:
			return 6
		case 0:
			return decompressGetLength6Bits(br)
		case 1:
			if br.getBit() {
				return 7
			}
			return pickByBit(br, 9, 8)
		default:
			v := uint(br.getBits(3))
			if v%2 == 1 {
				return 10 + v/2
			}
			return pickByBit(br, 18+v/2, 14+v/2)
		}
	case 1:
		return pickByBit(br, 0, 2)
	default:
		if br.getBit() {
			return 3
		}
		return pickByBit(br, 4, 5)
	}
}

var decompressDistance5BitsTable = [32]uint{
	1: 39, 2: 47, 3: 31, 5: 35, 6: 43, 7: 27, 9: 37, 10: 45, 11: 29, 13: 33, 14: 41, 15: 25,
	17: 38, 18: 46, 19: 30, 21: 34, 22: 42, 23: 26, 25: 36, 26: 44, 27: 28, 29: 32, 30: 40, 31: 24,
}

var decompressDistance5BitsPairTable = [8]uint{62, 54, 58, 50, 60, 52, 56, 48}

var decompressDistance4BitsTable = [16]uint{
	1: 14, 2: 18, 3: 10, 4: 20, 5: 12, 6: 16, 7: 8, 8: 21, 9: 13, 10: 17, 11: 9, 12: 19, 13: 11, 14: 15, 15: 7,
}

func decompressGetDistance(br *bitReader) uint {
	switch br.getBits(2) {
	case 3:
		return 0
	case 0:
		v := br.getBits(5)
		if v%4 == 0 {
			base := decompressDistance5BitsPairTable[v/4]
			return pickByBit(br, base, base+1)
		}
		return decompressDistance5BitsTable[v]
	case 1:
		switch br.getBits(2) {
		case 1:
			return 2
		case 3:
			return 1
		case 0:
			return pickByBit(br, 5, 6)
		default:
			return pickByBit(br, 3, 4)
		}
	default:
		v := br.getBits(4)
		if v == 0 {
			return pickByBit(br, 22, 23)
		}
		return decompressDistance4BitsTable[v]
	}
}

// Decompress decompresses PKWARE DCL data (binary mode) into output, which has the decompressed size.
func Decompress(input, output []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(formatError)
			if !ok {
				panic(r)
			}
			err = e
		}
	}()
	decompress(input, output)
	return nil
}

func decompress(input, output []byte) {
	br := &bitReader{input: input}
	compressionType := br.getU8()
	distanceBits := uint(br.getU8())

	if distanceBits != 4 && distanceBits != 5 && distanceBits != 6 {
		fail("invalid distance bits %d", distanceBits)
	}
	if compressionType != 0 {
		fail("type %d not supported", compressionType)
	}

	outputSize := uint(len(output))
	outPos := uint(0)
	for outPos != outputSize {
		if !br.getBit() {
			output[outPos] = br.getU8()
			outPos++
			continue
		}
		length := 2 + decompressGetLength(br)
		if length == 519 {
			fail("eof marker found too early")
		}
		var distance uint
		if length == 2 {
			distance = decompressGetDistance(br)<<2 | uint(br.getBits(2))
		} else {
			distance = decompressGetDistance(br)<<distanceBits | uint(br.getBits(distanceBits))
		}
		srcPos := outPos - 1 - distance
		if srcPos > outputSize {
			fail("distance %d out of range at %d", distance, outPos)
		}
		if srcPos+length > outputSize {
			length = outputSize - srcPos
		}
		if outPos+length > outputSize {
			length = outputSize - outPos
		}
		for i := uint(0); i != length; i++ {
			output[outPos+i] = output[srcPos+i]
		}
		outPos += length
	}
}
