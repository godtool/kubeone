package ldap

import (
	"encoding/json"
	"errors"
	"fmt"
	v1 "github.com/KubeOperator/kubepi/service/model/v1"
	v1Ldap "github.com/KubeOperator/kubepi/service/model/v1/ldap"
	v1Role "github.com/KubeOperator/kubepi/service/model/v1/role"
	v1User "github.com/KubeOperator/kubepi/service/model/v1/user"
	"github.com/KubeOperator/kubepi/service/server"
	"github.com/KubeOperator/kubepi/service/service/v1/common"
	"github.com/KubeOperator/kubepi/service/service/v1/rolebinding"
	"github.com/KubeOperator/kubepi/service/service/v1/user"
	ldapClient "github.com/KubeOperator/kubepi/pkg/util/ldap"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/google/uuid"
	"reflect"
	"strings"
	"time"
)

type Service interface {
	common.DBService
	Create(ldap *v1Ldap.Ldap, options common.DBOptions) error
	List(options common.DBOptions) ([]v1Ldap.Ldap, error)
	Update(id string, ldap *v1Ldap.Ldap, options common.DBOptions) error
	GetById(id string, options common.DBOptions) (*v1Ldap.Ldap, error)
	Delete(id string, options common.DBOptions) error
	Sync(id string, options common.DBOptions) error
	Login(user v1User.User, password string, options common.DBOptions) error
	TestConnect(ldap *v1Ldap.Ldap) (int, error)
	TestLogin(username string, password string) error
	ImportUsers(users []v1User.ImportUser) (v1User.ImportResult, error)
	CheckStatus() bool
	GetLdapUser() ([]v1User.ImportUser, error)
}

func NewService() Service {
	return &service{
		userService:        user.NewService(),
		roleBindingService: rolebinding.NewService(),
	}
}

type service struct {
	common.DefaultDBService
	userService        user.Service
	roleBindingService rolebinding.Service
}

func (l *service) Create(ldap *v1Ldap.Ldap, options common.DBOptions) error {
	m := make(map[string]string)
	err := json.Unmarshal([]byte(ldap.Mapping), &m)
	if err != nil {
		return err
	}
	lc := ldapClient.NewLdapClient(ldap.Address, ldap.Port, ldap.Username, ldap.Password, ldap.TLS)
	err = lc.Connect()
	if err != nil {
		return err
	}
	db := l.GetDB(options)
	ldap.UUID = uuid.New().String()
	ldap.CreateAt = time.Now()
	ldap.UpdateAt = time.Now()
	return db.Save(ldap)
}

func (l *service) List(options common.DBOptions) ([]v1Ldap.Ldap, error) {
	db := l.GetDB(options)
	ldap := make([]v1Ldap.Ldap, 0)
	if err := db.All(&ldap); err != nil {
		return nil, err
	}
	return ldap, nil
}

func (l *service) Update(id string, ldap *v1Ldap.Ldap, options common.DBOptions) error {
	m := make(map[string]string)
	err := json.Unmarshal([]byte(ldap.Mapping), &m)
	if err != nil {
		return err
	}
	lc := ldapClient.NewLdapClient(ldap.Address, ldap.Port, ldap.Username, ldap.Password, ldap.TLS)
	if err := lc.Connect(); err != nil {
		return err
	}
	old, err := l.GetById(id, options)
	if err != nil {
		return err
	}
	ldap.UUID = old.UUID
	ldap.CreateAt = old.CreateAt
	ldap.UpdateAt = time.Now()
	db := l.GetDB(options)
	if ldap.Enable != old.Enable {
		err = db.UpdateField(ldap, "Enable", ldap.Enable)
		if err != nil {
			return err
		}
	}
	if ldap.TLS != old.TLS {
		err = db.UpdateField(ldap, "TLS", ldap.TLS)
		if err != nil {
			return err
		}
	}
	return db.Update(ldap)
}

func (l *service) GetById(id string, options common.DBOptions) (*v1Ldap.Ldap, error) {
	db := l.GetDB(options)
	var ldap v1Ldap.Ldap
	query := db.Select(q.Eq("UUID", id))
	if err := query.First(&ldap); err != nil {
		return nil, err
	}
	return &ldap, nil
}

func (l *service) Delete(id string, options common.DBOptions) error {
	db := l.GetDB(options)
	ldap, err := l.GetById(id, options)
	if err != nil {
		return err
	}
	return db.DeleteStruct(ldap)
}

