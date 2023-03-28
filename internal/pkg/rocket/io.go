package rocket

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/ptonlix/spokenai/pkg/file"
)

type ioer interface {
	IsExists(userid, roleid string) bool
	ReadData(userid, roleid string) ([]byte, error)
	SaveData(userid, roleid string, data []byte) error
	BackupData(userid, roleid string) error
}

func NewIO(iotype bool, opt *option) ioer {
	if iotype {
		return &FileIOer{DirPath: opt.dataDir}
	}
	return &DbIOer{}
}

type FileIOer struct {
	DirPath string
}

func (f *FileIOer) IsExists(userid, roleid string) bool {

	filepath := f.DirPath + userid + "_" + roleid + ".json"

	_, flag := file.IsExists(filepath)

	return flag
}

func (f *FileIOer) ReadData(userid, roleid string) ([]byte, error) {

	filepath := f.DirPath + userid + "_" + roleid + ".json"

	return ioutil.ReadFile(filepath)
}

func (f *FileIOer) SaveData(userid, roleid string, data []byte) error {

	filepath := f.DirPath + userid + "_" + roleid + ".json"

	return ioutil.WriteFile(filepath, data, 666)
}

func (f *FileIOer) BackupData(userid, roleid string) error {

	filepath := f.DirPath + userid + "_" + roleid + ".json"
	//TODO 增加最大备份数限制
	return os.Rename(filepath, filepath+time.Now().Format("20060102150405"))
}

type DbIOer struct {
	// TODO
}

func (d *DbIOer) IsExists(userid, roleid string) bool {

	return true
}

func (d *DbIOer) ReadData(userid, roleid string) ([]byte, error) {

	return []byte{}, nil
}

func (d *DbIOer) SaveData(userid, roleid string, data []byte) error {

	return nil
}

func (d *DbIOer) BackupData(userid, roleid string) error {

	return nil
}
