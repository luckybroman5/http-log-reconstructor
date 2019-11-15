// Copyright 2019+ Klaus Post. All rights reserved.
// License information can be found in the LICENSE file.
// Based on work by Yann Collet, released under BSD License.

package zstd

import (
	"errors"
	"fmt"
	"math"
	"math/bits"

	"github.com/klauspost/compress/huff0"
)

type blockEnc struct {
	size      int
	literals  []byte
	sequences []seq
	coders    seqCoders
	litEnc    *huff0.Scratch
	wr        bitWriter

	extraLits int
	last      bool

	output            []byte
	recentOffsets     [3]uint32
	prevRecentOffsets [3]uint32
}

// init should be used once the block has been created.
// If called more than once, the effect is the same as calling reset.
func (b *blockEnc) init() {
	if cap(b.literals) < maxCompressedLiteralSize {
		b.literals = make([]byte, 0, maxCompressedLiteralSize)
	}
	const defSeqs = 200
	b.literals = b.literals[:0]
	if cap(b.sequences) < defSeqs {
		b.sequences = make([]seq, 0, defSeqs)
	}
	if cap(b.output) < maxCompressedBlockSize {
		b.output = make([]byte, 0, maxCompressedBlockSize)
	}
	if b.coders.mlEnc == nil {
		b.coders.mlEnc = &fseEncoder{}
		b.coders.mlPrev = &fseEncoder{}
		b.coders.ofEnc = &fseEncoder{}
		b.coders.ofPrev = &fseEncoder{}
		b.coders.llEnc = &fseEncoder{}
		b.coders.llPrev = &fseEncoder{}
	}
	b.litEnc = &huff0.Scratch{}
	b.reset(nil)
}

// initNewEncode can be used to reset offsets and encoders to the initial state.
func (b *blockEnc) initNewEncode() {
	b.recentOffsets = [3]uint32{1, 4, 8}
	b.litEnc.Reuse = huff0.ReusePolicyNone
	b.coders.setPrev(nil, nil, nil)
}

// reset will reset the block for a new encode, but in the same stream,
// meaning that state will be carried over, but the block content is reset.
// If a previous block is provided, the recent offsets are carried over.
func (b *blockEnc) reset(prev *blockEnc) {
	b.extraLits = 0
	b.literals = b.literals[:0]
	b.size = 0
	b.sequences = b.sequences[:0]
	b.output = b.output[:0]
	b.last = false
	if prev != nil {
		b.recentOffsets = prev.prevRecentOffsets
	}
}

// reset will reset the block for a new encode, but in the same stream,
// meaning that state will be carried over, but the block content is reset.
// If a previous block is provided, the recent offsets are carried over.
func (b *blockEnc) swapEncoders(prev *blockEnc) {
	b.coders.swap(&prev.coders)
	b.litEnc, prev.litEnc = prev.litEnc, b.litEnc
}

// blockHeader contains the information for a block header.
type blockHeader uint32

// setLast sets the 'last' indicator on a block.
func (h *blockHeader) setLast(b bool) {
	if b {
		*h = *h | 1
	} else {
		const mask = (1 << 24) - 2
		*h = *h & mask
	}
}

// setSize will store the compressed size of a block.
func (h *blockHeader) setSize(v uint32) {
	const mask = 7
	*h = (*h)&mask | blockHeader(v<<3)
}

// setType sets the block type.
func (h *blockHeader) setType(t blockType) {
	const mask = 1 | (((1 << 24) - 1) ^ 7)
	*h = (*h & mask) | blockHeader(t<<1)
}

// appendTo will append the block header to a slice.
func (h blockHeader) appendTo(b []byte) []byte {
	return append(b, uint8(h), uint8(h>>8), uint8(h>>16))
}

// String returns a string representation of the block.
func (h blockHeader) String() string {
	return fmt.Sprintf("Type: %d, Size: %d, Last:%t", (h>>1)&3, h>>3, h&1 == 1)
}

// literalsHeader contains literals header information.
type literalsHeader uint64

// setType can be used to set the type of literal block.
func (h *literalsHeader) setType(t literalsBlockType) {
	const mask = math.MaxUint64 - 3
	*h = (*h & mask) | literalsHeader(t)
}

