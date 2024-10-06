package campaigns

import (
	"errors"

	"github.com/DaanV2/mechanus/server/pkg/database"
)

type Service struct {
	db      *database.Database
	storage *database.Table[Campaign]
}

func (s *Service) Get(id string) (Campaign, error) {
	return s.storage.Get(id)
}

func (s *Service) Update(camp Campaign) (Campaign, error) {
	c, err := s.Get(camp.ID)
	if err != nil {
		return camp, nil
	}

	camp.BaseItem = c.BaseItem.Update()
	
	return camp, s.storage.Set(camp.ID, camp)
}

func (s *Service) Create(camp Campaign) (Campaign, error) {
	camp.BaseItem = database.NewBaseItem()

	_, err := s.storage.Get(camp.ID)
	if !errors.Is(err, database.ErrNotFound) {
		return camp, database.ErrAlreadyExists
	}

	err = s.storage.Set(camp.ID, camp)
	return camp, err
}