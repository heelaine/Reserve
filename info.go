package main

type InfoResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Ttl     int      `json:"ttl"`
	Data    InfoData `json:"data"`
}

type InfoData struct {
	UserReserveInfo map[string]InfoReserveInfo     `json:"user_reserve_info"`
	UserTicketInfo  map[string]InfoTicketInfo      `json:"user_ticket_info"`
	ReserveList     map[string][]InfoReserveDetail `json:"reserve_list"`
}

type InfoReserveInfo struct {
	TotalCount int `json:"total_count"`
	CurCount   int `json:"cur_count"`
}

type InfoTicketInfo struct {
	Sid        int    `json:"sid"`
	SkuName    string `json:"sku_name"`
	ScreenName string `json:"screen_name"`
	Type       int    `json:"type"`
	Ticket     string `json:"ticket"`
}

type InfoReserveDetail struct {
	ReserveID         int             `json:"reserve_id"`
	ActType           string          `json:"act_type"`
	ActTitle          string          `json:"act_title"`
	ActImg            string          `json:"act_img"`
	ActBeginTime      int64           `json:"act_begin_time"`
	ActEndTime        int64           `json:"act_end_time"`
	ReserveBeginTime  int64           `json:"reserve_begin_time"`
	ReserveEndTime    int64           `json:"reserve_end_time"`
	DescribeInfo      string          `json:"describe_info"`
	VipTicketNum      int             `json:"vip_ticket_num"`
	StandardTicketNum int             `json:"standard_ticket_num"`
	ScreenDate        int             `json:"screen_date"`
	IsVipTicket       int             `json:"is_vip_ticket"`
	State             int             `json:"state"`
	OnlineState       int             `json:"online_state"`
	DisplayIndex      int             `json:"display_index"`
	VipStock          int             `json:"vip_stock"`
	StandardStock     int             `json:"standard_stock"`
	NextReserve       InfoNextReserve `json:"next_reserve"`
}

type InfoNextReserve struct {
	ReserveBeginTime int64 `json:"reserve_begin_time"`
	ReserveEndTime   int64 `json:"reserve_end_time"`
	IsVipTicket      int   `json:"is_vip_ticket"`
}
