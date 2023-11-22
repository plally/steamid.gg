package steamid

import (
	"fmt"
	"strconv"
)

type SteamID uint64

const magicalSteamNumber = 76561197960265728

func SteamID32(steamString string) (SteamID, error) {
	Y, err := strconv.Atoi(steamString[8:9])
	if err != nil {
		return 0, err
	}

	Z, err := strconv.Atoi(steamString[10:])
	if err != nil {
		return 0, err
	}
	i := int64((Z * 2) + magicalSteamNumber + Y)

	return SteamID(i), nil
}

func SteamID64(steamString string) (SteamID, error) {
	i, err := strconv.ParseInt(steamString, 10, 64)
	if err != nil {
		return 0, err
	}

	if i < magicalSteamNumber {
		return 0, fmt.Errorf("SteamID64 is too small")
	}

	return SteamID(i), nil
}

func (s SteamID) SteamID32String() string {
	s = s - magicalSteamNumber
	remainder := s % 2
	s = s / 2
	return fmt.Sprintf("STEAM_0:%d:%d", remainder, s)
}

func (s SteamID) SteamID64String() string {
	return strconv.FormatInt(int64(s), 10)
}

func (s SteamID) SteamID3String() string {
	s = s - magicalSteamNumber
	return fmt.Sprintf("[U:1:%d]", s)
}
