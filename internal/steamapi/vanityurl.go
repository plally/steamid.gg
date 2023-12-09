package steamapi

const VanityURLSuccess = 1
const VanityURLNotFound = 42

type VanityURLResponse struct {
	SteamID string `json:"steamid"`
	Success int    `json:"success"`
	Message string `json:"message"`
}

func (s *SteamAPI) ResolveVanityURL(vanityURL string) (VanityURLResponse, error) {
	var vanityURLResponse Response[VanityURLResponse]
	err := s.Get("ISteamUser/ResolveVanityURL/v0001", map[string]string{
		"vanityurl": vanityURL,
	}, &vanityURLResponse)
	if err != nil {
		return VanityURLResponse{}, err
	}

	return vanityURLResponse.Response, nil
}
