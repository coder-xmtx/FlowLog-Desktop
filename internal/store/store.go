package store

import (
	"encoding/json"
	"os"
	"path/filepath"

	"flowlog-desktop/internal/model"
)

/*
结构体 - 存储

记录：数据目录
*/
type Store struct {
	dir string
}

/*
Store - filePath

参数：日期文件名

功能：设置 JSON 日期文件保存的文件名
*/
func (s *Store) filePath(date string) string {
	return filepath.Join(s.dir, date+".json")
}

/*
Store - LoadDay

参数：日期文件名

功能：读取某一天的JSON数据，写回 DayData 结构体
*/
func (s *Store) LoadDay(date string) (model.DayData, error) {
	// 读取 JSON 数据
	bytes, err := os.ReadFile(s.filePath(date))
	if os.IsNotExist(err) {
		return model.DayData{Date: date, Todos: []model.Todo{}}, nil
	}
	if err != nil {
		return model.DayData{}, err
	}

	// 将 JSON 数据写进 DayData 结构体
	var data model.DayData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return model.DayData{}, err
	}
	if data.Todos == nil {
		data.Todos = []model.Todo{}
	}
	return data, nil
}

/*
Store - SaveDay

参数：DayData 结构体

功能：把某一天的数据写入对应日期的 JSON 文件
*/
func (s *Store) SaveDay(data model.DayData) error {
	// 读取 DayData 结构体数据
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	// 写入 JSON 文件
	return os.WriteFile(s.filePath(data.Date), bytes, 0644)
}

/*
New

参数：想要创建的目录名

功能：创建 Store，并确保数据目录存在
*/
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}
