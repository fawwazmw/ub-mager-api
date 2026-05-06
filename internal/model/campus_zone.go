package model

type CampusZone string

const (
	ZoneFILKOM   CampusZone = "FILKOM"
	ZoneFTP      CampusZone = "FTP"
	ZoneFEB      CampusZone = "FEB"
	ZoneFH       CampusZone = "FH"
	ZoneFK       CampusZone = "FK"
	ZoneFKG      CampusZone = "FKG"
	ZoneFIA      CampusZone = "FIA"
	ZoneFISIP    CampusZone = "FISIP"
	ZoneFMIPA    CampusZone = "FMIPA"
	ZoneFPIK     CampusZone = "FPIK"
	ZoneFT       CampusZone = "FT"
	ZoneFPET     CampusZone = "FPET"
	ZoneRektorat CampusZone = "REKTORAT"
	ZoneGOR      CampusZone = "GOR"
	ZoneOther    CampusZone = "OTHER"
)

var ValidCampusZones = []CampusZone{
	ZoneFILKOM, ZoneFTP, ZoneFEB, ZoneFH, ZoneFK, ZoneFKG,
	ZoneFIA, ZoneFISIP, ZoneFMIPA, ZoneFPIK, ZoneFT, ZoneFPET,
	ZoneRektorat, ZoneGOR, ZoneOther,
}
