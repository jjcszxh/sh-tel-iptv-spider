package auth

import (
	"encoding/xml"
	"errors"
	"fmt"
	"iptv-spider-sh/global"
	"iptv-spider-sh/model"
	"iptv-spider-sh/modules/m3u"
	"iptv-spider-sh/utils"
	"net/url"
	"strings"

	"github.com/golang-module/carbon"
)

const timeFormat = carbon.ShortDateTimeLayout + " -0700"

func GenerateM3u8(udpxy, scheme, xteve, all string) []byte {
	m3uWriter := m3u.NewWriter()
	m3uWriter.WriteHeaderWithInfo(global.CONFIG.Epg.XmlUrl)

	// 查询数据库
	var channelInfoList []model.ChannelInfo
	global.DB.Order("mix_no asc").
		Find(&channelInfoList)
	// 去重
	newChanInfo := model.RemoveDuplicateChannelInfo(channelInfoList)
	for _, info := range newChanInfo {
		// 不展示
		if !info.IsShow {
			continue
		}
		channel := model.Channel{}
		global.DB.Where("user_channel_id = ?", info.MixNo).
			Find(&channel)

		m3u8Mapping := model.M3u8Mapping{}
		global.DB.Where("comm_name = ?", info.CommName).
			Find(&m3u8Mapping)

		// 如果没有 logo，就使用本地 logo
		if m3u8Mapping.Logo == "" {
			m3u8Mapping.Logo = fmt.Sprintf("http://127.0.0.1:8888/api/logo/%s.png", info.CommName)
		}

		if all != "true" && (m3u8Mapping.AutoGroups == "购物" ||
			m3u8Mapping.CustomGroups == "购物") {
			continue
		}
		//uri := assemblyUrl(udpxy, scheme, xteve, channel.ChannelURL)//修改
		uri := assemblyUrl(
			udpxy,
			scheme,
			xteve,
			channel.ChannelURL,
			channel.ChannelFCCIP,
			channel.ChannelFCCPort,
		)

		catchupSource := ""
		if channel.TimeShiftURL != "" {
			// 去掉数据库里的 rtsp:// 前缀
			trimmed := strings.TrimPrefix(channel.TimeShiftURL, "rtsp://")
			// 拼接固定前缀
			catchupSource = fmt.Sprintf("%s%s&playseek={utc:YmdHMS}-{utcend:YmdHMS}", global.CONFIG.Epg.RtspUrl, trimmed)
		}

		// 使用新的方法写入EXTINF
		m3uWriter.WriteWithCatchup(uri, catchupSource, info, m3u8Mapping)
	}
	return m3uWriter.Bytes()
}

func GenerateTimeShiftM3u8() []byte {
	m3uWriter := m3u.NewWriter()
	m3uWriter.WriteHeaderWithInfo(global.CONFIG.Epg.XmlUrl)
	// 查询数据库
	var channelInfoList []model.ChannelInfo
	global.DB.Find(&channelInfoList)
	// 去重
	newChanInfo := model.RemoveDuplicateChannelInfo(channelInfoList)
	for _, info := range newChanInfo {
		// 不展示
		if !info.IsShow {
			continue
		}
		channel := model.Channel{}
		global.DB.Where("user_channel_id = ?", info.MixNo).
			Find(&channel)

		m3u8Mapping := model.M3u8Mapping{}
		global.DB.Where("comm_name = ?", info.CommName).
			Find(&m3u8Mapping)

		if m3u8Mapping.AutoGroups == "购物" || m3u8Mapping.CustomGroups == "购物" {
			continue
		}
		uri := assemblyUrl("", "", "", channel.TimeShiftURL, "", "") //修改加上fcc端口和用户
		m3uWriter.Write(uri, info, m3u8Mapping)
	}
	return m3uWriter.Bytes()
}

