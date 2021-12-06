package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/patrickmn/go-cache"
)

type ListEntry struct {
	Host      string `json:"host"`
	Display   string `json:"display"`
	Flaggable bool   `json:"flaggable"`
}

type FilterConfig struct {
	ContentFilter struct {
		Categories struct {
			Blocked []int `json:"blocked"`
			blocked map[int]struct{}
		} `json:"categories"`
		Lists struct {
			Allowed []*ListEntry `json:"allowed"`
			Blocked []*ListEntry `json:"blocked"`
			allowed map[string]struct{}
			blocked map[string]struct{}
		} `json:"lists"`
	} `json:"content_filter"`
	Videos struct {
		SmartPlay struct {
			AllowedCategoryIDs []int `json:"allowed_category_ids"`
			allowedCategoryIDs map[int]struct{}
		} `json:"smart_play"`
		YouTubeLists struct {
			AllowedChannelIDs []string `json:"allowed_channel_ids"`
			BlockedChannelIDs []string `json:"blocked_channel_ids"`
			AllowedVideoIDs   []string `json:"allowed_video_ids"`
			BlockedVideoIDs   []string `json:"blocked_video_ids"`
			allowedChannelIDs map[string]struct{}
			blockedChannelIDs map[string]struct{}
			allowedVideoIDs   map[string]struct{}
			blockedVideoIDs   map[string]struct{}
		} `json:"youtube_lists"`
	} `json:"videos"`
}

func intSet(ints []int) map[int]struct{} {
	m := make(map[int]struct{})
	for _, i := range ints {
		m[i] = struct{}{}
	}
	return m
}

func stringSet(strings []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, s := range strings {
		m[s] = struct{}{}
	}
	return m
}

func listSet(list []*ListEntry) map[string]struct{} {
	m := make(map[string]struct{})
	for _, e := range list {
		m[e.Host] = struct{}{}
	}
	return m
}

func (c *Config) GetFilterConfig(email string) (*FilterConfig, error) {
	type request struct {
		Platform          string `json:"platform"`
		LastPolicyEpochMS int    `json:"lastPolicyEpochMS"`
		UserEmail         string `json:"userEmail,omitempty"`
	}

	if i, ok := c.Cache.FilterConfigs.Get(email); ok {
		return i.(*FilterConfig), nil
	}

	body := &request{Platform: "cros", LastPolicyEpochMS: 0, UserEmail: email}
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal request body: %w", err)
	}

	r, err := http.NewRequest(http.MethodPost, c.ConfigURL, bytes.NewBuffer(buf))
	if err != nil {
		return nil, fmt.Errorf("Unable to build request: %w", err)
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("X-API-KEY", c.ConfigKey)
	r.Header.Add("CUSTOMERID", c.CustomerID)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("Unable to complete request: %w", err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	f := new(FilterConfig)
	if err = dec.Decode(f); err != nil {
		return nil, fmt.Errorf("Unable to decode response: %w", err)
	}

	f.ContentFilter.Categories.blocked = intSet(f.ContentFilter.Categories.Blocked)
	f.ContentFilter.Lists.blocked = listSet(f.ContentFilter.Lists.Blocked)
	f.ContentFilter.Lists.allowed = listSet(f.ContentFilter.Lists.Allowed)
	f.Videos.SmartPlay.allowedCategoryIDs = intSet(f.Videos.SmartPlay.AllowedCategoryIDs)
	f.Videos.YouTubeLists.allowedChannelIDs = stringSet(f.Videos.YouTubeLists.AllowedChannelIDs)
	f.Videos.YouTubeLists.blockedChannelIDs = stringSet(f.Videos.YouTubeLists.BlockedChannelIDs)
	f.Videos.YouTubeLists.allowedVideoIDs = stringSet(f.Videos.YouTubeLists.AllowedVideoIDs)
	f.Videos.YouTubeLists.blockedVideoIDs = stringSet(f.Videos.YouTubeLists.BlockedVideoIDs)

	c.Cache.FilterConfigs.Set(email, f, cache.DefaultExpiration)
	return f, nil
}

type Host struct {
	Hostname   string
	CategoryID int `json:"cat"`
}

func (c *Config) GetHostInfo(hostname string) (*Host, error) {
	type request struct {
		Action     string `json:"action"`
		CustomerID string `json:"customerId"`
		Hostname   string `json:"host"`
	}

	if i, ok := c.Cache.Hosts.Get(hostname); ok {
		return i.(*Host), nil
	}

	url := fmt.Sprintf("%s?a=%s&customer_id=%s", c.CheckURL, c.CheckKey, c.CustomerID)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to dial websocket: %w", err)
	}
	defer conn.Close()

	body := &request{Action: "dy_lookup", CustomerID: c.CustomerID, Hostname: hostname}
	if err = conn.WriteJSON(body); err != nil {
		return nil, fmt.Errorf("Unable to write request: %w", err)
	}

	host := new(Host)
	if err = conn.ReadJSON(host); err != nil {
		return nil, fmt.Errorf("Unable to read response: %w", err)
	}

	host.Hostname = hostname

	c.Cache.Hosts.Set(hostname, host, cache.DefaultExpiration)
	return host, nil
}

