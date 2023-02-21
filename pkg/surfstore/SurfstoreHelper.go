package surfstore

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

/* Hash Related */
func GetBlockHashBytes(blockData []byte) []byte {
	h := sha256.New()
	h.Write(blockData)
	return h.Sum(nil)
}

func GetBlockHashString(blockData []byte) string {
	blockHash := GetBlockHashBytes(blockData)
	return hex.EncodeToString(blockHash)
}

/* File Path Related */
func ConcatPath(baseDir, fileDir string) string {
	return baseDir + "/" + fileDir
}

/*
	Writing Local Metadata File Related
*/

const createTable string = `create table if not exists indexes (
	fileName TEXT,
	version INT,
	hashIndex INT,
	hashValue TEXT
);`

const insertTuple string = `INSERT INTO indexes(fileName,version,hashIndex,hashValue) values(?,?,?,?)`

// WriteMetaFile writes the file meta map back to local metadata file index.db
func WriteMetaFile(fileMetas map[string]*FileMetaData, baseDir string) error {
	// remove index.db file if it exists
	outputMetaPath := ConcatPath(baseDir, DEFAULT_META_FILENAME)
	if _, err := os.Stat(outputMetaPath); err == nil {
		e := os.Remove(outputMetaPath)
		if e != nil {
			log.Fatal("Error During Meta Write Back: err", err)
			return err
		}
	}
	db, err := sql.Open("sqlite3", outputMetaPath)
	if err != nil {
		log.Fatal("Error During Meta Write Back: ", err)
		return err
	}
	statement, err := db.Prepare(createTable)

	if err != nil {
		log.Fatal("Error During Meta Write Back: ", err)
		return err
	}
	defer statement.Close()
	statement.Exec()
	stm, err := db.Prepare(insertTuple)
	if err != nil {
		log.Fatal("Error During initializing statement: ", err)
		return err
	}
	//traverse the map and store the information into the database for later reference
	for _, metaData := range fileMetas {
		//key == metaData.Filename
		for i, hl := range metaData.BlockHashList {
			_, err = stm.Exec(metaData.Filename, metaData.Version, i, hl)
			if err != nil {
				log.Fatal("can't write date into database: ", err)
				return err
			}
		}
	}
	return nil
}

/*
Reading Local Metadata File Related
*/
// const getDistinctFileName string = ``

// const getTuplesByFileName string = ``

// LoadMetaFromMetaFile loads the local metadata file into a file meta map.
// The key is the file's name and the value is the file's metadata.
// You can use this function to load the index.db file in this project.
func LoadMetaFromMetaFile(baseDir string) (fileMetaMap map[string]*FileMetaData, e error) {
	metaFilePath, _ := filepath.Abs(ConcatPath(baseDir, DEFAULT_META_FILENAME))
	fileMetaMap = make(map[string]*FileMetaData)
	metaFileStats, e := os.Stat(metaFilePath)
	if e != nil || metaFileStats.IsDir() {
		return fileMetaMap, nil
	}
	db, err := sql.Open("sqlite3", metaFilePath)
	if err != nil {
		log.Fatal("Error When Opening Meta: ", err)
		return fileMetaMap, err
	}
	rows, err := db.Query("SELECT * FROM `indexes`")
	if err != nil {
		log.Fatal("can't query data from database: ", err)
		return fileMetaMap, err
	}
	defer rows.Close()
	for rows.Next() {

		var fileName string
		var version int32
		var hashIndex int
		var hashValue string

		err = rows.Scan(&fileName, &version, &hashIndex, &hashValue)

		if err != nil {
			log.Fatal("err occurs when scan the database: ", err)
			return fileMetaMap, err
		}
		if _, ok := fileMetaMap[fileName]; ok {
			fileMetaMap[fileName].BlockHashList = append(fileMetaMap[fileName].BlockHashList, hashValue)
		} else {
			var data *FileMetaData = new(FileMetaData)
			data.Filename = fileName
			data.Version = version
			data.BlockHashList = append(data.BlockHashList, hashValue)
			fileMetaMap[fileName] = data
		}

	}
	return fileMetaMap, nil
}

/*
	Debugging Related
*/

// PrintMetaMap prints the contents of the metadata map.
// You might find this function useful for debugging.
func PrintMetaMap(metaMap map[string]*FileMetaData) {

	fmt.Println("--------BEGIN PRINT MAP--------")

	for _, filemeta := range metaMap {
		fmt.Println("\t", filemeta.Filename, filemeta.Version)
		for _, blockHash := range filemeta.BlockHashList {
			fmt.Println("\t", blockHash)
		}
	}

	fmt.Println("---------END PRINT MAP--------")

}
