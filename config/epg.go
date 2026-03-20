package config

type Epg struct {
	Generator       string           `mapstructure:"generator" json:"generator" yaml:"generator"`
	Source          string           `mapstructure:"source" json:"source" yaml:"source"`
	XmlUrl          string           `mapstructure:"xml_url" json:"xml_url" yaml:"xml_url"`
	RtspUrl         string           `mapstructure:"rtsp_url" json:"rtsp_url" yaml:"rtsp_url"`
	RtpUrl          string           `mapstructure:"rtp_url" json:"rtp_url" yaml:"rtp_url"`
	LogoUrl         string           `mapstructure:"logo_url" json:"logo_url" yaml:"logo_url"`
	FetchCron       string           `mapstructure:"fetch_cron" json:"fetch_cron" yaml:"fetch_cron"`
	Playseek        string           `mapstructure:"playseek" json:"playseek" yaml:"playseek"`
	ChannelMappings []ChannelMapping `mapstructure:"channel_mappings" yaml:"channel_mappings"`
}

type ChannelMapping struct {
	Id    string `mapstructure:"id" json:"id" yaml:"id"`
	Igmp  string `mapstructure:"Igmp" json:"igmp" yaml:"igmp"`
	Name  string `mapstructure:"name" json:"name" yaml:"name"`
	Logo  string `mapstructure:"logo" json:"logo" yaml:"logo"`
	Group string `mapstructure:"group" json:"group" yaml:"group"`
}
