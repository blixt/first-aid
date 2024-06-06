package chromecontrol

import (
	"encoding/json"
)

type TabDetails struct {
	ID    int    `json:"id"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

func (s *Server) GetTabs() ([]TabDetails, error) {
	result, err := s.sendRPC("getTabs", nil)
	if err != nil {
		return nil, err
	}

	var tabs []TabDetails
	if err := json.Unmarshal(result, &tabs); err != nil {
		return nil, err
	}
	return tabs, nil
}

func (s *Server) SetActiveTab(id int) error {
	_, err := s.sendRPC("setActiveTab", id)
	return err
}

func (s *Server) OpenTab(url string, background bool) (int, error) {
	params := map[string]interface{}{
		"url":        url,
		"background": background,
	}
	result, err := s.sendRPC("openTab", params)
	if err != nil {
		return 0, err
	}

	var tabID int
	if err := json.Unmarshal(result, &tabID); err != nil {
		return 0, err
	}

	return tabID, nil
}

func (s *Server) SearchWeb(query string, background bool) (int, error) {
	params := map[string]interface{}{
		"query":      query,
		"background": background,
	}
	result, err := s.sendRPC("searchWeb", params)
	if err != nil {
		return 0, err
	}

	var tabID int
	if err := json.Unmarshal(result, &tabID); err != nil {
		return 0, err
	}

	return tabID, nil
}

func (s *Server) ScreenshotTab(id int) (string, error) {
	params := map[string]interface{}{
		"id": id,
	}
	result, err := s.sendRPC("screenshotTab", params)
	if err != nil {
		return "", err
	}

	var base64Screenshot string
	if err := json.Unmarshal(result, &base64Screenshot); err != nil {
		return "", err
	}

	return base64Screenshot, nil
}
