package main

import (
	"errors"
	"github.com/imroc/req/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Cookie string

const UA = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/618.1.15.10.15 (KHTML, like Gecko) Mobile/21F90 BiliApp/77900100 os/ios model/iPhone 15 mobi_app/iphone build/77900100 osVer/17.5.1 network/2 channel/AppStore c_locale/zh-Hans_CN s_locale/zh-Hans_CH disable_rcmd/0123"

const InfoUrl = "https://api.bilibili.com/x/activity/bws/online/park/reserve/info?csrf=3f6fe6a573cad708f850e88aa2c37470&reserve_date=20240712,20240713,20240714"
const DoUrl = "https://api.bilibili.com/x/activity/bws/online/park/reserve/do"

var Client = req.C()
var logger *zap.Logger
var nameMap map[int]string
var currentTimeOffset time.Duration
var TicketData = map[string]InfoTicketInfo{}

// TargetPair Reserve ID: ticket ID
var TargetPair = map[int]string{}

func InitLogger() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	l, _ := config.Build()
	logger = l
}

func GetReservationInfo() (*InfoResponse, error) {
	var result InfoResponse
	_, err := Client.R().
		SetHeader("Cookie", Cookie).
		SetHeader("User-Agent", UA).
		SetSuccessResult(&result).Get(InfoUrl)
	if err != nil {
		logger.Error("获取Info接口错误", zap.Error(err))
		return nil, err
	}
	if result.Code != 0 {
		logger.Error("Info 返回不为0", zap.String("message", result.Message))
		return nil, err
	}

	return &result, nil
}

func GetUserTicketInfo(info *InfoData) {
	for _, ticket := range info.UserTicketInfo {
		println("当前可用票", ticket.SkuName, ticket.Ticket, ticket.ScreenName)
		TicketData[ticket.Ticket] = ticket
	}
}

func DoReservation(csrf string, reserveId int, ticketNo string) (*DoResponse, error) {
	var result DoResponse
	body := "csrf=" + csrf + "&inter_reserve_id=" + strconv.Itoa(reserveId) + "&ticket_no=" + ticketNo
	_, err := Client.R().
		SetHeader("Cookie", Cookie).
		SetHeader("User-Agent", UA).
		SetHeader("content-type", "application/x-www-form-urlencoded").
		SetBody(body).Post(DoUrl)

	if err != nil {
		logger.Error("获取Do接口错误", zap.Error(err))
		return nil, err
	}

	return &result, nil
}

func GetCSRFFromCookie(cookie string) string {
	//Split the cookie
	cookieArray := strings.Split(cookie, ";")
	for _, c := range cookieArray {
		if strings.Contains(c, "bili_jct") {
			return strings.Split(c, "=")[1]
		}
	}
	logger.Error("未找到CSRF Token")
	return ""
}

func getReservationStartDate(info InfoData, reserveId int) (int64, error) {
	for _, value := range info.ReserveList {
		for _, v := range value {
			if v.ReserveID == reserveId {
				return v.ReserveBeginTime, nil
			}
		}
	}
	return -1, errors.New("未找到预约信息")
}

func createReservationJob(reserveId int, ticketNo string, csrfToken string, info InfoData, wg *sync.WaitGroup) {
	reservationStartDate, err := getReservationStartDate(info, reserveId)
	if err != nil {
		logger.Error("无法获取预约开始时间", zap.Error(err))
	}

	go doReserve(reservationStartDate, reserveId, ticketNo, csrfToken, wg)

}

func doReserve(startTime int64, reservedId int, ticketId string, csrfToken string, wg *sync.WaitGroup) {
	defer wg.Done()
	//calculate the timer decay
	realStartTime := startTime * 1000
	ticket := TicketData[ticketId]
	for {
		//get start time
		currentTime := time.Now().Add(currentTimeOffset).UnixMilli()
		timeDifference := realStartTime - currentTime
		if timeDifference > 0 {
			// wait for half of the difference
			waitFor := timeDifference / 2
			logger.Info(nameMap[reservedId]+" @ "+ticket.ScreenName+" - 等待预约开始", zap.Time("开始时间", time.Unix(startTime, 0)), zap.Time("当前时间", time.UnixMilli(currentTime)), zap.Duration("时间偏移", currentTimeOffset), zap.Any("等待", time.Duration(timeDifference)*time.Millisecond/2))
			time.Sleep(time.Duration(waitFor) * time.Millisecond)
			continue
		}
		//do reserve
		reservation, err := DoReservation(csrfToken, reservedId, ticketId)
		if err != nil {
			logger.Error(nameMap[reservedId]+" @"+ticket.ScreenName+" - 预约失败，内部错误，重试中。", zap.Error(err))
			continue
		}
		if reservation.Code != 0 {
			logger.Error(nameMap[reservedId]+" @"+ticket.ScreenName+" - 预约失败, 返回不为0，触发重试。", zap.String("message", reservation.Message))
			continue
		}
		logger.Info(nameMap[reservedId]+" @ "+ticket.ScreenName+" - 预约成功", zap.String("message", reservation.Message))
	}
}

func createReservationIDandNameMap(info InfoData) {
	result := make(map[int]string)
	for _, value := range info.ReserveList {
		for _, v := range value {
			result[v.ReserveID] = v.ActTitle
		}
	}
	nameMap = result
}

func syncTimeOffset(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		timeOffset, err := GetNTPOffset()
		if err != nil {
			logger.Error("获取时间失败", zap.Error(err))
		}
		if timeOffset != nil {
			logger.Info("当前时间偏移", zap.Duration("时间偏移", *timeOffset))
			currentTimeOffset = *timeOffset
		} else {
			logger.Warn("未获取到时间偏移")
		}
		time.Sleep(1 * time.Hour)
	}

}
func main() {
	InitLogger()
	logger.Info("程序已启动。")
	var configFile string
	//load args
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	if configFile == "" {
		configFile = "config.json"
	}
	config, err := LoadConfig(configFile)
	if err != nil {
		logger.Error("无法加载配置文件", zap.Error(err))
		return
	}
	Cookie = config.Cookie
	TargetPair = convertJobKeyType(config.Job)
	timeOffset, err := GetNTPOffset()
	if err != nil {
		logger.Error("获取时间失败", zap.Error(err))
	}
	if timeOffset != nil {
		logger.Info("当前时间偏移", zap.Duration("时间偏移", *timeOffset))
		currentTimeOffset = *timeOffset
	} else {
		logger.Warn("未获取到时间偏移")
	}

	// 获取预约信息
	info, err := GetReservationInfo()
	if err != nil {
		logger.Error("获取预约信息失败", zap.Error(err))
		return
	}
	// 获取用户可用票
	GetUserTicketInfo(&info.Data)
	createReservationIDandNameMap(info.Data)
	csrfToken := GetCSRFFromCookie(Cookie)
	if csrfToken == "" {
		logger.Error("获取CSRF Token失败")
		return
	}
	logger.Info("CSRF Token", zap.String("token", csrfToken))

	var wg sync.WaitGroup
	//set up time sync
	wg.Add(1)
	go syncTimeOffset(&wg)

	// 预约
	for reserveId, ticketNo := range TargetPair {
		wg.Add(1)
		createReservationJob(reserveId, ticketNo, csrfToken, info.Data, &wg)
	}
	wg.Wait()
}
