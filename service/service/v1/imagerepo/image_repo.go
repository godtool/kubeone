package imagerepo

import (
	"errors"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	costomStorm "github.com/godtool/kubeone/pkg/storm"
	repoClient "github.com/godtool/kubeone/pkg/util/imagerepo"
	"github.com/godtool/kubeone/pkg/util/imagerepo/repos"
	"github.com/godtool/kubeone/pkg/util/lang"
	V1ClusterRepo "github.com/godtool/kubeone/service/model/v1/clusterrepo"
	V1ImageRepo "github.com/godtool/kubeone/service/model/v1/imagerepo"
	"github.com/godtool/kubeone/service/service/v1/common"
	"github.com/google/uuid"
)

type Service interface {
	common.DBService
	ListInternalRepos(repo V1ImageRepo.ImageRepo, page, limit int, search string) (names []string, err error)
	Search(num, size int, conditions common.Conditions, options common.DBOptions) (result []V1ImageRepo.ImageRepo, count int, err error)
	Create(repo *V1ImageRepo.ImageRepo, options common.DBOptions) (err error)
	Delete(name string, options common.DBOptions) (err error)
	GetByName(name string, options common.DBOptions) (repo V1ImageRepo.ImageRepo, err error)
	UpdateRepo(name string, repo *V1ImageRepo.ImageRepo, options common.DBOptions) (err error)
	ListByCluster(cluster string, options common.DBOptions) (result []V1ImageRepo.ImageRepo, err error)
	ListImages(repo, cluster string, options common.DBOptions) (names []string, err error)
	ListImagesByRepo(repo string, page, limit int, search, token string, options common.DBOptions) (res V1ImageRepo.RepoResponse, err error)
}

func NewService() Service {
	return &service{}
}

type service struct {
	common.DefaultDBService
}

func (s *service) ListInternalRepos(repo V1ImageRepo.ImageRepo, page, limit int, search string) (names []string, err error) {

	client := repoClient.NewClient(repos.Config{
		Type:     repo.Type,
		EndPoint: repo.EndPoint,
		Credential: repos.Credential{
			Username: repo.Credential.Username,
			Password: repo.Credential.Password,
		},
		Version: repo.Version,
	})
	if client == nil {
		return nil, errors.New("repo client is not found")
	}
	request := repos.ProjectRequest{
		Name:  search,
		Page:  page,
		Limit: limit,
	}

	return client.ListRepos(request)
}

func (s *service) ListImages(repo, cluster string, options common.DBOptions) (names []string, err error) {
	db := s.GetDB(options)
	query := db.Select(q.And(q.Eq("Cluster", cluster), q.Eq("Repo", repo)))
	var cRepo V1ClusterRepo.ClusterRepo
	if err = query.First(&cRepo); err != nil {
		return
	}
	rp, err := s.GetByName(repo, options)
	if err != nil {
		return
	}
	client := repoClient.NewClient(repos.Config{
		Type:     rp.Type,
		EndPoint: rp.EndPoint,
		Credential: repos.Credential{
			Username: rp.Credential.Username,
			Password: rp.Credential.Password,
		},
		Version: rp.Version,
	})
	images, err := client.ListImagesWithoutPage(rp.RepoName)
	if err != nil {
		return
	}
	for _, image := range images {
		names = append(names, rp.DownloadUrl+"/"+image)
	}
	return
}

func (s *service) ListImagesByRepo(repo string, page, limit int, search, token string, options common.DBOptions) (response V1ImageRepo.RepoResponse, err error) {
	rp, err1 := s.GetByName(repo, options)
	if err1 != nil {
		err = err1
		return
	}
	client := repoClient.NewClient(repos.Config{
		Type:     rp.Type,
		EndPoint: rp.EndPoint,
		Credential: repos.Credential{
			Username: rp.Credential.Username,
			Password: rp.Credential.Password,
		},
		Version: rp.Version,
	})
	request := repos.RepoRequest{
		Repo:          rp.RepoName,
		Page:          page,
		Limit:         limit,
		Search:        search,
		ContinueToken: token,
	}
	res, err2 := client.ListImages(request)
	if err2 != nil {
		err = err2
		return
	}
	var names []string
	for _, image := range res.Items {
		names = append(names, rp.DownloadUrl+"/"+image)
	}
	response.Items = names
	response.ContinueToken = res.ContinueToken
	return
}