func (l *service) GetLdapUser() ([]v1User.ImportUser, error) {
	users := []v1User.ImportUser{}
	ldaps, err := l.List(common.DBOptions{})
	if err != nil {
		return users, err
	}
	if len(ldaps) == 0 {
		return users, errors.New("请先保存LDAP配置")
	}
	ldap := ldaps[0]
	if !ldap.Enable {
		return users, errors.New("请先启用LDAP")
	}
	lc := ldapClient.NewLdapClient(ldap.Address, ldap.Port, ldap.Username, ldap.Password, ldap.TLS)
	if err := lc.Connect(); err != nil {
		return users, err
	}
	attributes, err := ldap.GetAttributes()
	if err != nil {
		return users, err
	}
	mappings, err := ldap.GetMappings()
	if err != nil {
		return users, err
	}
	entries, err := lc.Search(ldap.Dn, ldap.Filter, ldap.SizeLimit, ldap.TimeLimit, attributes)
	if err != nil {
		return users, err
	}
	if len(entries) == 0 {
		return users, nil
	}
	for _, entry := range entries {
		us := new(v1User.ImportUser)
		us.Available = true
		rv := reflect.ValueOf(&us).Elem().Elem()
		for _, at := range entry.Attributes {
			for k, v := range mappings {
				if v == at.Name && len(at.Values) > 0 {
					fv := rv.FieldByName(k)
					if fv.IsValid() {
						fv.Set(reflect.ValueOf(strings.Trim(at.Values[0], " ")))
					}
				}
			}
		}
		if us.Name == "" {
			continue
		}
		_, err = l.userService.GetByNameOrEmail(us.Name, common.DBOptions{})
		if err == nil {
			us.Available = false
		}
		users = append(users, *us)
	}
	return users, nil
}

func (l *service) TestConnect(ldap *v1Ldap.Ldap) (int, error) {
	users := 0
	if !ldap.Enable {
		return users, errors.New("请先启用LDAP")
	}

	lc := ldapClient.NewLdapClient(ldap.Address, ldap.Port, ldap.Username, ldap.Password, ldap.TLS)
	if err := lc.Connect(); err != nil {
		return users, err
	}
	attributes, err := ldap.GetAttributes()
	if err != nil {
		return users, err
	}
	entries, err := lc.Search(ldap.Dn, ldap.Filter, ldap.SizeLimit, ldap.TimeLimit, attributes)
	if err != nil {
		return users, err
	}
	if len(entries) == 0 {
		return users, nil
	}

	return len(entries), nil
}

func (l *service) CheckStatus() bool {
	ldaps, err := l.List(common.DBOptions{})
	if err != nil || len(ldaps) == 0 {
		return false
	}
	ldap := ldaps[0]
	return ldap.Enable
}

func (l *service) TestLogin(username string, password string) error {
	ldaps, err := l.List(common.DBOptions{})
	if err != nil {
		return err
	}
	if len(ldaps) == 0 {
		return errors.New("请先保存LDAP配置")
	}
	ldap := ldaps[0]

	mappings, err := ldap.GetMappings()
	if err != nil {
		return err
	}
	var userFilter string
	for k, v := range mappings {
		if k == "Name" {
			userFilter = "(" + v + "=" + username + ")"
		}
	}
	lc := ldapClient.NewLdapClient(ldap.Address, ldap.Port, ldap.Username, ldap.Password, ldap.TLS)
	if err := lc.Connect(); err != nil {
		return err
	}
	return lc.Login(ldap.Dn, userFilter, password, ldap.SizeLimit, ldap.TimeLimit)
}

func (l *service) Login(user v1User.User, password string, options common.DBOptions) error {
	ldaps, err := l.List(options)
	if err != nil {
		return err
	}
	ldap := ldaps[0]

	mappings, err := ldap.GetMappings()
	if err != nil {
		return err
	}
	var userFilter string
	for k, v := range mappings {
		if k == "Name" {
			userFilter = "(" + v + "=" + user.Name + ")"
		}
	}
	lc := ldapClient.NewLdapClient(ldap.Address, ldap.Port, ldap.Username, ldap.Password, ldap.TLS)
	if err := lc.Connect(); err != nil {
		return err
	}
	return lc.Login(ldap.Dn, userFilter, password, ldap.SizeLimit, ldap.TimeLimit)
}