type Video struct {
	ID                   string `json:"video_id"`
	CategoryID           int    `json:"yt_cat"`
	LightspeedCategoryID int    `json:"ls_cat"`
	ChannelID            string `json:"channel_id"`
}

func (c *Config) GetVideoInfo(id string) (*Video, error) {
	type request struct {
		Action     string `json:"action"`
		CustomerID string `json:"customerId"`
		VideoID    string `json:"yt_id"`
	}

	type response struct {
		Video *Video `json:"data"`
	}

	if i, ok := c.Cache.Videos.Get(id); ok {
		return i.(*Video), nil
	}

	url := fmt.Sprintf("%s?a=%s&customer_id=%s", c.CheckURL, c.CheckKey, c.CustomerID)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to dial websocket: %w", err)
	}
	defer conn.Close()

	body := &request{Action: "yt_video", CustomerID: c.CustomerID, VideoID: id}
	if err = conn.WriteJSON(body); err != nil {
		return nil, fmt.Errorf("Unable to write request: %w", err)
	}

	resp := new(response)
	if err = conn.ReadJSON(resp); err != nil {
		return nil, fmt.Errorf("Unable to read response: %w", err)
	}

	c.Cache.Videos.Set(id, resp.Video, cache.DefaultExpiration)
	return resp.Video, nil
}

type Channel struct {
	ID                   string
	LightspeedCategoryID int `json:"ls_cat"`
}

func (c *Config) GetChannelInfo(id string) (*Channel, error) {
	type request struct {
		Action     string `json:"action"`
		CustomerID string `json:"customerId"`
		ChannelID  string `json:"yt_channel"`
	}

	type response struct {
		Channel *Channel `json:"data"`
	}

	if i, ok := c.Cache.Channels.Get(id); ok {
		return i.(*Channel), nil
	}

	url := fmt.Sprintf("%s?a=%s&customer_id=%s", c.CheckURL, c.CheckKey, c.CustomerID)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to dial websocket: %w", err)
	}
	defer conn.Close()

	body := &request{Action: "yt_channel", CustomerID: c.CustomerID, ChannelID: id}
	if err = conn.WriteJSON(body); err != nil {
		return nil, fmt.Errorf("Unable to write request: %w", err)
	}

	resp := new(response)
	if err = conn.ReadJSON(resp); err != nil {
		return nil, fmt.Errorf("Unable to read response: %w", err)
	}

	resp.Channel.ID = id
	c.Cache.Channels.Set(id, resp.Channel, cache.DefaultExpiration)
	return resp.Channel, nil
}