func (s *service) ListByCluster(cluster string, options common.DBOptions) (result []V1ImageRepo.ImageRepo, err error) {
	db := s.GetDB(options)
	query := db.Select(q.Eq("Cluster", cluster))
	var clusterrepos []V1ClusterRepo.ClusterRepo
	if err = query.Find(&clusterrepos); err != nil && err != storm.ErrNotFound {
		return
	}
	if len(clusterrepos) > 0 {
		group := make([]string, 0)
		for _, repo := range clusterrepos {
			group = append(group, repo.Repo)
		}
		query2 := db.Select(q.Not(q.In("Name", group)))
		if err = query2.Find(&result); err != nil {
			return
		}
	} else {
		if err = db.All(&result); err != nil {
			return
		}
	}
	return
}

func (s *service) Search(num, size int, conditions common.Conditions, options common.DBOptions) (result []V1ImageRepo.ImageRepo, count int, err error) {
	db := s.GetDB(options)
	var ms []q.Matcher
	for k := range conditions {
		if conditions[k].Field == "quick" {
			ms = append(ms, q.Or(
				costomStorm.Like("Name", conditions[k].Value),
			))
		} else {
			field := lang.FirstToUpper(conditions[k].Field)
			value := lang.ParseValueType(conditions[k].Value)

			switch conditions[k].Operator {
			case "eq":
				ms = append(ms, q.Eq(field, value))
			case "ne":
				ms = append(ms, q.Not(q.Eq(field, value)))
			case "like":
				ms = append(ms, costomStorm.Like(field, value.(string)))
			case "not like":
				ms = append(ms, q.Not(costomStorm.Like(field, value.(string))))
			}
		}
	}
	query := db.Select(ms...).OrderBy("CreateAt").Reverse()
	count, err = query.Count(&V1ImageRepo.ImageRepo{})
	if err != nil {
		return
	}
	if size != 0 {
		query.Limit(size).Skip((num - 1) * size)
	}
	if err = query.Find(&result); err != nil {
		return
	}
	return
}

func (s *service) Create(repo *V1ImageRepo.ImageRepo, options common.DBOptions) (err error) {
	db := s.GetDB(options)
	repo.UUID = uuid.New().String()
	repo.CreateAt = time.Now()
	repo.UpdateAt = time.Now()
	return db.Save(repo)
}

func (s *service) Delete(name string, options common.DBOptions) (err error) {
	db := s.GetDB(options)
	item, err1 := s.GetByName(name, options)
	if err1 != nil {
		err = err1
		return
	}
	return db.DeleteStruct(&item)
}

func (s *service) GetByName(name string, options common.DBOptions) (repo V1ImageRepo.ImageRepo, err error) {
	db := s.GetDB(options)
	query := db.Select(q.Eq("Name", name))
	if err = query.First(&repo); err != nil {
		return
	}
	return
}

func (s *service) UpdateRepo(name string, repo *V1ImageRepo.ImageRepo, options common.DBOptions) (err error) {
	db := s.GetDB(options)
	old, err1 := s.GetByName(name, options)
	if err1 != nil {
		err = err1
		return
	}
	repo.UUID = old.UUID
	repo.CreateAt = old.CreateAt
	repo.UpdateAt = time.Now()

	if !old.Auth {
		repo.Credential.Password = ""
		repo.Credential.Username = ""
		repo.Credential = V1ImageRepo.Credential{}
		err = db.UpdateField(repo, "Credential", repo.Credential)
		if err != nil {
			return
		}
	}

	if old.AllowAnonymous != repo.AllowAnonymous {
		err = db.UpdateField(repo, "AllowAnonymous", repo.AllowAnonymous)
		if err != nil {
			return
		}
	}

	if old.Auth != repo.Auth {
		err = db.UpdateField(repo, "Auth", repo.Auth)
		if err != nil {
			return
		}
	}

	return db.Update(repo)
}