func (l *service) ImportUsers(users []v1User.ImportUser) (v1User.ImportResult, error) {
	var result v1User.ImportResult
	for _, imp := range users {
		us := &v1User.User{
			NickName: imp.NickName,
			Metadata: v1.Metadata{
				Name: imp.Name,
			},
			Type:  v1User.LDAP,
			Email: imp.Email,
		}
		if us.Email == "" {
			us.Email = us.Name + "@example.com"
		}
		if us.NickName == "" {
			us.NickName = us.Name
		}

		result.Failures = append(result.Failures, us.Name)
		tx, err := server.DB().Begin(true)
		if err != nil {
			server.Logger().Errorf("create tx err:  %s", err)
			continue
		}
		err = l.userService.Create(us, common.DBOptions{DB: tx})
		if err != nil {
			_ = tx.Rollback()
			server.Logger().Errorf("can not insert user %s , err:  %s", us.Name, err)
			continue
		}
		roleName := "Common User"
		binding := v1Role.Binding{
			BaseModel: v1.BaseModel{
				Kind:       "RoleBind",
				ApiVersion: "v1",
				CreatedBy:  "admin",
			},
			Metadata: v1.Metadata{
				Name: fmt.Sprintf("role-binding-%s-%s", roleName, us.Name),
			},
			Subject: v1Role.Subject{
				Kind: "User",
				Name: us.Name,
			},
			RoleRef: roleName,
		}
		if err := l.roleBindingService.CreateRoleBinding(&binding, common.DBOptions{DB: tx}); err != nil {
			_ = tx.Rollback()
			server.Logger().Errorf("can not create  user role %s , err:  %s", us.Name, err)
			continue
		}
		_ = tx.Commit()
		result.Failures = result.Failures[:len(result.Failures)-1]
	}
	if len(result.Failures) == 0 {
		result.Success = true
	}
	return result, nil
}

func (l *service) Sync(id string, options common.DBOptions) error {
	ldap, err := l.GetById(id, options)
	if err != nil {
		return err
	}
	lc := ldapClient.NewLdapClient(ldap.Address, ldap.Port, ldap.Username, ldap.Password, ldap.TLS)
	if err := lc.Connect(); err != nil {
		return err
	}
	go func() {
		server.Logger().Info("start sync ldap user")
		insertCount := 0
		attributes, err := ldap.GetAttributes()
		if err != nil {
			server.Logger().Errorf("can not get ldap map attributes")
			return
		}
		mappings, err := ldap.GetMappings()
		if err != nil {
			server.Logger().Errorf("can not get ldap mappings")
			return
		}
		entries, err := lc.Search(ldap.Dn, ldap.Filter, ldap.SizeLimit, ldap.TimeLimit, attributes)
		if err != nil {
			server.Logger().Errorf(err.Error())
			return
		}
		for _, entry := range entries {
			us := new(v1User.User)
			rv := reflect.ValueOf(&us).Elem().Elem()

			for _, at := range entry.Attributes {
				for k, v := range mappings {
					if v == at.Name && len(at.Values) > 0 {
						fv := rv.FieldByName(k)
						if fv.IsValid() {
							fv.Set(reflect.ValueOf(strings.Trim(at.Values[0], " ")))
						}
					}
				}
			}
			if us.Email == "" || us.Name == "" {
				continue
			}
			if us.NickName == "" {
				us.NickName = us.Name
			}
			us.Type = v1User.LDAP
			_, err := l.userService.GetByNameOrEmail(us.Name, options)
			if errors.Is(err, storm.ErrNotFound) {
				tx, err := server.DB().Begin(true)
				if err != nil {
					server.Logger().Errorf("create tx err:  %s", err)
				}
				err = l.userService.Create(us, common.DBOptions{DB: tx})
				if err != nil {
					_ = tx.Rollback()
					server.Logger().Errorf("can not insert user %s , err:  %s", us.Name, err)
					continue
				}
				roleName := "Common User"
				binding := v1Role.Binding{
					BaseModel: v1.BaseModel{
						Kind:       "RoleBind",
						ApiVersion: "v1",
						CreatedBy:  "admin",
					},
					Metadata: v1.Metadata{
						Name: fmt.Sprintf("role-binding-%s-%s", roleName, us.Name),
					},
					Subject: v1Role.Subject{
						Kind: "User",
						Name: us.Name,
					},
					RoleRef: roleName,
				}
				if err := l.roleBindingService.CreateRoleBinding(&binding, common.DBOptions{DB: tx}); err != nil {
					_ = tx.Rollback()
					server.Logger().Errorf("can not create  user role %s , err:  %s", us.Name, err)
					continue
				}
				_ = tx.Commit()
				insertCount++
			}
		}

		server.Logger().Infof("sync ldap user %d , insert user %d", len(entries), insertCount)
	}()

	return nil
}
