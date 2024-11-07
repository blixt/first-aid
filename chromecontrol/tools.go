package chromecontrol

import (
	"encoding/json"
	"fmt"

	"github.com/blixt/go-llms/llms"
	"github.com/blixt/go-llms/tools"
)

type ListTabsParams struct {
}

type SetActiveTabParams struct {
	ID int `json:"id"`
}

type OpenTabParams struct {
	URL        string `json:"url" description:"Must be a URL including the schema"`
	Background bool   `json:"background,omitempty"`
}

type ScreenshotTabParams struct {
	ID int `json:"id"`
}

type SearchWebParams struct {
	Query      string `json:"query"`
	Background bool   `json:"background,omitempty"`
}

func (s *Server) AddToolsToLLM(model *llms.LLM) {
	var t tools.Tool

	t = tools.Func("List browser tabs", "List info about the tabs in the browser, including their ids", "browser_list_tabs", func(r tools.Runner, params ListTabsParams) tools.Result {
		tabs, err := s.GetTabs()
		if err != nil {
			return tools.Error("List browser tabs", err)
		}
		jsonData, err := json.Marshal(tabs)
		if err != nil {
			return tools.Error("List browser tabs", err)
		}
		return tools.Success("List browser tabs", string(jsonData))
	})
	model.AddTool(t)

	t = tools.Func("Set active browser tab", "Switch to the browser tab with the specified id", "browser_set_active_tab", func(r tools.Runner, params SetActiveTabParams) tools.Result {
		err := s.SetActiveTab(params.ID)
		if err != nil {
			return tools.Error("Set active tab", err)
		}
		return tools.Success("Set active tab", "Tab set successfully")
	})
	model.AddTool(t)

	t = tools.Func("Open new tab", "Open a new tab in the browser", "browser_open_tab", func(r tools.Runner, params OpenTabParams) tools.Result {
		tabID, err := s.OpenTab(params.URL, params.Background)
		if err != nil {
			return tools.Error("Open new tab", err)
		}
		var content string
		if params.Background {
			content = fmt.Sprintf("Opened a new tab in the background. To switch to it or screenshot it, use id %d.", tabID)
		} else {
			content = fmt.Sprintf("Opened a new tab. To screenshot it, use id %d.", tabID)
		}
		return tools.Success("Open new tab", content)
	})
	model.AddTool(t)

	t = tools.Func("Look at browser tab", "Activate and take a screenshot of the specified tab in the browser", "browser_screenshot_tab", func(r tools.Runner, params ScreenshotTabParams) tools.Result {
		dataURI, err := s.ScreenshotTab(params.ID)
		if err != nil {
			return tools.Error("Screenshot tab", err)
		}
		var rb tools.ResultBuilder
		rb.AddImageURL("screenshot.png", dataURI)
		return rb.Success("Screenshot browser tab", "You will receive screenshot.png from the user as an automated message.")
	})
	model.AddTool(t)

	t = tools.Func("Search the web", "Search the web using the default search provider", "browser_search_web", func(r tools.Runner, params SearchWebParams) tools.Result {
		tabID, err := s.SearchWeb(params.Query, params.Background)
		if err != nil {
			return tools.Error("Search the web", err)
		}
		var content string
		if params.Background {
			content = fmt.Sprintf("A tab with the search results was opened in the background. To switch to it or screenshot it, use id %d.", tabID)
		} else {
			content = fmt.Sprintf("The search results are now in a new active tab. To screenshot it, use id %d.", tabID)
		}
		return tools.Success("Search the web", content)
	})
	model.AddTool(t)
}