// func assemblyUrl(udpxy, scheme, xteve, uri string) string //修改
func assemblyUrl(udpxy, scheme, xteve, uri, fccIp, fccPort string) string {
	u, _ := url.Parse(uri)

	// xteve模式
	if xteve == "true" {
		return fmt.Sprintf("udp://@%s", u.Host)
	}

	// udpxy模式
	if udpxy != "" {
		return fmt.Sprintf("http://%s/udp/%s", udpxy, u.Host)
	}

	// HTTP RTP + FCC 使用动态加载的 rtp_url
	if fccIp != "" && fccPort != "" {
		return fmt.Sprintf(
			"%s%s?fcc=%s:%s",
			global.CONFIG.Epg.RtpUrl, // 使用动态加载的 rtp_url
			u.Host,
			fccIp,
			fccPort,
		)
	}

	// HTTP RTP 无FCC 使用动态加载的 rtp_url
	return fmt.Sprintf(
		"%s%s",
		global.CONFIG.Epg.RtpUrl, // 使用动态加载的 rtp_url
		u.Host,
	)
}

func GenerateXmlTv(daysAgo int) ([]byte, error) {
	if daysAgo < 1 {
		daysAgo = 1
	} else if daysAgo > 7 {
		daysAgo = 7
	}
	var now = carbon.Now()
	var xmlTv = model.XmlTV{
		Generator: fmt.Sprintf("%s %s", global.CONFIG.Epg.Generator, now.ToDateTimeString()),
		Source:    global.CONFIG.Epg.Source,
	}
	// 取数据
	var channelInfoList []model.ChannelInfo
	global.DB.Find(&channelInfoList)
	// 去重
	newChanInfo := model.RemoveDuplicateChannelInfo(channelInfoList)
	for _, info := range newChanInfo {
		// 不展示
		if !info.IsShow {
			continue
		}
		chId := info.MixNo
		xmlTv.Channel = append(xmlTv.Channel, &model.XmlTvChannel{
			ID:          chId,
			DisplayName: []model.DisplayName{{Lang: "zh", Value: info.CommName}},
		})
		if !info.IsPullEPG {
			xmlTv.Program = append(xmlTv.Program, &model.Program{
				Channel: chId,
				Title:   []*model.Title{{Lang: "zh"}},
				Desc:    []*model.Desc{{Lang: "zh"}},
			})
			continue
		}

		var epgData []model.EPGDetails
		global.DB.Where("comm_name = ?", info.CommName).
			Where("end_time > ?", now.SubDays(daysAgo).TimestampMilli()).
			Order("start_time asc").
			Find(&epgData)

		for _, epg := range epgData {
			startTime := carbon.CreateFromTimestampMilli(epg.StartTime).Layout(timeFormat)
			endTime := carbon.CreateFromTimestampMilli(epg.EndTime).Layout(timeFormat)
			xmlTv.Program = append(xmlTv.Program, &model.Program{
				Channel: chId,
				Start:   startTime,
				Stop:    endTime,
				Title:   []*model.Title{{Lang: "zh", Value: epg.Name}},
				Desc:    []*model.Desc{{Lang: "zh"}},
			})
		}
	}
	// 序列化
	epgBytes, err := xml.MarshalIndent(&xmlTv, "", "  ")
	if err != nil {
		global.LOG.Error("节目表单生成出错: " + err.Error())
		return nil, errors.New("节目表单生成出错")
	}
	epgBytes = append([]byte(model.PrefixHeader+"\n"), epgBytes...)
	return epgBytes, nil
}

func GenerateAndUploadM3u() {
	m3uBytes := GenerateM3u8("", "", "true", "")
	utils.UploadToOSS("/sh/tel-xteve.m3u", m3uBytes)
}

func GenerateAndUploadXmlTv() {
	xmlTvBytes, _ := GenerateXmlTv(1)
	utils.UploadToOSS("/sh/tel-epg.xml", xmlTvBytes)
}

func GenerateAndUploadXmlTvDays7() {
	xmlTvBytes, _ := GenerateXmlTv(7)
	utils.UploadToOSS("/sh/tel-epg-7.xml", xmlTvBytes)
}
