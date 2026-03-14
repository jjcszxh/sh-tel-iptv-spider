package api

import (
	"iptv-spider-sh/global"
	"iptv-spider-sh/modules/auth"
	"iptv-spider-sh/utils"
	"time"

	"github.com/kataras/iris/v12"
)

func InitApiRouters(rg iris.Party) {
	rg.Get("/schedule", schedule)

	rg.Get("/run", func(ctx iris.Context) {
		taskName := ctx.FormValue("task")

		go func() {
			global.ConcurrencyControl.Do("", func() (interface{}, error) {
				switch taskName {
				case "clean-ch":
					auth.CleanChannelData()
				case "clean-chi":
					auth.CleanChannelInfoData()
				case "clean-epg":
					auth.CleanEPGDetailsData()
				case "clean":
					auth.CleanChannelData()
					auth.CleanChannelInfoData()
					auth.CleanEPGDetailsData()
				case "update-chi":
					auth.GetGlobalClient().FetchChannelList()
				case "update-epg":
					auth.GetGlobalClient().FetchChannelProg()
				case "upload-m3u":
					auth.GenerateAndUploadM3u()
				case "upload-xmltv":
					auth.GenerateAndUploadXmlTv()
				case "upload-xmltv7":
					auth.GenerateAndUploadXmlTvDays7()
				}
				return nil, nil
			})
		}()
		ctx.WriteString("OK")
	})

	rg.Get("/m3u8", generateM3u8)

	rg.Get("/tsM3u8", generateTsM3u8)

	rg.Get("/epg", generateXmlTv)

}

func schedule(ctx iris.Context) {
	type s struct {
		ID       int
		PreTime  time.Time
		NextTime time.Time
	}
	var schedule []s
	for _, entry := range global.CRON.Entries() {
		schedule = append(schedule, s{
			ID:       int(entry.ID),
			PreTime:  entry.Prev,
			NextTime: entry.Next,
		})
	}
	ctx.JSON(schedule)
}

// 生成m3u8文件 节目去重
func generateM3u8(ctx iris.Context) {
	// 获取query参数
	udpxy := ctx.FormValue("udpxy")
	scheme := ctx.FormValue("scheme")
	xteve := ctx.FormValue("xteve")
	all := ctx.FormValue("all")
	ref := ctx.FormValue("ref")

	var bufStr string
	if xteve == "true" {
		bufStr = "xteve"
	} else if udpxy != "" {
		bufStr = udpxy
	} else if scheme != "" {
		bufStr = scheme
	}
	if all == "true" {
		bufStr += all
	}
	reqMD5Key := utils.CalcMD5KeyForRequest("generateM3u8", bufStr)
	// 缓存机制
	if ref != "true" && global.CACHE.IsExist(reqMD5Key) {
		ctx.Header("Content-Disposition", "attachment; filename=iptv.m3u")
		ctx.Binary(global.CACHE.Get(reqMD5Key).([]byte))
		return
	}
	// 并发时合并请求
	resp, _, _ := global.ConcurrencyControl.Do(reqMD5Key, func() (interface{}, error) {
		respBytes := auth.GenerateM3u8(udpxy, scheme, xteve, all)
		timeOut := time.Duration(global.CONFIG.Cache.DefTimeOut)
		global.CACHE.Put(reqMD5Key, respBytes, time.Minute*timeOut)
		return respBytes, nil
	})
	ctx.Header("Content-Disposition", "attachment; filename=iptv.m3u")
	ctx.Binary(resp.([]byte))
}

func generateTsM3u8(ctx iris.Context) {
	ref := ctx.FormValue("ref")
	reqMD5Key := utils.CalcMD5KeyForRequest("generateTsM3u8")
	// 缓存机制
	if ref != "true" && global.CACHE.IsExist(reqMD5Key) {
		ctx.Header("Content-Disposition", "attachment; filename=iptv-ts.m3u")
		ctx.Binary(global.CACHE.Get(reqMD5Key).([]byte))
		return
	}
	// 并发时合并请求
	resp, _, _ := global.ConcurrencyControl.Do(reqMD5Key, func() (interface{}, error) {
		respBytes := auth.GenerateTimeShiftM3u8()
		timeOut := time.Duration(global.CONFIG.Cache.DefTimeOut)
		global.CACHE.Put(reqMD5Key, respBytes, time.Minute*timeOut)
		return respBytes, nil
	})
	ctx.Header("Content-Disposition", "attachment; filename=iptv-ts.m3u")
	ctx.Binary(resp.([]byte))
}