// setSize can be used to set a single size, for uncompressed and RLE content.
func (h *literalsHeader) setSize(regenLen int) {
	inBits := bits.Len32(uint32(regenLen))
	// Only retain 2 bits
	const mask = 3
	lh := uint64(*h & mask)
	switch {
	case inBits < 5:
		lh |= (uint64(regenLen) << 3) | (1 << 60)
		if debug {
			got := int(lh>>3) & 0xff
			if got != regenLen {
				panic(fmt.Sprint("litRegenSize = ", regenLen, "(want) != ", got, "(got)"))
			}
		}
	case inBits < 12:
		lh |= (1 << 2) | (uint64(regenLen) << 4) | (2 << 60)
	case inBits < 20:
		lh |= (3 << 2) | (uint64(regenLen) << 4) | (3 << 60)
	default:
		panic(fmt.Errorf("internal error: block too big (%d)", regenLen))
	}
	*h = literalsHeader(lh)
}

// setSizes will set the size of a compressed literals section and the input length.
func (h *literalsHeader) setSizes(compLen, inLen int) {
	compBits, inBits := bits.Len32(uint32(compLen)), bits.Len32(uint32(inLen))
	// Only retain 2 bits
	const mask = 3
	lh := uint64(*h & mask)
	switch {
	case compBits <= 10 && inBits <= 10:
		lh |= (1 << 2) | (uint64(inLen) << 4) | (uint64(compLen) << (10 + 4)) | (3 << 60)
		if debug {
			const mmask = (1 << 24) - 1
			n := (lh >> 4) & mmask
			if int(n&1023) != inLen {
				panic(fmt.Sprint("regensize:", int(n&1023), "!=", inLen, inBits))
			}
			if int(n>>10) != compLen {
				panic(fmt.Sprint("compsize:", int(n>>10), "!=", compLen, compBits))
			}
		}
	case compBits <= 14 && inBits <= 14:
		lh |= (2 << 2) | (uint64(inLen) << 4) | (uint64(compLen) << (14 + 4)) | (4 << 60)
	case compBits <= 18 && inBits <= 18:
		lh |= (3 << 2) | (uint64(inLen) << 4) | (uint64(compLen) << (18 + 4)) | (5 << 60)
	default:
		panic("internal error: block too big")
	}
	*h = literalsHeader(lh)
}

// appendTo will append the literals header to a byte slice.
func (h literalsHeader) appendTo(b []byte) []byte {
	size := uint8(h >> 60)
	switch size {
	case 1:
		b = append(b, uint8(h))
	case 2:
		b = append(b, uint8(h), uint8(h>>8))
	case 3:
		b = append(b, uint8(h), uint8(h>>8), uint8(h>>16))
	case 4:
		b = append(b, uint8(h), uint8(h>>8), uint8(h>>16), uint8(h>>24))
	case 5:
		b = append(b, uint8(h), uint8(h>>8), uint8(h>>16), uint8(h>>24), uint8(h>>32))
	default:
		panic(fmt.Errorf("internal error: literalsHeader has invalid size (%d)", size))
	}
	return b
}

// size returns the output size with currently set values.
func (h literalsHeader) size() int {
	return int(h >> 60)
}

func (h literalsHeader) String() string {
	return fmt.Sprintf("Type: %d, SizeFormat: %d, Size: 0x%d, Bytes:%d", literalsBlockType(h&3), (h>>2)&3, h&((1<<60)-1)>>4, h>>60)
}

// pushOffsets will push the recent offsets to the backup store.
func (b *blockEnc) pushOffsets() {
	b.prevRecentOffsets = b.recentOffsets
}

// pushOffsets will push the recent offsets to the backup store.
func (b *blockEnc) popOffsets() {
	b.recentOffsets = b.prevRecentOffsets
}

