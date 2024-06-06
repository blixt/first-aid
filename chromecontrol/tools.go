package chromecontrol

import (
	"encoding/json"
	"fmt"

	"github.com/blixt/first-aid/llm"
	"github.com/blixt/first-aid/tool"
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

func (s *Server) AddToolsToLLM(model *llm.LLM) {
	var t tool.Tool

	t = tool.Func("List browser tabs", "List info about the tabs in the browser, including their ids", "browser_list_tabs", func(r tool.Runner, params ListTabsParams) tool.Result {
		tabs, err := s.GetTabs()
		if err != nil {
			return tool.Error("List browser tabs", err)
		}
		jsonData, err := json.Marshal(tabs)
		if err != nil {
			return tool.Error("List browser tabs", err)
		}
		return tool.Success("List browser tabs", string(jsonData))
	})
	model.AddTool(t)

	t = tool.Func("Set active browser tab", "Switch to the browser tab with the specified id", "browser_set_active_tab", func(r tool.Runner, params SetActiveTabParams) tool.Result {
		err := s.SetActiveTab(params.ID)
		if err != nil {
			return tool.Error("Set active tab", err)
		}
		return tool.Success("Set active tab", "Tab set successfully")
	})
	model.AddTool(t)

	t = tool.Func("Open new tab", "Open a new tab in the browser", "browser_open_tab", func(r tool.Runner, params OpenTabParams) tool.Result {
		tabID, err := s.OpenTab(params.URL, params.Background)
		if err != nil {
			return tool.Error("Open new tab", err)
		}
		var content string
		if params.Background {
			content = fmt.Sprintf("Opened a new tab in the background. To switch to it or screenshot it, use id %d.", tabID)
		} else {
			content = fmt.Sprintf("Opened a new tab. To screenshot it, use id %d.", tabID)
		}
		return tool.Success("Open new tab", content)
	})
	model.AddTool(t)

	t = tool.Func("Look at browser tab", "Activate and take a screenshot of the specified tab in the browser", "browser_screenshot_tab", func(r tool.Runner, params ScreenshotTabParams) tool.Result {
		dataURI, err := s.ScreenshotTab(params.ID)
		if err != nil {
			return tool.Error("Screenshot tab", err)
		}
		var rb tool.ResultBuilder
		rb.AddImageURL("screenshot.png", dataURI)
		return rb.Success("Screenshot browser tab", "You will receive screenshot.png from the user as an automated message.")
	})
	model.AddTool(t)

	t = tool.Func("Search the web", "Search the web using the default search provider", "browser_search_web", func(r tool.Runner, params SearchWebParams) tool.Result {
		tabID, err := s.SearchWeb(params.Query, params.Background)
		if err != nil {
			return tool.Error("Search the web", err)
		}
		var content string
		if params.Background {
			content = fmt.Sprintf("A tab with the search results was opened in the background. To switch to it or screenshot it, use id %d.", tabID)
		} else {
			content = fmt.Sprintf("The search results are now in a new active tab. To screenshot it, use id %d.", tabID)
		}
		return tool.Success("Search the web", content)
	})
	model.AddTool(t)
}
