package campaigns

import (
	"errors"

	"github.com/DaanV2/mechanus/server/pkg/database"
	"github.com/charmbracelet/log"
)

const (
	TABLE_CAMPAIGNS database.TableName = "campaigns"
)

type Service struct {
	db      *database.Database
	dbtable *database.Table[Campaign]
	logger  *log.Logger
}

func NewService(db *database.Database) *Service {
	return &Service{
		db:      db,
		dbtable: database.GetTable[Campaign](db, TABLE_CAMPAIGNS),
		logger:  log.Default().WithPrefix("campaigns"),
	}
}

func (s *Service) Get(id string) (Campaign, error) {
	return s.dbtable.Get(id)
}

func (s *Service) Update(camp Campaign) (Campaign, error) {
	c, err := s.Get(camp.ID)
	if err != nil {
		return camp, nil
	}

	camp.BaseItem = c.BaseItem.Update()

	return camp, s.dbtable.Set(camp.ID, camp)
}

func (s *Service) Create(camp Campaign) (Campaign, error) {
	camp.BaseItem = database.NewBaseItem()

	_, err := s.dbtable.Get(camp.ID)
	if !errors.Is(err, database.ErrNotFound) {
		return camp, database.ErrAlreadyExists
	}

	err = s.dbtable.Set(camp.ID, camp)
	return camp, err
}