// matchOffset will adjust recent offsets and return the adjusted one,
// if it matches a previous offset.
func (b *blockEnc) matchOffset(offset, lits uint32) uint32 {
	// Check if offset is one of the recent offsets.
	// Adjusts the output offset accordingly.
	// Gives a tiny bit of compression, typically around 1%.
	if true {
		if lits > 0 {
			switch offset {
			case b.recentOffsets[0]:
				offset = 1
			case b.recentOffsets[1]:
				b.recentOffsets[1] = b.recentOffsets[0]
				b.recentOffsets[0] = offset
				offset = 2
			case b.recentOffsets[2]:
				b.recentOffsets[2] = b.recentOffsets[1]
				b.recentOffsets[1] = b.recentOffsets[0]
				b.recentOffsets[0] = offset
				offset = 3
			default:
				b.recentOffsets[2] = b.recentOffsets[1]
				b.recentOffsets[1] = b.recentOffsets[0]
				b.recentOffsets[0] = offset
				offset += 3
			}
		} else {
			switch offset {
			case b.recentOffsets[1]:
				b.recentOffsets[1] = b.recentOffsets[0]
				b.recentOffsets[0] = offset
				offset = 1
			case b.recentOffsets[2]:
				b.recentOffsets[2] = b.recentOffsets[1]
				b.recentOffsets[1] = b.recentOffsets[0]
				b.recentOffsets[0] = offset
				offset = 2
			case b.recentOffsets[0] - 1:
				b.recentOffsets[2] = b.recentOffsets[1]
				b.recentOffsets[1] = b.recentOffsets[0]
				b.recentOffsets[0] = offset
				offset = 3
			default:
				b.recentOffsets[2] = b.recentOffsets[1]
				b.recentOffsets[1] = b.recentOffsets[0]
				b.recentOffsets[0] = offset
				offset += 3
			}
		}
	} else {
		offset += 3
	}
	return offset
}

// encodeRaw can be used to set the output to a raw representation of supplied bytes.
func (b *blockEnc) encodeRaw(a []byte) {
	var bh blockHeader
	bh.setLast(b.last)
	bh.setSize(uint32(len(a)))
	bh.setType(blockTypeRaw)
	b.output = bh.appendTo(b.output[:0])
	b.output = append(b.output, a...)
	if debug {
		println("Adding RAW block, length", len(a))
	}
}

// encodeLits can be used if the block is only litLen.
func (b *blockEnc) encodeLits() error {
	var bh blockHeader
	bh.setLast(b.last)
	bh.setSize(uint32(len(b.literals)))

	// Don't compress extremely small blocks
	if len(b.literals) < 32 {
		if debug {
			println("Adding RAW block, length", len(b.literals))
		}
		bh.setType(blockTypeRaw)
		b.output = bh.appendTo(b.output)
		b.output = append(b.output, b.literals...)
		return nil
	}

	// TODO: Switch to 1X when less than x bytes.
	out, reUsed, err := huff0.Compress4X(b.literals, b.litEnc)
	// Bail out of compression is too little.
	if len(out) > (len(b.literals) - len(b.literals)>>4) {
		err = huff0.ErrIncompressible
	}
	switch err {
	case huff0.ErrIncompressible:
		if debug {
			println("Adding RAW block, length", len(b.literals))
		}
		bh.setType(blockTypeRaw)
		b.output = bh.appendTo(b.output)
		b.output = append(b.output, b.literals...)
		return nil
	case huff0.ErrUseRLE:
		if debug {
			println("Adding RLE block, length", len(b.literals))
		}
		bh.setType(blockTypeRLE)
		b.output = bh.appendTo(b.output)
		b.output = append(b.output, b.literals[0])
		return nil
	default:
		return err
	case nil:
	}
	// Compressed...
	// Now, allow reuse
	b.litEnc.Reuse = huff0.ReusePolicyAllow
	bh.setType(blockTypeCompressed)
	var lh literalsHeader
	if reUsed {
		if debug {
			println("Reused tree, compressed to", len(out))
		}
		lh.setType(literalsBlockTreeless)
	} else {
		if debug {
			println("New tree, compressed to", len(out), "tree size:", len(b.litEnc.OutTable))
		}
		lh.setType(literalsBlockCompressed)
	}
	// Set sizes
	lh.setSizes(len(out), len(b.literals))
	bh.setSize(uint32(len(out) + lh.size() + 1))

	// Write block headers.
	b.output = bh.appendTo(b.output)
	b.output = lh.appendTo(b.output)
	// Add compressed data.
	b.output = append(b.output, out...)
	// No sequences.
	b.output = append(b.output, 0)
	return nil
}

