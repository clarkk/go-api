package api

import (
	"math"
	"github.com/go-json-experiment/json"
)

type (
	Limit struct {
		offset		uint32
		limit		uint8
		entries		uint32
	}
	
	limit_json struct {
		Offset		uint32		`json:"offset"`
		Limit		uint8		`json:"limit"`
		Entries		uint32		`json:"entries"`
	}
)

func NewLimit(offset, limit int) (*Limit, bool){
	if limit < 0 || limit > math.MaxUint8 {
		return nil, true
	}
	if offset < 0 || offset > math.MaxUint32 {
		return nil, true
	}
	return &Limit{
		offset:	uint32(offset),
		limit:	uint8(limit),
	}, false
}

func (l *Limit) Offset() uint32 {
	return l.offset
}

func (l *Limit) Limit() uint8 {
	return l.limit
}

func (l *Limit) Limit_max(max uint8){
	l.limit = min(l.limit, max)
}

func (l *Limit) Count(count uint32) bool {
	l.entries = count
	//	Offset out of range
	if l.offset != 0 && l.offset + 1 > count {
		return false
	}
	return true
}

func (l Limit) MarshalJSON() ([]byte, error){
	return json.Marshal(limit_json{
		Offset:		l.offset,
		Limit:		l.limit,
		Entries:	l.entries,
	})
}
