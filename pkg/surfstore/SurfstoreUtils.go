package surfstore

import (
	"database/sql"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"reflect"
)

// Implement the logic for a client syncing with the server here.
func ClientSync(client RPCClient) {
	indexPath := ConcatPath(client.BaseDir, "index.db")
	if _, err := os.Stat(indexPath); errors.Is(err, os.ErrNotExist) {
		_, err := os.Create(indexPath)
		if err != nil {
			log.Println("error when creating file: ", err)
		}
		// defer indexFile.Close()
	}
	db, err := sql.Open("sqlite3", "./"+indexPath)
	if err != nil {
		log.Fatal("Error During opening the database: ", err)
	}
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal("Error During creating the table: ", err)
	}

	files, err := ioutil.ReadDir(client.BaseDir)
	if err != nil {
		log.Println("Error when reading basedir: ", err)
	}
	localIndex, err := LoadMetaFromMetaFile(client.BaseDir)
	// PrintMetaMap(localIndex)
	// log.Println("                 ")
	if err != nil {
		log.Println("Could not load meta from meta file: ", err)
	}
	//Sync local index
	hashMap := make(map[string][]string)
	for _, file := range files {
		if file.Name() == "index.db" {
			continue
		}
		var numBlocks int = int(math.Ceil(float64(file.Size()) / float64(client.BlockSize)))
		fileToRead, err := os.Open(client.BaseDir + "/" + file.Name())
		if err != nil {
			log.Println("Error reading file in basedir: ", err)
		}

		for i := 0; i < numBlocks; i++ {
			byteSlice := make([]byte, client.BlockSize)
			len, err := fileToRead.Read(byteSlice)
			if err != nil {
				log.Println("Error reading bytes from file in basedir: ", err)
			}
			byteSlice = byteSlice[:len]
			hash := GetBlockHashString(byteSlice)
			hashMap[file.Name()] = append(hashMap[file.Name()], hash)
		}

		if val, ok := localIndex[file.Name()]; ok {
			if !reflect.DeepEqual(hashMap[file.Name()], val.BlockHashList) {
				localIndex[file.Name()].BlockHashList = hashMap[file.Name()]
				localIndex[file.Name()].Version++
			}
		} else {
			// New file
			meta := FileMetaData{Filename: file.Name(), Version: 1, BlockHashList: hashMap[file.Name()]}
			localIndex[file.Name()] = &meta
		}
	}
	// PrintMetaMap(localIndex)
	// log.Println("                 ")

	//Check for deleted files
	for fileName, metaData := range localIndex {
		if _, ok := hashMap[fileName]; !ok {
			//not being deleted by far
			if len(metaData.BlockHashList) != 1 || metaData.BlockHashList[0] != "0" {
				metaData.Version++
				metaData.BlockHashList = []string{"0"}
			}
		}
	}
	//now localindex has been fully adjusted according to the files in root dir
	// PrintMetaMap(localIndex)
	// log.Println("                 ")

	var blockStoreAddr string
	if err := client.GetBlockStoreAddr(&blockStoreAddr); err != nil {
		log.Println("Could not get blockStoreAddr: ", err)
	}

	remoteIndex := make(map[string]*FileMetaData)
	if err := client.GetFileInfoMap(&remoteIndex); err != nil {
		log.Println("Error getting index from server: ", err)
	}

	//Check if server has local files, upload changes
	for fileName, localMetaData := range localIndex {
		if remoteMetaData, ok := remoteIndex[fileName]; ok {
			if localMetaData.Version > remoteMetaData.Version {
				//if there are some files in local index that have been updated
				upload(client, localMetaData, blockStoreAddr)
			}
		} else {
			//if there are some files that are newly added
			upload(client, localMetaData, blockStoreAddr)
		}
	}

	//Check for updates on server, download
	for filename, remoteMetaData := range remoteIndex {
		if localMetaData, ok := localIndex[filename]; ok {
			if localMetaData.Version < remoteMetaData.Version {
				//if there are some files that need to be downlowded
				download(client, localMetaData, remoteMetaData, blockStoreAddr)
			} else if localMetaData.Version == remoteMetaData.Version && !reflect.DeepEqual(localMetaData.BlockHashList, remoteMetaData.BlockHashList) {
				download(client, localMetaData, remoteMetaData, blockStoreAddr)
			}
		} else {
			localIndex[filename] = &FileMetaData{}
			localMetaData := localIndex[filename]
			download(client, localMetaData, remoteMetaData, blockStoreAddr)
		}
	}
	//before closeing the client, update the local database
	err = WriteMetaFile(localIndex, client.BaseDir)
	if err != nil {
		log.Println("error occurs when syn: ", err)
	}
}

func upload(client RPCClient, metaData *FileMetaData, blockStoreAddr string) error {
	path := ConcatPath(client.BaseDir, metaData.Filename)
	var latestVersion int32
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err = client.UpdateFile(metaData, &latestVersion)
		if err != nil {
			log.Println("Could not upload file: ", err)
		}
		metaData.Version = latestVersion
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		log.Println("Error opening file: ", err)
	}
	defer file.Close()

	fileStat, _ := os.Stat(path)
	var numBlocks int = int(math.Ceil(float64(fileStat.Size()) / float64(client.BlockSize)))
	for i := 0; i < numBlocks; i++ {
		byteSlice := make([]byte, client.BlockSize)
		len, err := file.Read(byteSlice)
		if err != nil && err != io.EOF {
			log.Println("Error reading bytes from file in basedir: ", err)
		}
		byteSlice = byteSlice[:len]

		block := Block{BlockData: byteSlice, BlockSize: int32(len)}

		var succ bool
		if err := client.PutBlock(&block, blockStoreAddr, &succ); err != nil {
			log.Println("Failed to put block: ", err)
		}
	}

	if err := client.UpdateFile(metaData, &latestVersion); err != nil {
		log.Println("Failed to update file: ", err)
		metaData.Version = -1
	}
	metaData.Version = latestVersion
	return nil
}

func download(client RPCClient, localMetaData *FileMetaData, remoteMetaData *FileMetaData, blockStoreAddr string) error {
	path := ConcatPath(client.BaseDir, remoteMetaData.Filename)
	file, err := os.Create(path)
	if err != nil {
		log.Println("Error creating file: ", err)
	}
	defer file.Close()
	*localMetaData = *remoteMetaData
	//File deleted in server
	if len(remoteMetaData.BlockHashList) == 1 && remoteMetaData.BlockHashList[0] == "0" {
		if err := os.Remove(path); err != nil {
			log.Println("can't remove local file: ", err)
			return err
		}
		return nil
	}

	data := ""
	for _, hash := range remoteMetaData.BlockHashList {
		var block Block
		if err := client.GetBlock(hash, blockStoreAddr, &block); err != nil {
			log.Println("can't get block: ", err)
		}

		data += string(block.BlockData)
	}
	file.WriteString(data)

	return nil
}
