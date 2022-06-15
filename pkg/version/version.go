package version

import (
	"strconv"
	"time"
)

type CniVersion struct {
	Version string
}

var Version string

func init() {
	currentTime := time.Now()
	_, month, day := currentTime.Date()
	Version = month.String() + "/" + strconv.Itoa(day)
}

func GetCniVersion() CniVersion {
	cniver := CniVersion{
		Version: Version,
	}
	return cniver
}
