// 存放数据结构

package model

/*
结构体 - 每日总数据

记录：日期、日记、待办
*/
type DayData struct {
	Date  string `json:"date"`
	Diary string `json:"diary"`
	Todos []Todo `json:"todos"`
}

/*
结构体 - 单条 todo 数据

记录：ID、文本、是否完成
*/
type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}
