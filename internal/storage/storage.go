package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	errText "github.com/Dorrrke/GophKeeper-server/internal/domain/errors"
	models "github.com/Dorrrke/GophKeeper-server/internal/domain/models"
	sqlquere "github.com/Dorrrke/GophKeeper-server/internal/domain/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

var (
	ErrUserAlredyExist  = errors.New(errText.UserExistsError)
	ErrUserNotExist     = errors.New(errText.UserNotExistError)
	ErrCardAlredyExist  = errors.New(errText.CardExistsError)
	ErrLoginAlredyExist = errors.New(errText.LoginExistsError)
	ErrTextAlredyExist  = errors.New(errText.TextExistsError)
	ErrBinAlredyExist   = errors.New(errText.BinDataExistsError)
	ErrCardNotExist     = errors.New(errText.CardNotExistsError)
	ErrLoginNotExist    = errors.New(errText.LoginNotExistsError)
	ErrTextNotExist     = errors.New(errText.TextNotExistsError)
	ErrBinDataNotExist  = errors.New(errText.BinDataNotExistsError)
)

type KeepStorage struct {
	db   *pgxpool.Pool
	zlog *zerolog.Logger
}

func New(pool *pgxpool.Pool, zlog *zerolog.Logger) *KeepStorage {
	return &KeepStorage{
		zlog: zlog,
		db:   pool,
	}
}

func (s *KeepStorage) SaveUser(ctx context.Context, user models.UserModel) (int64, error) {
	row := s.db.QueryRow(ctx, "INSERT INTO users(login, hash) VALUES($1, $2) RETURNING  uid", user.Login, user.Hash)
	var uid int64
	if err := row.Scan(&uid); err != nil {
		s.zlog.Debug().Err(err).Msg("Save user into db error")
		return -1, err
	}
	return uid, nil
}

func (s *KeepStorage) GetUserHash(ctx context.Context, login string) (int64, string, error) {
	row := s.db.QueryRow(ctx, "SELECT uId, hash FROM users WHERE login = $1", login)
	var uID int64
	var hash string
	if err := row.Scan(&uID, &hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, "", ErrUserNotExist
		}
		s.zlog.Debug().Err(err).Msg("Get user hash from db error")
		return -1, "", err
	}
	return uID, strings.TrimSpace(hash), nil
}

