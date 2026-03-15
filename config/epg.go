package config

type Epg struct {
	Generator string `mapstructure:"generator" json:"generator" yaml:"generator"`
	Source    string `mapstructure:"source" json:"source" yaml:"source"`
	XmlUrl    string `mapstructure:"xml_url" json:"xml_url" yaml:"xml_url"`
	RtspUrl   string `mapstructure:"rtsp_url" json:"rtsp_url" yaml:"rtsp_url"`
	RtpUrl    string `mapstructure:"rtp_url" json:"rtp_url" yaml:"rtp_url"`
	FetchCron string `mapstructure:"fetch_cron" json:"fetch_cron" yaml:"fetch_cron"`
}
