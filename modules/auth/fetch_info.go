package auth

import (
	"encoding/json"
	"fmt"
	"iptv-spider-sh/global"
	"iptv-spider-sh/model"
	"strconv"
	"strings"
	"time"

	"github.com/golang-module/carbon"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

func (c *Client) checkSessionState() error {
	now := carbon.Now()
	if c.AuthInfo.UpdatedAt.Unix() > now.SubHours(1).Timestamp() {
		global.LOG.Info("AuthInfo 更新时间在60分钟以内, 跳过检测")
		return nil
	}
	global.LOG.Info("Check session state")
	p := "service/auth/AuthByAjax.jsp?action=auth"
	uri := fmt.Sprintf("%s/%s", c.EPGHostUrl, p)
	resp := c.httpClient.Request(uri, "GET", nil)
	cont := resp.GetResp().Header().Get("Content-Type")
	if !strings.Contains(cont, "json") {
		global.LOG.Info("Session expired, reAuth")
		// 过期了, 重新认证
		err := c.StartAuth()
		return err
	}
	// 有效期内, 更新 UpdatedAt 为当前时间
	c.AuthInfo.UpdatedAt = now.ToStdTime()
	c.AuthInfo.BimAuthInfo = string(resp.GetRespBytes())
	// 保存到数据库
	global.DB.Updates(&c.AuthInfo)
	global.LOG.Info("Session OK")
	return nil
}

func (c *Client) FetchChannelList() {
	err := c.checkSessionState()
	if err != nil {
		global.LOG.Error("FetchChannelList checkSessionState Err: " + err.Error())
		global.LOG.Error("跳过此次更新")
		return
	}
	global.LOG.Info("开始更新频道信息列表")
	p := "function/ajax/epg7getChannelByAjax.jsp"
	uri := fmt.Sprintf("%s/%s", c.EPGHostUrl, p)
	resp := c.httpClient.Request(uri, "POST", map[string]string{
		"action": "getChannelList",
		"cateID": "000406",
	})
	var respJson model.JsonResponse[model.ChannelInfo]
	err = json.Unmarshal(resp.GetRespBytes(), &respJson)
	if err != nil {
		global.LOG.Error("FetchChannelList Unmarshal Err: " + err.Error())
		return
	}
	if len(respJson.Data) == 0 {
		global.LOG.Error("FetchChannelList Err: No Data!")
		return
	}
	global.LOG.Info(fmt.Sprintf("FetchChannelList Data Length: %d", len(respJson.Data)))
	for i := range respJson.Data {
		// 避免接口返回 0000-00-00
		if respJson.Data[i].LastFetchTime.IsZero() {
			// 使用 carbon.Now() 来代替 time.Now()
			respJson.Data[i].LastFetchTime = carbon.Now().SubHours(5).ToStdTime()
		}

		global.LOG.Info("FetchChannelList Data:",
			zap.Any("Channel Info", respJson.Data[i]))
	}
	/*
		TsTime        int       `gorm:"comment:TimeShiftTime 时移时间" json:"tsTime"`
		Code          string    `gorm:"uniqueIndex;comment:频道代码" json:"code"`
		AuthCode      string    `gorm:"comment:付费认证代码" json:"authCode"`
		Name          string    `gorm:"comment:频道名称" json:"name"`
		ChID          string    `gorm:"uniqueIndex;comment:频道ID" json:"ID"`
		MixNo         string    `gorm:"comment:用户频道映射" json:"mixNo"`
		MediaID       string    `gorm:"comment:未知" json:"mediaID"`
		IsTs          string    `gorm:"comment:是否支持回放" json:"isTs"`
		IsCharge      string    `gorm:"comment:是否需要付费" json:"isCharge"`
		IsHD          bool      `gorm:"default:false;comment:是否是高清频道" json:"-"`
		Is4K          bool      `gorm:"default:false;comment:是否是4K频道" json:"-"`
		IsPullEPG     bool      `gorm:"default:true;comment:是否拉取节目单" json:"-"`
		IsShow        bool      `gorm:"default:true;comment:是否展示该节目" json:"-"`
		CommName      string    `gorm:"comment:通用标题" json:"-"`
		LastFetchTime time.Time `gorm:"comment:节目单最后更新时间" json:"-"`
	*/
	// 数据入库
	result := global.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "mix_no"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"code",
			"auth_code",
			"name",
			"ch_id",
			"is_charge",
			"is_hd",
			"is4_k",
			"comm_name",
		}),
	}).Create(&respJson.Data)

	// 增加判断插入操作是否成功，日志方便排查
	if result.Error != nil {
		global.LOG.Error("数据库插入失败", zap.Error(result.Error))
	} else {
		global.LOG.Info("数据插入成功")
		global.LOG.Info("频道信息列表更新完成")
	}

}

