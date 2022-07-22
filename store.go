package dbstore

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// Store database token store
type Store struct {
	tableName string
	db        *sqlx.DB
	stdout    io.Writer
	ticker    *time.Ticker
}

// SetStdout set error output
func (s *Store) SetStdout(stdout io.Writer) *Store {
	s.stdout = stdout
	return s
}

// Close close the store
func (s *Store) Close() {
	s.ticker.Stop()
	_ = s.db.DB.Close()
}

func (s *Store) gc() {
	for range s.ticker.C {
		s.clean()
	}
}

func (s *Store) clean() {
	filterMap := map[string]interface{}{"expired_at": time.Now().Unix()}
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE expired_at<=:expired_at OR (code='' AND access='' AND refresh='')", s.tableName)
	var n int
	rSQL, rQueryParas, err := s.db.BindNamed(query, filterMap)
	if err != nil {
		s.errorf(err.Error())
		return
	}
	err = s.db.QueryRowx(rSQL, rQueryParas...).Scan(&n)
	if err != nil || n == 0 {
		if err != nil {
			s.errorf(err.Error())
		}
		return
	}
	_, err = s.db.NamedExec(fmt.Sprintf("DELETE FROM %s WHERE expired_at<=:expired_at OR (code='' AND access='' AND refresh='')", s.tableName), filterMap)
	if err != nil {
		s.errorf(err.Error())
	}
}

func (s *Store) errorf(format string, args ...interface{}) {
	if s.stdout != nil {
		buf := fmt.Sprintf("[OAUTH2-database-ERROR]: "+format, args...)
		_, _ = s.stdout.Write([]byte(buf))
	}
}

// Create create and store the new token information
func (s *Store) Create(ctx context.Context, info oauth2.TokenInfo) error {
	buf, _ := jsoniter.Marshal(info)
	item := &StoreItem{
		ID:   time.Now().UnixMicro(),
		Data: string(buf),
	}

	if code := info.GetCode(); code != "" {
		item.Code = code
		item.ExpiredAt = info.GetCodeCreateAt().Add(info.GetCodeExpiresIn()).Unix()
	} else {
		item.Access = info.GetAccess()
		item.ExpiredAt = info.GetAccessCreateAt().Add(info.GetAccessExpiresIn()).Unix()

		if refresh := info.GetRefresh(); refresh != "" {
			item.Refresh = info.GetRefresh()
			item.ExpiredAt = info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn()).Unix()
		}
	}
	query := fmt.Sprintf("INSERT INTO %s(id,expired_at,code,access,refresh,data)VALUES(:id,:expired_at,:code,:access,:refresh,:data)", s.tableName)
	_, err := s.db.NamedExec(query, item)
	return err
}

// RemoveByCode delete the authorization code
func (s *Store) RemoveByCode(ctx context.Context, code string) error {
	query := fmt.Sprintf("UPDATE %s SET code='' WHERE code=:code", s.tableName)
	_, err := s.db.NamedExec(query, map[string]interface{}{"code": code})
	if err != nil && err == sql.ErrNoRows {
		return nil
	}
	return err
}

// RemoveByAccess use the access token to delete the token information
func (s *Store) RemoveByAccess(ctx context.Context, access string) error {
	query := fmt.Sprintf("UPDATE %s SET access='' WHERE access=:access", s.tableName)
	_, err := s.db.NamedExec(query, map[string]interface{}{"access": access})
	if err != nil && err == sql.ErrNoRows {
		return nil
	}
	return err
}

// RemoveByRefresh use the refresh token to delete the token information
func (s *Store) RemoveByRefresh(ctx context.Context, refresh string) error {
	query := fmt.Sprintf("UPDATE %s SET refresh='' WHERE refresh=:refresh", s.tableName)
	_, err := s.db.NamedExec(query, map[string]interface{}{"refresh": refresh})
	if err != nil && err == sql.ErrNoRows {
		return nil
	}
	return err
}

func (s *Store) toTokenInfo(data string) oauth2.TokenInfo {
	var tm models.Token
	_ = jsoniter.Unmarshal([]byte(data), &tm)
	return &tm
}

// GetByCode use the authorization code for token information data
func (s *Store) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	if code == "" {
		return nil, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE code=:code", s.tableName)
	var item StoreItem
	rSQL, rQueryParas, err := s.db.BindNamed(query, map[string]interface{}{"code": code})
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRowx(rSQL, rQueryParas...).StructScan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s.toTokenInfo(item.Data), nil
}

// GetByAccess use the access token for token information data
func (s *Store) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	if access == "" {
		return nil, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE access=:access", s.tableName)
	var item StoreItem
	rSQL, rQueryParas, err := s.db.BindNamed(query, map[string]interface{}{"access": access})
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRowx(rSQL, rQueryParas...).StructScan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s.toTokenInfo(item.Data), nil
}

// GetByRefresh use the refresh token for token information data
func (s *Store) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	if refresh == "" {
		return nil, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE refresh=:refresh", s.tableName)
	var item StoreItem
	rSQL, rQueryParas, err := s.db.BindNamed(query, map[string]interface{}{"refresh": refresh})
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRowx(rSQL, rQueryParas...).StructScan(&item)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s.toTokenInfo(item.Data), nil
}
