package models

import (
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartblock/gta-api/pkg/e"
)

// EntSocialMedia struct
type EntSocialMedia struct {
	ID        int       `gorm:"primary_key" json:"id"`
	Name      string    `json:"name"`
	IconPath  string    `json:"icon_path"`
	Url       string    `json:"url"`
	Status    string    `json:"status"`
	SeqNo     int       `json:"seq_no"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// GetEntSocialMediaFn
func GetEntSocialMediaFn(arrCond []WhereCondFn, debug bool) ([]*EntSocialMedia, error) {
	var result []*EntSocialMedia

	tx := db.Table("ent_social_media").
		Order("ent_social_media.seq_no ASC")

	if len(arrCond) > 0 {
		for _, v := range arrCond {
			tx = tx.Where(v.Condition, v.CondValue)
		}
	}
	if debug {
		tx = tx.Debug()
	}
	err := tx.Find(&result).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, &e.CustomError{HTTPCode: http.StatusInternalServerError, Code: e.ERROR, Msg: err.Error(), Data: err}
	}

	return result, nil
}
