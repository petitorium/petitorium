package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
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

		// Sync query params from UI before saving
		currentRequest.QueryParams = getQueryParamsFromUI()

		// Sync body content based on content type
		if currentRequest.ContentType == "Multipart" {
			multipartBody := collectMultipartFieldsFromUI()
			// Only update body if we have multipart fields
			// This preserves JSON when switching to multipart with no fields
			if multipartBody != "" {
				currentRequest.Body = multipartBody
			} else {
				// If multipart is empty, only overwrite if current body doesn't look like JSON
				// This handles case where user clears all multipart fields
				trimmedBody := strings.TrimSpace(currentRequest.Body)
				if !strings.HasPrefix(trimmedBody, "{") && !strings.HasPrefix(trimmedBody, "[") {
					currentRequest.Body = multipartBody // empty
				}
			}
		}

		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error (could show in status or log)
			return
		}
	}
}

func syncCookiesFromUI(workspaceData *workspace.Workspace) {
	if workspaceData == nil {
		return
	}

	workspaceData.CookieJar.Cookies = nil

	for _, row := range currentCookieRows {
		domain := row.DomainInput.GetText()
		name := row.NameInput.GetText()
		value := row.ValueInput.GetText()
		path := row.PathInput.GetText()
		secureStr := row.SecureInput.GetText()
		httpOnlyStr := row.HttpOnlyInput.GetText()
		enabled := row.Checkbox.IsEnabled()

		secure := secureStr == "true"
		httpOnly := httpOnlyStr == "true"

		cookie := workspace.Cookie{
			Name:     name,
			Value:    value,
			Domain:   domain,
			Path:     path,
			Secure:   secure,
			HttpOnly: httpOnly,
			Enabled:  enabled,
		}

		workspaceData.CookieJar.Cookies = append(workspaceData.CookieJar.Cookies, cookie)
	}
}

func SaveCookies(workspaceData *workspace.Workspace) {
	if workspaceData == nil {
		return
	}
	workspace.SaveWorkspace(workspaceData)
}

