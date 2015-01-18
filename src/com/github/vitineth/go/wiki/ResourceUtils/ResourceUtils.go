package ResourceUtils

import (
	"os"
	"bufio"
	"strconv"
)

const (
	DATA_PATH_BASE = "data/"
)

type MetaData struct {
	PageCreationDate string
	PageCreationTime string
	LastSaveDate     string
	LastSaveTime     string
	Author           string
	Views            int

}

func AddPageViewToMetadata(meta *MetaData) (newMeta *MetaData) {
	var retMeta *MetaData = &MetaData{
		PageCreationDate: meta.PageCreationDate,
		PageCreationTime: meta.PageCreationTime,
		LastSaveDate: meta.LastSaveDate,
		LastSaveTime: meta.LastSaveTime,
		Author: meta.Author,
		Views: meta.Views + 1,
	}
	return retMeta
}

func SaveFileMetadata(meta *MetaData, pageName string) (err error) {
	filename := "data/meta/" + pageName + ".txt"

	error := os.Remove(filename)
	if error != nil {return error}

	writer, error := os.Create(filename)
	if error != nil {return error}

	bufWriter := bufio.NewWriter(writer)

	_, error = bufWriter.WriteString(meta.PageCreationDate+"\n")
	if error != nil {return error}

	_, error = bufWriter.WriteString(meta.PageCreationTime+"\n")
	if error != nil {return error}

	_, error = bufWriter.WriteString(meta.LastSaveDate+"\n")
	if error != nil {return error}

	_, error = bufWriter.WriteString(meta.LastSaveTime+"\n")
	if error != nil {return error}

	_, error = bufWriter.WriteString(meta.Author+"\n")
	if error != nil {return error}

	_, error = bufWriter.WriteString(strconv.Itoa(meta.Views)+"\n")
	if error != nil {return error}

	bufWriter.Flush()

	return nil
}

func LoadFileMetadata(pageName string) (metadata *MetaData, err error) {
	filename := "data/meta/" + pageName + ".txt"
	reader, error := os.Open(filename)
	if error != nil {
		return &MetaData{}, error
	}
	bufReader := bufio.NewReader(reader)


	creationDate, _, error := bufReader.ReadLine()
	if error != nil {return &MetaData{}, error}

	creationTime, _, error := bufReader.ReadLine()
	if error != nil {return &MetaData{}, error}

	lastEdited, _, error := bufReader.ReadLine()
	if error != nil {return &MetaData{}, error}

	lastEditTime, _, error := bufReader.ReadLine()
	if error != nil {return &MetaData{}, error}

	author, _, error := bufReader.ReadLine()
	if error != nil {return &MetaData{}, error}

	views, _, error := bufReader.ReadLine()
	if error != nil {return &MetaData{}, error}

	iViews, err := strconv.Atoi(string(views))
	if error != nil {return &MetaData{}, error}

	return &MetaData{PageCreationDate: string(creationDate), PageCreationTime: string(creationTime), LastSaveDate: string(lastEdited), LastSaveTime: string(lastEditTime), Author: string(author), Views: iViews}, nil
}

