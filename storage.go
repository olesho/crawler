package crawler

import (
	"github.com/jinzhu/gorm"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlStorage struct {
	db *gorm.DB

	list []string
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
		make([]string, 0),
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
	tasks := make([]*Task, 0)
	err := s.db.Where("UserId = ?", uid).Find(&tasks).Error
	return tasks, err
}

func (s *MysqlStorage) AddRecord(r *Record) error {
	s.list = append(s.list, r.Url)
	return s.db.Create(fromRecord(r)).Error
}

func (s *MysqlStorage) ListRecords(taskId int64) ([]*Record, error) {
	tasks := make([]*Record, 0)
	err := s.db.Where("TaskId = ?", taskId).Find(&tasks).Error
	return tasks, err
}

func (s *MysqlStorage) Exists(url string, maxTimeStamp *time.Time) (bool, error) {
	for _, u := range s.list {
		if u == url {
			return true, nil
		}
	}
	return false, nil


	var r GormRecord
	var err error
	if maxTimeStamp != nil {
		err = s.db.First(&r, "createdAt >= ?", time.Time{}).Error
	} else {
		err = s.db.First(&r).Error
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, err
}