// syncMethodDropdown syncs the method dropdown with the current request's method
func syncMethodDropdown(currentRequest *workspace.Request, methodDropdown *tview.DropDown, programmaticallyUpdatingMethod *bool) {
	if currentRequest != nil {
		for i, method := range workspace.HTTPMethods {
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

// syncContentTypeDropdown syncs the content type dropdown with the current request's content type
func syncContentTypeDropdown(currentRequest *workspace.Request, contentTypeDropdown *tview.DropDown, programmaticallyUpdatingContentType *bool) {
	if currentRequest != nil {
		contentTypes := []string{"JSON", "Multipart", "No Body"}
		for i, contentType := range contentTypes {
			if contentType == currentRequest.ContentType {
				// Set flag to prevent the SetSelectedFunc from firing
				*programmaticallyUpdatingContentType = true
				contentTypeDropdown.SetCurrentOption(i)
				*programmaticallyUpdatingContentType = false
				break
			}
		}
	}
}

// updateResponseTabs updates the response tabs with new response data
func updateResponseTabs(resp *HTTPResponse, lastTime *time.Time, response *tview.Flex, responseTabHeader *tview.Flex, responseInfoBar **tview.Flex, responseTimeText **tview.TextView, lastResponseTime **time.Time, responsePreviewPanel *tview.TextView, responseHeadersPanel tview.Primitive, responseCookiesPanel *tview.TextView, responseTimelinePanel tview.Primitive, colors *ColorManager, copyCallback func(), saveCallback func()) {
	// Update the info bar - replace it in the top row
	newInfoBar, newTimeText, infoBarWidth := createResponseInfoBar(colors, resp, lastTime, copyCallback, saveCallback)
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
		// headersTable.SetCell(0, 0,
		// 	tview.NewTableCell("Name").
		// 		SetTextColor(colors.BorderFocus).
		// 		SetAlign(tview.AlignLeft).
		// 		SetSelectable(false))
		// headersTable.SetCell(0, 1,
		// 	tview.NewTableCell("Value").
		// 		SetTextColor(colors.BorderFocus).
		// 		SetAlign(tview.AlignLeft).
		// 		SetSelectable(false))

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
						SetTextColor(colors.BorderFocus).
						SetAlign(tview.AlignLeft).
						SetSelectable(true))
				headersTable.SetCell(row, 1,
					tview.NewTableCell(truncatedValue).
						SetTextColor(colors.Success).
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
	timelineTable, ok := responseTimelinePanel.(*tview.Table)
	if !ok {
		return
	}
	timelineTable.Clear()

	if resp != nil {
		// timelineTable.SetCell(0, 0,
		// 	tview.NewTableCell("Field").
		// 		SetTextColor(colors.LabelColor).
		// 		SetAlign(tview.AlignLeft).
		// 		SetSelectable(false))
		// timelineTable.SetCell(0, 1,
		// 	tview.NewTableCell("Value").
		// 		SetTextColor(colors.LabelColor).
		// 		SetAlign(tview.AlignLeft).
		// 		SetSelectable(false))

		timelineData := []struct {
			field string
			value string
		}{
			{"Request sent", resp.Timestamp.Format("2006-01-02 15:04:05")},
			{"Response received", resp.Timestamp.Add(resp.Duration).Format("2006-01-02 15:04:05")},
			{"HTTP duration", resp.Duration.Round(time.Millisecond).String()},
		}
		if config.C.ShowPreparationTime {
			timelineData = append(timelineData, struct {
				field string
				value string
			}{"Total duration (incl. prep)", resp.TotalDuration.Round(time.Millisecond).String()})
		}
		timelineData = append(timelineData, struct {
			field string
			value string
		}{"Response size", fmt.Sprintf("%d bytes", resp.BodySize)})

		for i, item := range timelineData {
			timelineTable.SetCell(i+1, 0,
				tview.NewTableCell(item.field).
					SetTextColor(colors.BorderFocus).
					SetAlign(tview.AlignLeft).
					SetSelectable(true))
			timelineTable.SetCell(i+1, 1,
				tview.NewTableCell(item.value).
					SetTextColor(colors.Success).
					SetAlign(tview.AlignLeft).
					SetSelectable(true))
		}
	} else {
		timelineTable.SetCell(0, 0,
			tview.NewTableCell("No request timeline available").
				SetTextColor(colors.Foreground).
				SetAlign(tview.AlignCenter).
				SetSelectable(false))
	}
}

// getNodeName extracts the name from a tree node's reference
func getNodeName(node *tview.TreeNode) string {
	ref := node.GetReference()
	if col, ok := ref.(workspace.Collection); ok {
		return col.Name
	}
	if req, ok := ref.(workspace.Request); ok {
		return req.Name
	}
	return ""
}

// findNodePath finds the path of names from root to the target node
func findNodePath(root, target *tview.TreeNode) []string {
	if root == target {
		return []string{}
	}

	for _, child := range root.GetChildren() {
		if child == target {
			return []string{getNodeName(child)}
		}

		path := findNodePath(child, target)
		if path != nil {
			return append([]string{getNodeName(child)}, path...)
		}
	}
	return nil
}

// findNodeByPath finds a node in the tree matching the given path of names
func findNodeByPath(root *tview.TreeNode, path []string) *tview.TreeNode {
	if len(path) == 0 {
		return root
	}

	targetName := path[0]
	for _, child := range root.GetChildren() {
		if getNodeName(child) == targetName {
			if len(path) == 1 {
				return child
			}
			result := findNodeByPath(child, path[1:])
			if result != nil {
				return result
			}
		}
	}
	return nil
}

// expandCollectionAndLoadChildren expands a collection node and loads its children if not already done.
// This is useful when navigating to a request in a collapsed collection - we need to ensure
// the collection is expanded and children are loaded before we can find the target request.
func expandCollectionAndLoadChildren(node *tview.TreeNode) {
	ref := node.GetReference()
	col, ok := ref.(workspace.Collection)
	if !ok {
		return
	}

	if !node.IsExpanded() {
		addChildrenToCollectionNode(node, col)
		node.SetExpanded(true)
		node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionExpandedIcon, col.Name))
	}
}

type RequestSearchResult struct {
	Request        *workspace.Request
	CollectionPath []string
}

func searchRequests(query string, collections []workspace.Collection) []RequestSearchResult {
	if query == "" {
		return nil
	}

	var results []RequestSearchResult
	lowerQuery := strings.ToLower(query)

	var traverse func(collections []workspace.Collection, path []string)
	traverse = func(collections []workspace.Collection, path []string) {
		for i := range collections {
			col := &collections[i]
			currentPath := append(path, col.Name)
			for j := range col.Requests {
				req := &col.Requests[j]
				if strings.Contains(strings.ToLower(req.Name), lowerQuery) ||
					strings.Contains(strings.ToLower(req.URL), lowerQuery) {
					resultPath := make([]string, len(currentPath))
					copy(resultPath, currentPath)
					results = append(results, RequestSearchResult{
						Request:        req,
						CollectionPath: resultPath,
					})
				}
			}
			if col.Collections != nil {
				traverse(col.Collections, currentPath)
			}
		}
	}

	traverse(collections, nil)
	return results
}
