package models

import gophkeeperv1 "github.com/Dorrrke/goph-keeper-proto/gen/go/gophkeeper"

type CardModel struct {
	Name    string
	Number  string
	Date    string
	CVVCode int
}

type LoginModel struct {
	Name     string
	Login    string
	Password string
}

type TextDataModel struct {
	Name string
	Data string
}

type BinaryDataModel struct {
	Name string
	Data []byte
}

type SyncModel struct {
	Cards []SyncCardModel
	Texts []SyncTextDataModel
	Bins  []SyncBinaryDataModel
	Auth  []SyncLoginModel
}

type SyncCardModel struct {
	UserID  int
	Name    string
	Number  string
	Date    string
	CVVCode int
	Deleted bool
	Updated string
}

type SyncLoginModel struct {
	UserID   int
	Name     string
	Login    string
	Password string
	Deleted  bool
	Updated  string
}

type SyncTextDataModel struct {
	UserID  int
	Name    string
	Data    string
	Deleted bool
	Updated string
}

type SyncBinaryDataModel struct {
	UserID  int
	Name    string
	Data    []byte
	Deleted bool
	Updated string
}

type ProtoSyncModel struct {
	Cards []*gophkeeperv1.SyncCard
	Texts []*gophkeeperv1.SyncText
	Bins  []*gophkeeperv1.SyncBinData
	Auth  []*gophkeeperv1.SyncAuth
}

type UserModel struct {
	UserID int64  `json:"u_id"`
	Login  string `json:"login"`
	Hash   string `json:"hash"`
}
