package ublox

// Profile описывает профиль возвращаемых данных для данного устройства.
type Profile struct {
	Datatype    []string // A comma separated list of the data types required by the client (eph, alm, aux, pos)
	Format      string   // Specifies the format of the data returned (mga = UBX- MGA-* (M8 onwards); aid = UBX-AID-* (u7 or earlier))
	GNSS        []string // A comma separated list of the GNSS for which data should be returned (gps, qzss, glo)
	FilterOnPos bool     // If present, the ephemeris data returned to the client will only contain data for the satellites which are likely to be visible from the approximate position provided
}

// DefaultProfile описывает профиль запроса по умолчанию для запросов.
var DefaultProfile = Profile{
	Datatype:    []string{"pos", "eph", "aux"},
	Format:      "aid",
	GNSS:        []string{"gps"},
	FilterOnPos: true,
}
