package main

import "fmt"

type BlockedError string

func (b BlockedError) Error() string {
	return string(b)
}

func (c *Config) FilterHost(email, hostname string) (int, error) {
	filter, err := c.GetFilterConfig(email)
	if err != nil {
		return 0, fmt.Errorf("Unable to get FilterConfig: %w", err)
	}

	host, err := c.GetHostInfo(hostname)
	if err != nil {
		return 0, fmt.Errorf("Unable to get Host: %w", err)
	}

	var why error

	if _, ok := filter.ContentFilter.Categories.blocked[host.CategoryID]; ok {
		why = BlockedError("blocked category")
	}
	if _, ok := filter.ContentFilter.Lists.blocked[hostname]; ok {
		why = BlockedError("blocked list")
	}
	if _, ok := filter.ContentFilter.Lists.allowed[hostname]; ok {
		why = nil
	}

	return host.CategoryID, why
}

func (c *Config) FilterVideo(email, id string) (int, error) {
	filter, err := c.GetFilterConfig(email)
	if err != nil {
		return 0, fmt.Errorf("Unable to get FilterConfig: %w", err)
	}

	video, err := c.GetVideoInfo(id)
	if err != nil {
		return 0, fmt.Errorf("Unable to get Video: %w", err)
	}

	var why error

	if _, ok := filter.Videos.SmartPlay.allowedCategoryIDs[video.LightspeedCategoryID]; !ok {
		why = BlockedError("blocked category")
	}
	if _, ok := filter.Videos.YouTubeLists.blockedChannelIDs[video.ChannelID]; ok {
		why = BlockedError("blocked channel list")
	}
	if _, ok := filter.Videos.YouTubeLists.allowedChannelIDs[video.ChannelID]; ok {
		why = nil
	}
	if _, ok := filter.Videos.YouTubeLists.blockedVideoIDs[video.ID]; ok {
		why = BlockedError("blocked video list")
	}
	if _, ok := filter.Videos.YouTubeLists.allowedVideoIDs[video.ID]; ok {
		why = nil
	}

	return video.LightspeedCategoryID, why
}

func (c *Config) FilterChannel(email, id string) (int, error) {
	filter, err := c.GetFilterConfig(email)
	if err != nil {
		return 0, fmt.Errorf("Unable to get FilterConfig: %w", err)
	}

	channel, err := c.GetChannelInfo(id)
	if err != nil {
		return 0, fmt.Errorf("Unable to get Channel: %w", err)
	}

	var why error

	if _, ok := filter.Videos.SmartPlay.allowedCategoryIDs[channel.LightspeedCategoryID]; !ok {
		why = BlockedError("blocked category")
	}
	if _, ok := filter.Videos.YouTubeLists.blockedChannelIDs[channel.ID]; ok {
		why = BlockedError("blocked channel list")
	}
	if _, ok := filter.Videos.YouTubeLists.allowedChannelIDs[channel.ID]; ok {
		why = nil
	}

	return channel.LightspeedCategoryID, why
}
