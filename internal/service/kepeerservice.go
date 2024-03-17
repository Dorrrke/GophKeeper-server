package service

import (
	"context"
	"errors"
	"strconv"

	errText "github.com/Dorrrke/GophKeeper-server/internal/domain/errors"
	"github.com/Dorrrke/GophKeeper-server/internal/domain/models"
	"github.com/Dorrrke/GophKeeper-server/internal/storage"
	gophkeeperv1 "github.com/Dorrrke/goph-keeper-proto/gen/go/gophkeeper"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidPassword = errors.New(errText.InvalidPasswordError)

type KeepService struct {
	stor storage.KeepStorage
	log  *zerolog.Logger
}

func New(stor storage.KeepStorage, zlog *zerolog.Logger) *KeepService {
	return &KeepService{
		stor: stor,
		log:  zlog,
	}
}

func (kp *KeepService) RegisterUser(login string, pass string) (int64, error) {
	kp.log.Debug().Msg("called 'service.RegisterUser'")
	hash, err := hashPass(pass)
	if err != nil {
		kp.log.Error().Err(err).Msg("Password hashing error ")
		return -1, err
	}
	uid, err := kp.stor.SaveUser(context.Background(), models.UserModel{
		Login: login,
		Hash:  hash,
	})
	if err != nil {
		kp.log.Error().Err(err).Msg("Error when saving a user")
		return -1, err
	}
	return uid, nil
}

func (kp *KeepService) LoginUser(login string, pass string) (models.UserModel, error) {
	kp.log.Debug().Msg("called 'service.LoginUser'")
	uID, hashFromDB, err := kp.stor.GetUserHash(context.Background(), login)
	if err != nil {
		kp.log.Error().Err(err).Msg("Getting user password hash from db error")
		return models.UserModel{}, err
	}
	if !matchPass(pass, hashFromDB) {
		kp.log.Debug().Msg("Entered password and hash do not match")
		return models.UserModel{}, ErrInvalidPassword
	}
	return models.UserModel{
		UserID: uID,
		Login:  login,
		Hash:   pass,
	}, nil
}

func (kp *KeepService) SyncDB(pModel models.ProtoSyncModel, uID int) (models.ProtoSyncModel, error) {
	sModel, err := protoModelToModel(pModel, uID)
	if err != nil {
		return models.ProtoSyncModel{}, err
	}
	res, err := kp.stor.SyncDB(context.Background(), sModel, uID)
	if err != nil {
		return models.ProtoSyncModel{}, err
	}
	if err = kp.stor.ClearDB(context.Background(), uID); err != nil {
		return models.ProtoSyncModel{}, err
	}

	return modelToProtoModel(res), nil
}

func modelToProtoModel(model models.SyncModel) models.ProtoSyncModel {
	var pModel models.ProtoSyncModel
	for _, data := range model.Bins {
		bin := gophkeeperv1.SyncBinData{
			Name:    data.Name,
			Data:    data.Data,
			Deleted: data.Deleted,
			Updated: data.Updated,
		}
		pModel.Bins = append(pModel.Bins, &bin)
	}
	for _, data := range model.Auth {
		auth := gophkeeperv1.SyncAuth{
			Name:     data.Name,
			Login:    data.Login,
			Password: data.Password,
			Deleted:  data.Deleted,
			Updated:  data.Updated,
		}
		pModel.Auth = append(pModel.Auth, &auth)
	}
	for _, data := range model.Cards {
		card := gophkeeperv1.SyncCard{
			Name:    data.Name,
			Number:  data.Number,
			Date:    data.Date,
			Cvv:     strconv.Itoa(data.CVVCode),
			Deleted: data.Deleted,
			Updated: data.Updated,
		}
		pModel.Cards = append(pModel.Cards, &card)
	}
	for _, data := range model.Texts {
		text := gophkeeperv1.SyncText{
			Name:    data.Name,
			Data:    data.Data,
			Deleted: data.Deleted,
			Updated: data.Updated,
		}
		pModel.Texts = append(pModel.Texts, &text)
	}

	return pModel
}

func protoModelToModel(model models.ProtoSyncModel, uID int) (models.SyncModel, error) {
	var sModel models.SyncModel
	for _, data := range model.Bins {
		bin := models.SyncBinaryDataModel{
			UserID:  uID,
			Name:    data.Name,
			Data:    data.Data,
			Deleted: data.Deleted,
			Updated: data.Updated,
		}
		sModel.Bins = append(sModel.Bins, bin)
	}
	for _, data := range model.Auth {
		auth := models.SyncLoginModel{
			UserID:   uID,
			Name:     data.Name,
			Login:    data.Login,
			Password: data.Password,
			Deleted:  data.Deleted,
			Updated:  data.Updated,
		}
		sModel.Auth = append(sModel.Auth, auth)
	}
	for _, data := range model.Cards {
		cvv, err := strconv.Atoi(data.Cvv)
		if err != nil {
			return models.SyncModel{}, err
		}
		card := models.SyncCardModel{
			UserID:  uID,
			Name:    data.Name,
			Number:  data.Number,
			Date:    data.Date,
			CVVCode: cvv,
			Deleted: data.Deleted,
			Updated: data.Updated,
		}
		sModel.Cards = append(sModel.Cards, card)
	}
	for _, data := range model.Texts {
		text := models.SyncTextDataModel{
			UserID:  uID,
			Name:    data.Name,
			Data:    data.Data,
			Deleted: data.Deleted,
			Updated: data.Updated,
		}
		sModel.Texts = append(sModel.Texts, text)
	}

	return sModel, nil
}

func hashPass(pass string) (string, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hashedPass), nil
}

func matchPass(pass string, hashFromDB string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashFromDB), []byte(pass))
	return err == nil
}