func (c *Client) FetchChannelProg() {
	err := c.checkSessionState()
	if err != nil {
		global.LOG.Error("FetchChannelProg checkSessionState Err: " + err.Error())
		global.LOG.Error("跳过此次更新")
		return
	}
	global.LOG.Info("开始更新节目信息列表; 如果后续没有任何输出, 可能是近期更新过")

	p := "function/ajax/epg7getChannelByAjax.jsp"
	uri := fmt.Sprintf("%s/%s", c.EPGHostUrl, p)

	var channelInfoList []model.ChannelInfo
	// 解决
	global.DB.
		Select("ANY_VALUE(code) as code, ANY_VALUE(ch_id) as ch_id, comm_name, ANY_VALUE(last_fetch_time) as last_fetch_time, ANY_VALUE(is_pull_epg) as is_pull_epg, ANY_VALUE(is_show) as is_show").
		Group("comm_name").
		Find(&channelInfoList)
	now := carbon.Now()
	for _, ch := range channelInfoList {
		// 4 个小时之内更新过，跳过此次更新
		lft := carbon.FromStdTime(ch.LastFetchTime)
		if lft.Gt(now.SubHours(4)) || !ch.IsPullEPG || !ch.IsShow {
			continue
		}
		endTime := now.AddDays(3).TimestampMilli()
		startTime := now.SubDays(7).TimestampMilli()
		params := map[string]string{
			"action":    "getChannelProg",
			"code":      ch.Code,
			"channelID": ch.ChID,
			"endTime":   strconv.FormatInt(endTime, 10),
			"startTime": strconv.FormatInt(startTime, 10),
			"offset":    "0",
			"limit":     "2000",
		}
		resp := c.httpClient.Request(uri, "POST", params)
		var respJson model.JsonResponse[model.EPGDetails]
		err := json.Unmarshal(resp.GetRespBytes(), &respJson)
		if err != nil {
			global.LOG.Error("FetchChannelProg Unmarshal Err: "+err.Error(),
				zap.Any("SessionID", c.JSESSIONID),
				zap.Any("Params", params),
				zap.Any("resp", respJson))
			return
		}
		if len(respJson.Data) == 0 {
			global.LOG.Warn("FetchChannelProg Err: No Data!",
				zap.Any("SessionID", c.JSESSIONID),
				zap.Any("Params", params),
				zap.Any("resp", respJson))
			continue
		}
		// 避免节目时间重合问题，删除所有老数据
		global.DB.Unscoped().Where("comm_name = ?", ch.CommName).Delete(&model.EPGDetails{})
		daysAgo := now.SubDays(7).TimestampMilli()
		length := 0
		var des []model.EPGDetails
		for _, details := range respJson.Data {
			if details.EndTime < daysAgo {
				continue
			}
			details.CommName = ch.CommName
			length++
			des = append(des, details)
		}
		global.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(&des)
		global.LOG.Info("GetChannelProg: ",
			zap.Any("CommName", ch.CommName),
			zap.Any("Data Length", length))
		global.DB.Model(&model.ChannelInfo{}).
			Where("comm_name = ?", ch.CommName).
			Updates(model.ChannelInfo{LastFetchTime: time.Now()})
		time.Sleep(time.Millisecond * 500)
	}
	global.LOG.Info("更新节目信息列表完成")
}
