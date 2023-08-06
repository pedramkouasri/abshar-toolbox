package db

import (
	"fmt"
)

const (
	Format      = "PATCH-NAME-%s-%s"
	IsCompleted = "IS_COMPLETED"
	IsFailed    = "IS_FAILED"
	MessageFail = "MESSAGE_FAIL"
	Percent     = "PERCENT"
	State       = "STATE"
)

type PackageStore struct {
	version string
	db      *BoltDB
}

var store PackageStore

func init() {
	dbs := NewBoltDB()

	store = PackageStore{
		version: "",
		db:      dbs,
	}
}

func StoreError(err error) {
	store.storeError(err)
}

func StoreSuccess() {
	store.storeSuccess()
}

func StoreInfo(message string) {
	store.storeInfo(message)
}
func StorePercent(value string) {
	store.storePercent(value)
}
func StoreInit(version string) {
	store.setVersion(version)
	store.storeInit()
}

func (pdb *PackageStore) setVersion(version string) {
	pdb.version = version
}

func (pdb *PackageStore) storeError(err error) {
	pdb.db.Set(fmt.Sprintf(Format, pdb.version, IsFailed), []byte{1})
	pdb.db.Set(fmt.Sprintf(Format, pdb.version, MessageFail), []byte(err.Error()))
}

func (pdb *PackageStore) storeSuccess() {
	pdb.db.Set(fmt.Sprintf(Format, pdb.version, IsCompleted), []byte{1})
	pdb.db.Set(fmt.Sprintf(Format, pdb.version, State), []byte("Completed :)"))
}

func (pdb *PackageStore) storeInfo(message string) {
	pdb.db.Set(fmt.Sprintf(Format, pdb.version, State), []byte(message))
}

func (pdb *PackageStore) storePercent(value string) {
	pdb.db.Set(fmt.Sprintf(Format, pdb.version, Percent), []byte(value))
}

func (pdb *PackageStore) storeInit() {
	pdb.db.Set(fmt.Sprintf(Format, pdb.version, IsCompleted), []byte{0})
	pdb.db.Set(fmt.Sprintf(Format, pdb.version, IsFailed), []byte{0})
	pdb.db.Set(fmt.Sprintf(Format, pdb.version, MessageFail), []byte{})
	pdb.db.Set(fmt.Sprintf(Format, pdb.version, Percent), []byte("0"))
	pdb.db.Set(fmt.Sprintf(Format, pdb.version, State), []byte{})
}
