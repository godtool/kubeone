package system

import (
	"fmt"
	"time"

	v1System "github.com/KubeOperator/kubepi/service/model/v1/system"
	"github.com/KubeOperator/kubepi/service/service/v1/common"
	costomStorm "github.com/KubeOperator/kubepi/pkg/storm"
	"github.com/KubeOperator/kubepi/pkg/util/lang"
	"github.com/asdine/storm/v3/q"
	"github.com/google/uuid"
)

type Service interface {
	common.DBService
	CreateOperationLog(log *v1System.OperationLog, options common.DBOptions)
	CreateLoginLog(log *v1System.LoginLog, options common.DBOptions)
	SearchOperationLogs(num, size int, conditions common.Conditions, options common.DBOptions) ([]v1System.OperationLog, int, error)
	SearchLoginLogs(num, size int, conditions common.Conditions, options common.DBOptions) ([]v1System.LoginLog, int, error)
}

func NewService() Service {
	return &service{}
}

type service struct {
	common.DefaultDBService
}

func (u *service) CreateOperationLog(log *v1System.OperationLog, options common.DBOptions) {
	db := u.GetDB(options)
	log.UUID = uuid.New().String()
	log.CreateAt = time.Now()
	log.UpdateAt = time.Now()
	if err := db.Save(log); err != nil {
		fmt.Printf("operation log %s by user %s write failure, error is %s", log.Operation, log.Operator, err.Error())
	}
}

func (u *service) CreateLoginLog(log *v1System.LoginLog, options common.DBOptions) {
	db := u.GetDB(options)
	log.UUID = uuid.New().String()
	log.CreateAt = time.Now()
	log.UpdateAt = time.Now()
	if err := db.Save(log); err != nil {
		fmt.Printf("login logs by user %s write failure, error is %s", log.UserName, err.Error())
	}
}

func (s *service) SearchOperationLogs(num, size int, conditions common.Conditions, options common.DBOptions) ([]v1System.OperationLog, int, error) {
	db := s.GetDB(options)

	var ms []q.Matcher
	for k := range conditions {
		if conditions[k].Field == "quick" {
			ms = append(ms, q.Or(
				costomStorm.Like("Operator", conditions[k].Value),
				costomStorm.Like("Operation", conditions[k].Value),
				costomStorm.Like("Detail", conditions[k].Value),
			))
		} else {
			field := lang.FirstToUpper(conditions[k].Field)
			value := conditions[k].Value

			switch conditions[k].Operator {
			case "eq":
				ms = append(ms, q.Eq(field, value))
			case "ne":
				ms = append(ms, q.Not(q.Eq(field, value)))
			case "like":
				ms = append(ms, costomStorm.Like(field, value))
			case "not like":
				ms = append(ms, q.Not(costomStorm.Like(field, value)))
			}
		}
	}
	query := db.Select(ms...).OrderBy("CreateAt").Reverse()
	count, err := query.Count(&v1System.OperationLog{})
	if err != nil {
		return nil, 0, err
	}
	if size != 0 {
		query.Limit(size).Skip((num - 1) * size)
	}
	logs := make([]v1System.OperationLog, 0)
	if err := query.Find(&logs); err != nil {
		return nil, 0, err
	}
	return logs, count, nil
}

func (s *service) SearchLoginLogs(num, size int, conditions common.Conditions, options common.DBOptions) ([]v1System.LoginLog, int, error) {
	db := s.GetDB(options)

	var ms []q.Matcher
	for k := range conditions {
		if conditions[k].Field == "quick" {
			ms = append(ms, q.Or(
				costomStorm.Like("UserName", conditions[k].Value),
				costomStorm.Like("Ip", conditions[k].Value),
				costomStorm.Like("City", conditions[k].Value),
			))
		} else {
			field := lang.FirstToUpper(conditions[k].Field)
			value := conditions[k].Value

			switch conditions[k].Operator {
			case "eq":
				ms = append(ms, q.Eq(field, value))
			case "ne":
				ms = append(ms, q.Not(q.Eq(field, value)))
			case "like":
				ms = append(ms, costomStorm.Like(field, value))
			case "not like":
				ms = append(ms, q.Not(costomStorm.Like(field, value)))
			}
		}
	}
	query := db.Select(ms...).OrderBy("CreateAt").Reverse()
	count, err := query.Count(&v1System.LoginLog{})
	if err != nil {
		return nil, 0, err
	}
	if size != 0 {
		query.Limit(size).Skip((num - 1) * size)
	}
	logs := make([]v1System.LoginLog, 0)
	if err := query.Find(&logs); err != nil {
		return nil, 0, err
	}
	return logs, count, nil
}
