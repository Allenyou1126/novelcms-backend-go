package utils

import (
	"sync"
	"time"
)

type IdGenerator struct {
	mu         sync.Mutex
	counter    int64
	symbol     int64
	symbolBit  int64
	counterBit int64
	lastGen    int64
}

func (g *IdGenerator) Next() int64 {
	g.mu.Lock()
	var ret int64 = 0
	timeBits := (time.Now().Unix() % (1 << (64 - 1 - g.symbolBit - g.counterBit))) << (g.symbolBit + g.counterBit)
	if (timeBits & g.lastGen) != 0 {
		g.counter += 1
	} else {
		g.counter = 1
	}
	ret |= timeBits
	ret |= (g.counter) << g.symbolBit
	ret |= g.symbol
	g.mu.Unlock()
	return ret
}

func GetIdGenerator(symbolBit int64, counterBit int64, symbol int64) IdGenerator {
	return IdGenerator{
		counter:    0,
		counterBit: counterBit,
		symbol:     symbol,
		symbolBit:  symbolBit,
		lastGen:    0,
	}
}
