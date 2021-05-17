package tenhou

import (
	"encoding/xml"
)

// 需要注意的是，牌譜並未記錄捨牌是手切還是摸切，
// 這裏認為在摸牌後，只要切出的牌和摸的牌相同就認為是摸切，否則認為是手切
type RecordAction struct {
	XMLName xml.Name
	message
}

func (a *RecordAction) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	a.Tag = start.Name.Local
	type action RecordAction // 防止無限遞歸
	return d.DecodeElement((*action)(a), &start)
}

type Record struct {
	XMLName xml.Name        `xml:"mjloggm"`
	Actions []*RecordAction `xml:",any"`
}