// encode will encode the block and put the output in b.output.
func (b *blockEnc) encode() error {
	if len(b.sequences) == 0 {
		return b.encodeLits()
	}
	// We want some difference
	if len(b.literals) > (b.size - (b.size >> 5)) {
		return errIncompressible
	}

	var bh blockHeader
	var lh literalsHeader
	bh.setLast(b.last)
	bh.setType(blockTypeCompressed)
	b.output = bh.appendTo(b.output)

	var (
		out    []byte
		reUsed bool
		err    error
	)
	if len(b.literals) > 32 {
		// TODO: Switch to 1X on small blocks.
		out, reUsed, err = huff0.Compress4X(b.literals, b.litEnc)
		if len(out) > len(b.literals)-len(b.literals)>>4 {
			err = huff0.ErrIncompressible
		}
	} else {
		err = huff0.ErrIncompressible
	}
	switch err {
	case huff0.ErrIncompressible:
		lh.setType(literalsBlockRaw)
		lh.setSize(len(b.literals))
		b.output = lh.appendTo(b.output)
		b.output = append(b.output, b.literals...)
		if debug {
			println("Adding literals RAW, length", len(b.literals))
		}
	case huff0.ErrUseRLE:
		lh.setType(literalsBlockRLE)
		lh.setSize(len(b.literals))
		b.output = lh.appendTo(b.output)
		b.output = append(b.output, b.literals[0])
		if debug {
			println("Adding literals RLE")
		}
	default:
		if debug {
			println("Adding literals ERROR:", err)
		}
		return err
	case nil:
		// Compressed litLen...
		if reUsed {
			if debug {
				println("reused tree")
			}
			lh.setType(literalsBlockTreeless)
		} else {
			if debug {
				println("new tree, size:", len(b.litEnc.OutTable))
			}
			lh.setType(literalsBlockCompressed)
			if debug {
				_, _, err := huff0.ReadTable(out, nil)
				if err != nil {
					panic(err)
				}
			}
		}
		lh.setSizes(len(out), len(b.literals))
		if debug {
			printf("Compressed %d literals to %d bytes", len(b.literals), len(out))
			println("Adding literal header:", lh)
		}
		b.output = lh.appendTo(b.output)
		b.output = append(b.output, out...)
		b.litEnc.Reuse = huff0.ReusePolicyAllow
		if debug {
			println("Adding literals compressed")
		}
	}
	// Sequence compression

	// Write the number of sequences
	switch {
	case len(b.sequences) < 128:
		b.output = append(b.output, uint8(len(b.sequences)))
	case len(b.sequences) < 0x7f00: // TODO: this could be wrong
		n := len(b.sequences)
		b.output = append(b.output, 128+uint8(n>>8), uint8(n))
	default:
		n := len(b.sequences) - 0x7f00
		b.output = append(b.output, 255, uint8(n), uint8(n>>8))
	}
	if debug {
		println("Encoding", len(b.sequences), "sequences")
	}
	b.genCodes()
	llEnc := b.coders.llEnc
	ofEnc := b.coders.ofEnc
	mlEnc := b.coders.mlEnc
	err = llEnc.normalizeCount(len(b.sequences))
	if err != nil {
		return err
	}
	err = ofEnc.normalizeCount(len(b.sequences))
	if err != nil {
		return err
	}
	err = mlEnc.normalizeCount(len(b.sequences))
	if err != nil {
		return err
	}

	// Choose the best compression mode for each type.
	// Will evaluate the new vs predefined and previous.
	chooseComp := func(cur, prev, preDef *fseEncoder) (*fseEncoder, seqCompMode) {
		// See if predefined/previous is better
		hist := cur.count[:cur.symbolLen]
		nSize := cur.approxSize(hist) + cur.maxHeaderSize()
		predefSize := preDef.approxSize(hist)
		prevSize := prev.approxSize(hist)

		// Add a small penalty for new encoders.
		// Don't bother with extremely small (<2 byte gains).
		nSize = nSize + (nSize+2*8*16)>>4
		switch {
		case predefSize <= prevSize && predefSize <= nSize || forcePreDef:
			if debug {
				println("Using predefined", predefSize>>3, "<=", nSize>>3)
			}
			return preDef, compModePredefined
		case prevSize <= nSize:
			if debug {
				println("Using previous", prevSize>>3, "<=", nSize>>3)
			}
			return prev, compModeRepeat
		default:
			if debug {
				println("Using new, predef", predefSize>>3, ". previous:", prevSize>>3, ">", nSize>>3, "header max:", cur.maxHeaderSize()>>3, "bytes")
				println("tl:", cur.actualTableLog, "symbolLen:", cur.symbolLen, "norm:", cur.norm[:cur.symbolLen], "hist", cur.count[:cur.symbolLen])
			}
			return cur, compModeFSE
		}
	}

	// Write compression mode
	var mode uint8
	if llEnc.useRLE {
		mode |= uint8(compModeRLE) << 6
		llEnc.setRLE(b.sequences[0].llCode)
		if debug {
			println("llEnc.useRLE")
		}
	} else {
		var m seqCompMode
		llEnc, m = chooseComp(llEnc, b.coders.llPrev, &fsePredefEnc[tableLiteralLengths])
		mode |= uint8(m) << 6
	}
	if ofEnc.useRLE {
		mode |= uint8(compModeRLE) << 4
		ofEnc.setRLE(b.sequences[0].ofCode)
		if debug {
			println("ofEnc.useRLE")
		}
	} else {
		var m seqCompMode
		ofEnc, m = chooseComp(ofEnc, b.coders.ofPrev, &fsePredefEnc[tableOffsets])
		mode |= uint8(m) << 4
	}

	if mlEnc.useRLE {
		mode |= uint8(compModeRLE) << 2
		mlEnc.setRLE(b.sequences[0].mlCode)
		if debug {
			println("mlEnc.useRLE, code: ", b.sequences[0].mlCode, "value", b.sequences[0].matchLen)
		}
	} else {
		var m seqCompMode
		mlEnc, m = chooseComp(mlEnc, b.coders.mlPrev, &fsePredefEnc[tableMatchLengths])
		mode |= uint8(m) << 2
	}
	b.output = append(b.output, mode)
	if debug {
		printf("Compression modes: 0b%b", mode)
	}
	b.output, err = llEnc.writeCount(b.output)
	if err != nil {
		return err
	}
	start := len(b.output)
	b.output, err = ofEnc.writeCount(b.output)
	if err != nil {
		return err
	}
	if false {
		println("block:", b.output[start:], "tablelog", ofEnc.actualTableLog, "maxcount:", ofEnc.maxCount)
		fmt.Printf("selected TableLog: %d, Symbol length: %d\n", ofEnc.actualTableLog, ofEnc.symbolLen)
		for i, v := range ofEnc.norm[:ofEnc.symbolLen] {
			fmt.Printf("%3d: %5d -> %4d \n", i, ofEnc.count[i], v)
		}
	}
	b.output, err = mlEnc.writeCount(b.output)
	if err != nil {
		return err
	}

	// Maybe in block?
	wr := &b.wr
	wr.reset(b.output)

	var ll, of, ml cState

	// Current sequence
	seq := len(b.sequences) - 1
	s := b.sequences[seq]
	llEnc.setBits(llBitsTable[:])
	mlEnc.setBits(mlBitsTable[:])
	ofEnc.setBits(nil)

	llTT, ofTT, mlTT := llEnc.ct.symbolTT[:256], ofEnc.ct.symbolTT[:256], mlEnc.ct.symbolTT[:256]

	// We have 3 bounds checks here (and in the loop).
	// Since we are iterating backwards it is kinda hard to avoid.
	llB, ofB, mlB := llTT[s.llCode], ofTT[s.ofCode], mlTT[s.mlCode]
	ll.init(wr, &llEnc.ct, llB)
	of.init(wr, &ofEnc.ct, ofB)
	wr.flush32()
	ml.init(wr, &mlEnc.ct, mlB)

	// Each of these lookups also generates a bounds check.
	wr.addBits32NC(s.litLen, llB.outBits)
	wr.addBits32NC(s.matchLen, mlB.outBits)
	wr.flush32()
	wr.addBits32NC(s.offset, ofB.outBits)
	if debugSequences {
		println("Encoded seq", seq, s, "codes:", s.llCode, s.mlCode, s.ofCode, "states:", ll.state, ml.state, of.state, "bits:", llB, mlB, ofB)
	}
	seq--
	if llEnc.maxBits+mlEnc.maxBits+ofEnc.maxBits <= 32 {
		// No need to flush (common)
		for seq >= 0 {
			s = b.sequences[seq]
			wr.flush32()
			llB, ofB, mlB := llTT[s.llCode], ofTT[s.ofCode], mlTT[s.mlCode]
			// tabelog max is 8 for all.
			of.encode(ofB)
			ml.encode(mlB)
			ll.encode(llB)
			wr.flush32()

			// We checked that all can stay within 32 bits
			wr.addBits32NC(s.litLen, llB.outBits)
			wr.addBits32NC(s.matchLen, mlB.outBits)
			wr.addBits32NC(s.offset, ofB.outBits)

			if debugSequences {
				println("Encoded seq", seq, s)
			}

			seq--
		}
	} else {
		for seq >= 0 {
			s = b.sequences[seq]
			wr.flush32()
			llB, ofB, mlB := llTT[s.llCode], ofTT[s.ofCode], mlTT[s.mlCode]
			// tabelog max is below 8 for each.
			of.encode(ofB)
			ml.encode(mlB)
			ll.encode(llB)
			wr.flush32()

			// ml+ll = max 32 bits total
			wr.addBits32NC(s.litLen, llB.outBits)
			wr.addBits32NC(s.matchLen, mlB.outBits)
			wr.flush32()
			wr.addBits32NC(s.offset, ofB.outBits)

			if debugSequences {
				println("Encoded seq", seq, s)
			}

			seq--
		}
	}
	ml.flush(mlEnc.actualTableLog)
	of.flush(ofEnc.actualTableLog)
	ll.flush(llEnc.actualTableLog)
	err = wr.close()
	if err != nil {
		return err
	}
	b.output = wr.out

	if len(b.output)-3 >= b.size {
		// Maybe even add a bigger margin.
		b.litEnc.Reuse = huff0.ReusePolicyNone
		return errIncompressible
	}

	// Size is output minus block header.
	bh.setSize(uint32(len(b.output)) - 3)
	if debug {
		println("Rewriting block header", bh)
	}
	_ = bh.appendTo(b.output[:0])
	b.coders.setPrev(llEnc, mlEnc, ofEnc)
	return nil
}

