package api

//import "math"

type Limit struct {
	Offset		uint32		`json:"offset"`
	Limit		uint8		`json:"limit"`
	Entries		uint32		`json:"entries"`
}

func (l *Limit) Limit_max(max uint8){
	l.Limit = min(l.Limit, max)
	//	Adjust relationship between offset and limit (pagination)
	/*if l.Offset != 0 {
		f := float64(l.Offset) / float64(l.Limit)
		l.Offset = uint32(math.Round(f)) * uint32(l.Limit)
	}*/
}

func (l *Limit) Count(count uint32) bool {
	l.Entries = count
	//	Offset out of range
	if l.Offset != 0 && l.Offset + 1 > count {
		return false
		//	Adjust offset to last page
		/*f := float64(l.Entries) / float64(l.Limit)
		l.Offset = uint32(math.Floor(f)) * uint32(l.Limit)*/
	}
	return true
}