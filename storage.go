package crawler

import (
	"github.com/jinzhu/gorm"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlStorage struct {
	db *gorm.DB
}

type MysqlConfig struct {
	User string
	Password string
	Host string
	Port string
	DBName string
}

type GormTask struct {
	*gorm.Model
	Task
}

type GormRecord struct {
	*gorm.Model
	Record
}

func fromTask(t *Task) *GormTask {
	return &GormTask{Task: *t}
}

func fromRecord(r *Record) *GormRecord {
	return &GormRecord{Record: *r}
}

func NewMysqlStorage(conf *MysqlConfig) (*MysqlStorage, error) {
	db, err := gorm.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8&parseTime=True&loc=Local", conf.User, conf.Password, conf.Host, conf.Port, conf.DBName))
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&GormTask{})
	db.AutoMigrate(&GormRecord{})
	return &MysqlStorage{
		db,
	}, nil
}

func (s *MysqlStorage) GetTask(taskId int64) (*Task, error) {
	var t GormTask
	err := s.db.First(&t, "id = ?", taskId).Error
	if err != nil {
		return nil, err
	}
	return &t.Task, nil
}


func (s *MysqlStorage) AddTask(t *Task) (int64, error) {
	gormT := fromTask(t)
	err := s.db.Create(&gormT).Error
	return int64(gormT.ID), err
}

func (s *MysqlStorage) RemoveTask(id int64) error {
	return s.db.Delete("id = ?", id).Error
}

func (s *MysqlStorage) ListTasks(uid int64) ([]*Task, error) {
	tasks := make([]*GormTask, 0)
	err := s.db.Where("user_id = ?", uid).Find(&tasks).Error
	result := make([]*Task, len(tasks))
	for i, _ := range tasks {
		result[i] = &tasks[i].Task
		result[i].id = int64(tasks[i].ID)
	}
 	return result, err
}

func (s *MysqlStorage) AddRecord(r *Record) error {
	return s.db.Create(fromRecord(r)).Error
}

func (s *MysqlStorage) SetRecordChecked(url string) error {
	return s.db.Model(&GormRecord{}).Where("url = ?", url).Update("checked", true).Error
}

func (s *MysqlStorage) ListUncheckedRecords(taskId int64) ([]*Record, error) {
	tasks := make([]*GormRecord, 0)
	err := s.db.Where("task_id = ?", taskId).Find(&tasks, "checked = ?", false).Error
	results := make([]*Record, len(tasks))
	for i, _ := range tasks {
		results[i] = &tasks[i].Record
	}
	return results, err
}

func (s *MysqlStorage) Exists(url string, maxTimeStamp *time.Time) (bool, error) {
	var r GormRecord
	var err error
	if maxTimeStamp != nil {
		err = s.db.First(&r, "createdAt >= ? AND url = ? AND checked = ?", time.Time{}, url, false).Error
	} else {
		err = s.db.First(&r, "url = ? AND checked = ?", url, false).Error
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, err
}