func (s *KeepStorage) ClearDB(ctx context.Context, uID int) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Prepare(ctx, "authClear", "DELETE FROM logins WHERE deleted = true AND uId = $1"); err != nil {
		return err
	}
	if _, err := tx.Prepare(ctx, "textClear", "DELETE FROM text_data WHERE deleted = true AND uId = $1"); err != nil {
		return err
	}
	if _, err := tx.Prepare(ctx, "binClear", "DELETE FROM binares_data WHERE deleted = true AND uId = $1"); err != nil {
		return err
	}
	if _, err := tx.Prepare(ctx, "cardClear", "DELETE FROM cards WHERE deleted = true AND uId = $1"); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "authClear", uID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "textClear", uID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "binClear", uID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "cardClear", uID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *KeepStorage) SyncDB(
	ctx context.Context, model models.SyncModel, uId int) (models.SyncModel, error) {
	s.zlog.Debug().Any("sync data", model).Int("User ID", uId).Msg("Run sync")
	var sTexts []models.SyncTextDataModel
	var logins []models.SyncLoginModel
	var sBins []models.SyncBinaryDataModel
	var sCards []models.SyncCardModel
	// textCh := make(chan []models.SyncTextDataModel)
	group, gCtx := errgroup.WithContext(ctx)
	// Текст.
	group.Go(func() error {
		s.zlog.Debug().Msg("Run text sync")
		tx, err := s.db.Begin(gCtx)
		if err != nil {
			s.zlog.Debug().Err(err).Msg("Begin tx error")
			return err
		}
		defer tx.Rollback(gCtx)
		var names []string
		for _, data := range model.Texts {
			names = append(names, data.Name)
			cTag, err := tx.Exec(gCtx, sqlquere.SyncUpdateTextTable,
				data.Name, data.Data, data.UserID, data.Deleted, data.Updated, data.Name, data.Updated, data.UserID,
				data.Name, data.Data, data.UserID, data.Deleted, data.Updated,
				data.Name, data.UserID)
			if err != nil {
				s.zlog.Error().Err(err).Msg("Update text table error")
				return err
			}
			s.zlog.Debug().Any("commandTag", cTag).Msg("update text result")

			rows := tx.QueryRow(gCtx, sqlquere.SynceTextTableActual, data.Name, data.Updated)
			var text models.SyncTextDataModel
			var updated time.Time
			err = rows.Scan(&text.Name, &text.Data, &text.UserID, &text.Deleted, &updated)
			text.Updated = updated.Format(time.RFC3339)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					s.zlog.Error().Err(err).Msg("Scan actual text data error")
					return err
				}
			} else {
				text.Name = strings.TrimSpace(text.Name)
				text.Data = strings.TrimSpace(text.Data)
				sTexts = append(sTexts, text)
			}
		}
		placeholders := make([]string, len(names))
		for i, name := range names {
			placeholders[i] = fmt.Sprintf("'%s'", name)
		}
		query := fmt.Sprintf(sqlquere.SynceNewTextData, strings.Join(placeholders, ","))
		s.zlog.Debug().Str("queue", query).Msg("Text new data quere")
		s.zlog.Debug().Any("Names", names).Msg("Check")
		rows, err := tx.Query(gCtx, query, uId)
		if err != nil {
			s.zlog.Error().Err(err).Msg("sql query error")
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var data models.SyncTextDataModel
			var updated time.Time
			err := rows.Scan(&data.Name, &data.Data, &data.UserID, &data.Deleted, &updated)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					s.zlog.Error().Err(err).Msg("Scan new text data error")
					return err
				}
			} else {
				data.Updated = updated.Format(time.RFC3339)
				data.Name = strings.TrimSpace(data.Name)
				data.Data = strings.TrimSpace(data.Data)
				s.zlog.Debug().Any("New text", data).Msg("Get new text")
				sTexts = append(sTexts, data)
			}
		}
		s.zlog.Debug().Msg("Sync text end")
		return tx.Commit(gCtx)
	})

	// Логины.
	group.Go(func() error {
		s.zlog.Debug().Msg("Run auth sync")
		tx, err := s.db.Begin(gCtx)
		if err != nil {
			s.zlog.Debug().Err(err).Msg("Begin tx error")
			return err
		}
		defer tx.Rollback(gCtx)

		var names []string
		for _, data := range model.Auth {
			names = append(names, data.Name)
			cTag, err := tx.Exec(gCtx, sqlquere.SyncUpdateAuthTable,
				data.Name, data.Login, data.Password, data.UserID, data.Deleted, data.Updated, data.Name, data.Updated, data.UserID,
				data.Name, data.Login, data.Password, data.UserID, data.Deleted, data.Updated, data.Name, data.UserID)
			if err != nil {
				s.zlog.Error().Err(err).Msg("Update auth table error")
				return err
			}
			s.zlog.Debug().Any("commandTag", cTag).Msg("update auth result")
			row := tx.QueryRow(gCtx, sqlquere.SynceLoginsTableActual, data.Name, data.Updated)
			var login models.SyncLoginModel
			var updated time.Time
			err = row.Scan(&login.Name, &login.Login, &login.Password, &login.UserID, &login.Deleted, &updated)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					s.zlog.Error().Err(err).Msg("Scan actual auth data error")
					return err
				}
			} else {
				login.Updated = updated.Format(time.RFC3339)
				login.Name = strings.TrimSpace(login.Name)
				login.Login = strings.TrimSpace(login.Login)
				login.Password = strings.TrimSpace(login.Password)
				logins = append(logins, login)
			}
		}
		placeholders := make([]string, len(names))
		for i, name := range names {
			placeholders[i] = fmt.Sprintf("'%s'", name)
		}
		query := fmt.Sprintf(sqlquere.SynceNewAuthData, strings.Join(placeholders, ","))
		s.zlog.Debug().Str("queue", query).Msg("Text new data quere")
		s.zlog.Debug().Any("Names", names).Msg("Check")
		rows, err := tx.Query(gCtx, query, uId)
		if err != nil {
			s.zlog.Error().Err(err).Msg("sql query error")
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var login models.SyncLoginModel
			var updated time.Time
			err := rows.Scan(&login.Name, &login.Login, &login.Password, &login.UserID, &login.Deleted, &updated)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					s.zlog.Error().Err(err).Msg("Scan new auth data error")
					return err
				}
			} else {
				login.Updated = updated.Format(time.RFC3339)
				login.Name = strings.TrimSpace(login.Name)
				login.Login = strings.TrimSpace(login.Login)
				login.Password = strings.TrimSpace(login.Password)
				logins = append(logins, login)
			}
		}
		s.zlog.Debug().Msg("Sync auth end")
		return tx.Commit(gCtx)
	})

	// Бинари.
	group.Go(func() error {
		s.zlog.Debug().Msg("Run bin sync")
		tx, err := s.db.Begin(gCtx)
		if err != nil {
			s.zlog.Debug().Err(err).Msg("Begin tx error")
			return err
		}
		defer tx.Rollback(gCtx)

		var names []string
		for _, data := range model.Bins {
			names = append(names, data.Name)
			cTag, err := tx.Exec(gCtx, sqlquere.SyncUpdateBinTable,
				data.Name, data.Data, data.UserID, data.Deleted, data.Updated, data.Name,
				data.Updated, data.UserID, data.Name, data.Data, data.UserID, data.Deleted, data.Updated,
				data.Name, data.UserID)
			if err != nil {
				s.zlog.Error().Err(err).Msg("Update bin table error")
				return err
			}
			s.zlog.Debug().Any("commandTag", cTag).Msg("update bin result")
			row := tx.QueryRow(gCtx, sqlquere.SynceBinTableActual, data.Name, data.Updated)
			var bin models.SyncBinaryDataModel
			var updated time.Time
			err = row.Scan(&bin.Name, &bin.Data, &bin.UserID, &bin.Deleted, &updated)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					s.zlog.Error().Err(err).Msg("Scan actual bin data error")
					return err
				}
			} else {
				bin.Updated = updated.Format(time.RFC3339)
				bin.Name = strings.TrimSpace(bin.Name)
				sBins = append(sBins, bin)
			}
		}
		placeholders := make([]string, len(names))
		for i, name := range names {
			placeholders[i] = fmt.Sprintf("'%s'", name)
		}
		query := fmt.Sprintf(sqlquere.SynceNewBinData, strings.Join(placeholders, ","))

		s.zlog.Debug().Str("queue", query).Msg("Text new data quere")
		s.zlog.Debug().Any("Names", names).Msg("Check")
		rows, err := tx.Query(gCtx, query, uId)
		if err != nil {
			s.zlog.Error().Err(err).Msg("sql query error")
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var bin models.SyncBinaryDataModel
			var updated time.Time
			err := rows.Scan(&bin.Name, &bin.Data, &bin.UserID, &bin.Deleted, &updated)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					s.zlog.Error().Err(err).Msg("Scan new bin data error")
					return err
				}
			} else {
				bin.Updated = updated.Format(time.RFC3339)
				bin.Name = strings.TrimSpace(bin.Name)
				sBins = append(sBins, bin)
			}
		}

		s.zlog.Debug().Msg("Sync bin end")
		return tx.Commit(gCtx)
	})

	// Карты
	group.Go(func() error {
		s.zlog.Debug().Msg("Run card sync")
		tx, err := s.db.Begin(gCtx)
		if err != nil {
			s.zlog.Debug().Err(err).Msg("Begin tx error")
			return err
		}
		defer tx.Rollback(gCtx)

		var names []string
		for _, data := range model.Cards {
			names = append(names, data.Name)
			cTag, err := tx.Exec(gCtx, sqlquere.SyncUpdateCardTable,
				data.Name, data.Number, data.Date, data.CVVCode, data.UserID, data.Deleted, data.Updated, data.Name, data.Updated, data.UserID,
				data.Name, data.Number, data.Date, data.CVVCode, data.UserID, data.Deleted, data.Updated, data.Name, data.UserID)
			if err != nil {
				s.zlog.Error().Err(err).Msg("Update Card table error")
				return err
			}
			s.zlog.Debug().Any("commandTag", cTag).Msg("update Card result")
			row := tx.QueryRow(gCtx, sqlquere.SynceCardTableActual, data.Name, data.Updated)
			var card models.SyncCardModel
			var updated time.Time
			err = row.Scan(&card.Name, &card.Number, &card.Date, &card.CVVCode, &card.UserID, &card.Deleted, &updated)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					s.zlog.Error().Err(err).Msg("Scan actual card data error")
					return err
				}
			} else {
				card.Updated = updated.Format(time.RFC3339)
				card.Name = strings.TrimSpace(card.Name)
				card.Number = strings.TrimSpace(card.Number)
				card.Date = strings.TrimSpace(card.Date)
				sCards = append(sCards, card)
			}
		}

		placeholders := make([]string, len(names))
		for i, name := range names {
			placeholders[i] = fmt.Sprintf("'%s'", name)
		}
		query := fmt.Sprintf(sqlquere.SynceNewCardData, strings.Join(placeholders, ","))
		s.zlog.Debug().Str("queue", query).Msg("Text new data quere")
		s.zlog.Debug().Any("Names", names).Msg("Check")
		rows, err := tx.Query(gCtx, query, uId)
		if err != nil {
			s.zlog.Error().Err(err).Msg("sql query error")
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var card models.SyncCardModel
			var updated time.Time
			err := rows.Scan(&card.Name, &card.Number, &card.Date, &card.CVVCode, &card.UserID, &card.Deleted, &updated)
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					s.zlog.Error().Err(err).Msg("Scan new card data error")
					return err
				}
			} else {
				card.Updated = updated.Format(time.RFC3339)
				card.Name = strings.TrimSpace(card.Name)
				card.Number = strings.TrimSpace(card.Number)
				card.Date = strings.TrimSpace(card.Date)
				sCards = append(sCards, card)
			}
		}

		s.zlog.Debug().Msg("Sync card end")
		return tx.Commit(gCtx)
	})

	err := group.Wait()
	if err != nil {
		s.zlog.Error().Err(err).Msg("Error group error")
		return models.SyncModel{}, err
	}
	s.zlog.Debug().Msg("Sync done")
	s.zlog.Debug().Int("actual text", len(sTexts)).
		Int("actual bin", len(sBins)).
		Int("actual auth", len(logins)).
		Int("actual cards", len(sCards)).Any("text", sTexts).Msg("Actual data len")
	return models.SyncModel{
		Texts: sTexts,
		Auth:  logins,
		Cards: sCards,
		Bins:  sBins,
	}, nil
}
