package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/workspace"
)

// updateEnvironmentDropdown updates the environment dropdown with current environments
func updateEnvironmentDropdown(dropdown *tview.DropDown, environments []workspace.Environment) {
	options := []string{"Base Environment"}
	for _, env := range environments {
		options = append(options, env.Name)
	}
	dropdown.SetOptions(options, nil)
}

// saveCurrentRequest saves the current request changes to workspace
func saveCurrentRequest(currentRequest *workspace.Request, workspaceData *workspace.Workspace) {
	if currentRequest != nil {
		// Sync headers from UI before saving
		currentRequest.Headers = getHeadersFromUI()
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error (could show in status or log)
			return
		}
	}
}

// syncMethodDropdown syncs the method dropdown with the current request's method
func syncMethodDropdown(currentRequest *workspace.Request, methodDropdown *tview.DropDown, programmaticallyUpdatingMethod *bool) {
	if currentRequest != nil {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
		for i, method := range methods {
			if method == currentRequest.Method {
				// Set flag to prevent the SetSelectedFunc from firing
				*programmaticallyUpdatingMethod = true
				methodDropdown.SetCurrentOption(i)
				*programmaticallyUpdatingMethod = false
				break
			}
		}
	}
}

// updateResponseTabs updates the response tabs with new response data
func updateResponseTabs(resp *HTTPResponse, lastTime *time.Time, response *tview.Flex, responseTabHeader *tview.Flex, responseInfoBar **tview.Flex, responseTimeText **tview.TextView, lastResponseTime **time.Time, responsePreviewPanel *tview.TextView, responseHeadersPanel tview.Primitive, responseCookiesPanel *tview.TextView, responseTimelinePanel *tview.TextView, colors *ColorManager) {
	// Update the info bar - replace it in the top row
	newInfoBar, newTimeText, infoBarWidth := createResponseInfoBar(colors, resp, lastTime)
	// The response container has: topRow (item 0), tabPages (item 1)
	// topRow has: tabHeader, spacer, infoBar
	if topRow, ok := response.GetItem(0).(*tview.Flex); ok {
		// Clear and rebuild the top row with the new info bar
		spacer := tview.NewBox().SetBackgroundColor(colors.Background)
		topRow.Clear()
		topRow.AddItem(responseTabHeader, 0, 1, false)
		topRow.AddItem(spacer, 0, 1, false)
		topRow.AddItem(newInfoBar, infoBarWidth, 0, false)
		*responseInfoBar = newInfoBar
		*responseTimeText = newTimeText
		*lastResponseTime = lastTime
	}

	// Update preview tab
	if resp != nil && resp.Body != "" {
		// Pretty-print JSON if it's valid JSON
		bodyToFormat := resp.Body
		if strings.TrimSpace(resp.Body)[0] == '{' || strings.TrimSpace(resp.Body)[0] == '[' {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(resp.Body), &jsonData); err == nil {
				// It's valid JSON, pretty-print it
				prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
				if err == nil {
					bodyToFormat = string(prettyJSON)
				}
			}
		}
		formattedBody := formatBodyContent(bodyToFormat)
		responsePreviewPanel.SetText(formattedBody)
	} else {
		responsePreviewPanel.SetText("(empty response)")
	}

	// Update headers tab
	headersTable, ok := responseHeadersPanel.(*tview.Table)
	if !ok {
		// If it's not a table, skip updating headers
		return
	}
	headersTable.Clear()

	if resp != nil && len(resp.Headers) > 0 {
		// Add header row
		headersTable.SetCell(0, 0,
			tview.NewTableCell("Name").
				SetTextColor(colors.BorderFocus).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
		headersTable.SetCell(0, 1,
			tview.NewTableCell("Value").
				SetTextColor(colors.BorderFocus).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))

		// Add data rows
		row := 1
		for key, values := range resp.Headers {
			for _, value := range values {
				// Truncate long values for display
				truncatedValue := value
				if len(truncatedValue) > 60 {
					truncatedValue = truncatedValue[:57] + "..."
				}

				headersTable.SetCell(row, 0,
					tview.NewTableCell(key).
						SetTextColor(colors.Foreground).
						SetAlign(tview.AlignLeft).
						SetSelectable(true))
				headersTable.SetCell(row, 1,
					tview.NewTableCell(truncatedValue).
						SetTextColor(colors.Foreground).
						SetAlign(tview.AlignLeft).
						SetSelectable(true))
				row++
			}
		}
	} else {
		// No headers - show message
		headersTable.SetCell(0, 0,
			tview.NewTableCell("No response headers").
				SetTextColor(colors.Foreground).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
		headersTable.SetCell(0, 1,
			tview.NewTableCell("").
				SetTextColor(colors.Foreground).
				SetAlign(tview.AlignLeft).
				SetSelectable(false))
	}

	// Update cookies tab
	if resp != nil && len(resp.Cookies) > 0 {
		var cookiesText strings.Builder
		cookiesText.WriteString("Response Cookies:\n\n")
		for _, cookie := range resp.Cookies {
			cookiesText.WriteString(fmt.Sprintf("Name: %s\n", cookie.Name))
			cookiesText.WriteString(fmt.Sprintf("Value: %s\n", cookie.Value))
			if cookie.Domain != "" {
				cookiesText.WriteString(fmt.Sprintf("Domain: %s\n", cookie.Domain))
			}
			if cookie.Path != "" {
				cookiesText.WriteString(fmt.Sprintf("Path: %s\n", cookie.Path))
			}
			if !cookie.Expires.IsZero() {
				cookiesText.WriteString(fmt.Sprintf("Expires: %s\n", cookie.Expires.Format("2006-01-02 15:04:05")))
			}
			cookiesText.WriteString(fmt.Sprintf("Secure: %t\n", cookie.Secure))
			cookiesText.WriteString(fmt.Sprintf("HttpOnly: %t\n\n", cookie.HttpOnly))
		}
		responseCookiesPanel.SetText(cookiesText.String())
	} else {
		responseCookiesPanel.SetText("No response cookies")
	}

	// Update timeline tab
	if resp != nil {
		var timelineText strings.Builder
		timelineText.WriteString("Request Timeline:\n\n")
		timelineText.WriteString(fmt.Sprintf("Request sent: %s\n", resp.Timestamp.Format("2006-01-02 15:04:05")))
		timelineText.WriteString(fmt.Sprintf("Response received: %s\n", resp.Timestamp.Add(resp.Duration).Format("2006-01-02 15:04:05")))
		timelineText.WriteString(fmt.Sprintf("Total duration: %v\n", resp.Duration.Round(time.Millisecond)))
		timelineText.WriteString(fmt.Sprintf("Response size: %d bytes\n", resp.BodySize))
		responseTimelinePanel.SetText(timelineText.String())
	} else {
		responseTimelinePanel.SetText("No request timeline available")
	}
}
