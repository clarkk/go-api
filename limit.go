package api

import "math"

type Limit struct {
	Offset		uint32		`json:"offset"`
	Limit		uint8		`json:"limit"`
}

func (l *Limit) Max(max uint8){
	l.Limit = min(l.Limit, max)
	if l.Offset != 0 {
		f := float64(l.Offset) / float64(l.Limit)
		l.Offset = uint32(math.Round(f)) * uint32(l.Limit)
	}
}

func (l *Limit) Count(){
	
}