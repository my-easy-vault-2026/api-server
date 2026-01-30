package services

import (
	"shared-modules/entities"
	"testing"

	"github.com/jinzhu/copier"
)

func TestLoginOrCreate(t *testing.T) {

	form := &entities.LoginOrCreateForm{
		Email: "asdadsad",
	}
	vo := &entities.LoginOrCreateVO{}
	copier.Copy(vo, form)
	t.Log(vo)
}
