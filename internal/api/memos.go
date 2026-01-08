package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

func (c *Client) ListMemos(pageSize int, pageToken string) (*ListMemosResponse, error) {
	path := "/memos"

	params := url.Values{}
	if pageSize > 0 {
		params.Set("pageSize", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		params.Set("pageToken", pageToken)
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response ListMemosResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

func (c *Client) GetMemo(name string) (*Memo, error) {
	path := "/" + name

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var memo Memo
	if err := json.Unmarshal(respBody, &memo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &memo, nil
}

func (c *Client) CreateMemo(content string) (*Memo, error) {
	reqBody := CreateMemoRequest{
		Content: content,
	}

	respBody, err := c.doRequest("POST", "/memos", reqBody)
	if err != nil {
		return nil, err
	}

	var memo Memo
	if err := json.Unmarshal(respBody, &memo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &memo, nil
}

func (c *Client) UpdateMemo(name string, content string) (*Memo, error) {
	path := "/" + name + "?updateMask=content"

	reqBody := UpdateMemoRequest{
		Content: content,
	}

	respBody, err := c.doRequest("PATCH", path, reqBody)
	if err != nil {
		return nil, err
	}

	var memo Memo
	if err := json.Unmarshal(respBody, &memo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &memo, nil
}

func (c *Client) DeleteMemo(name string) error {
	path := "/" + name

	_, err := c.doRequest("DELETE", path, nil)
	return err
}

func (c *Client) SetMemoPinned(name string, pinned bool) (*Memo, error) {
	path := "/" + name + "?updateMask=pinned"

	reqBody := UpdateMemoRequest{
		Pinned: &pinned,
	}

	respBody, err := c.doRequest("PATCH", path, reqBody)
	if err != nil {
		return nil, err
	}

	var memo Memo
	if err := json.Unmarshal(respBody, &memo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &memo, nil
}

func (c *Client) ArchiveMemo(name string) (*Memo, error) {
	path := "/" + name + "?updateMask=rowStatus"

	reqBody := UpdateMemoRequest{
		RowStatus: RowStatusArchived,
	}

	respBody, err := c.doRequest("PATCH", path, reqBody)
	if err != nil {
		return nil, err
	}

	var memo Memo
	if err := json.Unmarshal(respBody, &memo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &memo, nil
}

func (c *Client) UnarchiveMemo(name string) (*Memo, error) {
	path := "/" + name + "?updateMask=rowStatus"

	reqBody := UpdateMemoRequest{
		RowStatus: RowStatusNormal,
	}

	respBody, err := c.doRequest("PATCH", path, reqBody)
	if err != nil {
		return nil, err
	}

	var memo Memo
	if err := json.Unmarshal(respBody, &memo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &memo, nil
}

func (c *Client) SetMemoVisibility(name string, visibility string) (*Memo, error) {
	path := "/" + name + "?updateMask=visibility"

	reqBody := UpdateMemoRequest{
		Visibility: visibility,
	}

	respBody, err := c.doRequest("PATCH", path, reqBody)
	if err != nil {
		return nil, err
	}

	var memo Memo
	if err := json.Unmarshal(respBody, &memo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &memo, nil
}
