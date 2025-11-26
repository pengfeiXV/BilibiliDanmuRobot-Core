package model

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type (
	DanmuModel interface {
		Insert(ctx context.Context, tx *gorm.DB, data *DanmuBase) error
		FindOne(ctx context.Context, id int64, date string) (*DanmuBase, error)
		UpdateCount(ctx context.Context, uid int64) error
		GetRecent3DayRecords(ctx context.Context, uid int64) ([]DanmuBase, error)
		GetDateStr(daysFromToday int) string
	}
	defaultDanmuModel struct {
		conn  *gorm.DB
		table string
	}
	DanmuBase struct {
		ID         int64 `gorm:"primaryKey;autoIncrement"`
		RoomID     int64
		Uid        int64
		Username   string
		CreateTime time.Time
		Content    string
	}
)

func NewDanmuModel(conn *gorm.DB, RoomID int64) DanmuModel {
	err := conn.Table(fmt.Sprintf("danmu_%v", RoomID)).AutoMigrate(&DanmuBase{})
	if err != nil {
		logx.Error(err)
	}
	return &defaultDanmuModel{
		conn:  conn,
		table: fmt.Sprintf("danmu_%v", RoomID),
	}
}

func (m *defaultDanmuModel) GetDateStr(daysFromToday int) string {
	dateStr := time.Now().AddDate(0, 0, -daysFromToday).Format("2006-01-02")
	return dateStr
}

func (m *defaultDanmuModel) Insert(ctx context.Context, tx *gorm.DB, data *DanmuBase) error {
	db := m.conn
	if tx != nil {
		db = tx
	}
	err := db.WithContext(ctx).Table(m.table).Save(&data).Error
	return err
}

func (m *defaultDanmuModel) FindOne(ctx context.Context, uid int64, date string) (*DanmuBase, error) {
	var resp DanmuBase
	err := m.conn.WithContext(ctx).Table(m.table).Model(&DanmuBase{}).Where("uid = ? AND date = ?", uid, date).Take(&resp).Error
	switch err {
	case nil:
		return &resp, nil
	case ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultDanmuModel) GetRecent3DayRecords(ctx context.Context, uid int64) ([]DanmuBase, error) {
	var resp []DanmuBase
	endDate := m.GetDateStr(0)
	startDate := m.GetDateStr(2)
	err := m.conn.WithContext(ctx).Table(m.table).Model(&DanmuBase{}).Where("uid = ? AND date BETWEEN ? AND ?", uid, startDate, endDate).Order("date asc").Take(&resp).Error
	switch err {
	case nil:
		return resp, nil
	case ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultDanmuModel) UpdateCount(ctx context.Context, uid int64) error {
	today := m.GetDateStr(0)
	err := m.conn.WithContext(ctx).Table(m.table).Model(&SingInBase{}).Where("uid = ? AND date = ?", uid, today).UpdateColumn("count", gorm.Expr("count + ?", 1)).Error
	return err
}