var errIncompressible = errors.New("uncompressible")

func (b *blockEnc) genCodes() {
	if len(b.sequences) == 0 {
		// nothing to do
		return
	}

	if len(b.sequences) > math.MaxUint16 {
		panic("can only encode up to 64K sequences")
	}
	// No bounds checks after here:
	llH := b.coders.llEnc.Histogram()[:256]
	ofH := b.coders.ofEnc.Histogram()[:256]
	mlH := b.coders.mlEnc.Histogram()[:256]
	for i := range llH {
		llH[i] = 0
	}
	for i := range ofH {
		ofH[i] = 0
	}
	for i := range mlH {
		mlH[i] = 0
	}

	var llMax, ofMax, mlMax uint8
	for i, seq := range b.sequences {
		v := llCode(seq.litLen)
		seq.llCode = v
		llH[v]++
		if v > llMax {
			llMax = v
		}

		v = ofCode(seq.offset)
		seq.ofCode = v
		ofH[v]++
		if v > ofMax {
			ofMax = v
		}

		v = mlCode(seq.matchLen)
		seq.mlCode = v
		mlH[v]++
		if v > mlMax {
			mlMax = v
			if debug && mlMax > maxMatchLengthSymbol {
				panic(fmt.Errorf("mlMax > maxMatchLengthSymbol (%d), matchlen: %d", mlMax, seq.matchLen))
			}
		}
		b.sequences[i] = seq
	}
	maxCount := func(a []uint32) int {
		var max uint32
		for _, v := range a {
			if v > max {
				max = v
			}
		}
		return int(max)
	}
	if mlMax > maxMatchLengthSymbol {
		panic(fmt.Errorf("mlMax > maxMatchLengthSymbol (%d)", mlMax))
	}
	if ofMax > maxOffsetBits {
		panic(fmt.Errorf("ofMax > maxOffsetBits (%d)", ofMax))
	}
	if llMax > maxLiteralLengthSymbol {
		panic(fmt.Errorf("llMax > maxLiteralLengthSymbol (%d)", llMax))
	}

	b.coders.mlEnc.HistogramFinished(mlMax, maxCount(mlH[:mlMax+1]))
	b.coders.ofEnc.HistogramFinished(ofMax, maxCount(ofH[:ofMax+1]))
	b.coders.llEnc.HistogramFinished(llMax, maxCount(llH[:llMax+1]))
}
