package main

import (
	"encoding/xml"
	"fmt"
)

func getLogicCheck(inp, nonce string) string {
	if len(inp) < 16 {
		return ""
	}
	out := ""
	for _, c := range nonce {
		out += string(inp[int(c)&0xf])
	}
	return out
}

type FUSMsg struct {
	XMLName xml.Name `xml:"FUSMsg"`
	FUSHdr  FUSHdr   `xml:"FUSHdr"`
	FUSBody FUSBody  `xml:"FUSBody"`
}

type FUSHdr struct {
	ProtoVer string `xml:"ProtoVer"`
}

type FUSBody struct {
	Put FUSPut `xml:"Put"`
}

type FUSPut struct {
	Elements []FUSElement
}

type FUSElement struct {
	XMLName xml.Name
	Data    string `xml:"Data"`
}

func (p FUSPut) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	e.EncodeToken(start)
	for _, elem := range p.Elements {
		elemStart := xml.StartElement{Name: elem.XMLName}
		e.EncodeToken(elemStart)
		e.EncodeElement(elem.Data, xml.StartElement{Name: xml.Name{Local: "Data"}})
		e.EncodeToken(elemStart.End())
	}
	e.EncodeToken(start.End())
	return nil
}

func binaryInform(fwv, model, region, imei, nonce string) string {
	elements := []FUSElement{
		{XMLName: xml.Name{Local: "ACCESS_MODE"}, Data: "2"},
		{XMLName: xml.Name{Local: "BINARY_NATURE"}, Data: "1"},
		{XMLName: xml.Name{Local: "CLIENT_PRODUCT"}, Data: "Smart Switch"},
		{XMLName: xml.Name{Local: "DEVICE_FW_VERSION"}, Data: fwv},
		{XMLName: xml.Name{Local: "DEVICE_LOCAL_CODE"}, Data: region},
		{XMLName: xml.Name{Local: "DEVICE_MODEL_NAME"}, Data: model},
		{XMLName: xml.Name{Local: "UPGRADE_VARIABLE"}, Data: "0"},
		{XMLName: xml.Name{Local: "OBEX_SUPPORT"}, Data: "0"},
		{XMLName: xml.Name{Local: "DEVICE_IMEI_PUSH"}, Data: imei},
		{XMLName: xml.Name{Local: "DEVICE_PLATFORM"}, Data: "Android"},
		{XMLName: xml.Name{Local: "CLIENT_VERSION"}, Data: "4.3.23123_1"},
		{XMLName: xml.Name{Local: "LOGIC_CHECK"}, Data: getLogicCheck(fwv, nonce)},
	}

	if region == "EUX" {
		elements = append(elements,
			FUSElement{XMLName: xml.Name{Local: "DEVICE_AID_CODE"}, Data: region},
			FUSElement{XMLName: xml.Name{Local: "DEVICE_CC_CODE"}, Data: "DE"},
			FUSElement{XMLName: xml.Name{Local: "MCC_NUM"}, Data: "262"},
			FUSElement{XMLName: xml.Name{Local: "MNC_NUM"}, Data: "01"},
		)
	} else if region == "EUY" {
		elements = append(elements,
			FUSElement{XMLName: xml.Name{Local: "DEVICE_AID_CODE"}, Data: region},
			FUSElement{XMLName: xml.Name{Local: "DEVICE_CC_CODE"}, Data: "RS"},
			FUSElement{XMLName: xml.Name{Local: "MCC_NUM"}, Data: "220"},
			FUSElement{XMLName: xml.Name{Local: "MNC_NUM"}, Data: "01"},
		)
	}

	msg := FUSMsg{
		FUSHdr: FUSHdr{ProtoVer: "1.0"},
		FUSBody: FUSBody{
			Put: FUSPut{Elements: elements},
		},
	}

	data, _ := xml.Marshal(msg)
	return string(data)
}

func binaryInit(filename, nonce string) string {
	checkInp := filename
	if len(filename) > 16 {
		base := filename
		for i := len(filename) - 1; i >= 0; i-- {
			if filename[i] == '.' {
				base = filename[:i]
				break
			}
		}
		if len(base) >= 16 {
			checkInp = base[len(base)-16:]
		}
	}

	elements := []FUSElement{
		{XMLName: xml.Name{Local: "BINARY_FILE_NAME"}, Data: filename},
		{XMLName: xml.Name{Local: "LOGIC_CHECK"}, Data: getLogicCheck(checkInp, nonce)},
	}

	msg := FUSMsg{
		FUSHdr: FUSHdr{ProtoVer: "1.0"},
		FUSBody: FUSBody{
			Put: FUSPut{Elements: elements},
		},
	}

	data, _ := xml.Marshal(msg)
	return string(data)
}

func getCheckInput(filename string) string {
	base := filename
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			base = filename[:i]
			break
		}
	}
	if len(base) < 16 {
		return base
	}
	return base[len(base)-16:]
}

func init() {
	_ = fmt.Sprintf
}
