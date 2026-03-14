//go:build darwin

// Package smc reads Apple SMC temperature sensors natively via IOKit (no exec).
// Adapted from https://github.com/dkorunic/iSMC (GPLv2+, Copyright 2019 Dinko Korunic).
package smc

import (
	"encoding/binary"
	"math"
	"strings"

	"github.com/cherepovskiy/air-temp-scraper/internal/smc/gosmc"
)

const appleSMC = "AppleSMC"

// Reading is a single temperature sensor reading.
type Reading struct {
	Desc  string
	Key   string
	Value float32
	Type  string
}

// Reader holds an open SMC connection and is safe to reuse across scrape cycles.
type Reader struct {
	conn uint
}

// Open opens the SMC connection. Call Close when done.
func Open() (*Reader, error) {
	conn, res := gosmc.SMCOpen(appleSMC)
	if res != gosmc.IOReturnSuccess {
		return nil, smcError("SMCOpen", res)
	}
	return &Reader{conn: conn}, nil
}

// Close releases the SMC connection.
func (r *Reader) Close() {
	gosmc.SMCClose(r.conn)
}

// ReadTemperatures iterates all known temperature sensor keys and returns
// valid (non-zero, non-sentinel) readings. Allocates only the result slice.
func (r *Reader) ReadTemperatures() []Reading {
	// Preallocate with a reasonable capacity to avoid repeated grows.
	results := make([]Reading, 0, len(appleTemp))

	for _, s := range appleTemp {
		if !strings.Contains(s.key, "%") {
			if rd, ok := r.readOne(s.key, s.desc); ok {
				results = append(results, rd)
			}
			continue
		}
		// Expand wildcard: "TC%C" → "TC0C" … "TC9C"
		for i := range 10 {
			expanded := strings.Replace(s.key, "%", itoa1(i), 1)
			desc := strings.Replace(s.desc, "%", itoa1(i+1), 1)
			if rd, ok := r.readOne(expanded, desc); ok {
				results = append(results, rd)
			}
		}
	}

	return results
}

// readOne reads a single key and returns a Reading if the value is valid.
func (r *Reader) readOne(key, desc string) (Reading, bool) {
	val, res := gosmc.SMCReadKey(r.conn, key)
	if res != gosmc.IOReturnSuccess || val.DataSize == 0 {
		return Reading{}, false
	}

	t := smcTypeToString(val.DataType)
	f, err := toFloat32(key, t, val.Bytes, val.DataSize)
	if err != nil {
		return Reading{}, false
	}

	// Filter sentinel / zero values.
	if f == -127.0 || f == 0.0 || math.Round(float64(f)*100)/100 == 0.0 {
		return Reading{}, false
	}
	if f < 0 {
		f = -f
	}

	return Reading{Desc: desc, Key: key, Value: f, Type: t}, true
}

// ── type conversion ──────────────────────────────────────────────────────────

// fpConv holds divisor and signedness for Apple fixed-point SMC types.
type fpConv struct {
	div    float32
	signed bool
}

// appleFixedPoint maps SMC type strings to their conversion parameters.
var appleFixedPoint = map[string]fpConv{
	"fp1f": {32768.0, false},
	"fp2e": {16384.0, false},
	"fp3d": {8192.0, false},
	"fp4c": {4096.0, false},
	"fp5b": {2048.0, false},
	"fp6a": {1024.0, false},
	"fp79": {512.0, false},
	"fp88": {256.0, false},
	"fpa6": {64.0, false},
	"fpc4": {16.0, false},
	"fpe2": {4.0, false},
	"sp1e": {16384.0, true},
	"sp2d": {8192.0, true},
	"sp3c": {4096.0, true},
	"sp4b": {2048.0, true},
	"sp5a": {1024.0, true},
	"sp69": {512.0, true},
	"sp78": {256.0, true},
	"sp87": {128.0, true},
	"sp96": {64.0, true},
	"spa5": {32.0, true},
	"spb4": {16.0, true},
	"spf0": {1.0, true},
}

func toFloat32(key, t string, x gosmc.SMCBytes, size uint32) (float32, error) {
	// Ta0P is mislabeled as 'flt' but uses sp78 format.
	if t == gosmc.TypeFLT && key == "Ta0P" && size >= 2 {
		return fpFixed("sp78", x, size)
	}

	switch t {
	case gosmc.TypeFLT:
		if size < 4 {
			return 0, smcError("flt size", int(size))
		}
		return math.Float32frombits(binary.LittleEndian.Uint32(x[:4])), nil

	case gosmc.TypeUI8, gosmc.TypeUI16, gosmc.TypeUI32, "hex_":
		return float32(smcBytesToUint32(x, size)), nil

	case "ioft":
		if size < 8 {
			return 0, smcError("ioft size", int(size))
		}
		res := binary.LittleEndian.Uint64(x[:8])
		return float32(res) / 65536.0, nil

	default:
		return fpFixed(t, x, size)
	}
}

func fpFixed(t string, x gosmc.SMCBytes, size uint32) (float32, error) {
	v, ok := appleFixedPoint[t]
	if !ok {
		return 0, smcError("unknown fp type", 0)
	}
	if size < 2 {
		return 0, smcError("fp size", int(size))
	}
	raw := binary.BigEndian.Uint16(x[:2])
	if v.signed {
		return float32(int16(raw)) / v.div, nil
	}
	return float32(raw) / v.div, nil
}

func smcBytesToUint32(x gosmc.SMCBytes, size uint32) uint32 {
	var total uint32
	for i := range size {
		total += uint32(x[i]) << ((size - 1 - i) * 8)
	}
	return total
}

func smcTypeToString(x gosmc.UInt32Char) string {
	return strings.TrimRight(x.ToString(), "\x00 ")
}

// itoa1 converts 0–9 to its ASCII digit string without allocating.
var digits = [10]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

func itoa1(i int) string {
	if i >= 0 && i < 10 {
		return digits[i]
	}
	return "0"
}

// smcError is a minimal error type to avoid importing fmt.
type smcErr struct {
	msg string
	val int
}

func (e smcErr) Error() string { return e.msg }

func smcError(msg string, val int) error { return smcErr{msg, val} }
