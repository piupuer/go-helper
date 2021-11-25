package req

import (
	"fmt"
	"github.com/golang-module/carbon"
	"github.com/pkg/errors"
	"regexp"
)

const (
	// chunk tmp path
	ChunkTmpPath = "chunks"
)

type FilePartInfo struct {
	SaveDir                 string `json:"-"`
	SingleMaxSize           int64  `json:"-"`
	CurrentSize             *uint  `json:"-"`
	CurrentCheckChunkNumber uint   `json:"-"`
	// uploaded block numbers
	Uploaded []uint `json:"uploaded"`
	// whether transfer complete
	Complete    bool   `json:"complete"`
	ChunkNumber uint   `json:"chunkNumber" form:"chunkNumber"`
	ChunkSize   uint   `json:"chunkSize" form:"chunkSize"`
	TotalSize   uint   `json:"totalSize" form:"totalSize"`
	Identifier  string `json:"identifier" form:"identifier"`
	Filename    string `json:"filename" form:"filename"`
}

// Remove special characters
func (pt *FilePartInfo) CleanIdentifier() string {
	re, _ := regexp.Compile("[^0-9A-Za-z_-]")
	return re.ReplaceAllString(pt.Identifier, "")
}

func (pt *FilePartInfo) GetTotalChunk() uint {
	// The remainder will be merged with the last block instead of + 1
	// 105 / 25 => 4 chunk
	// 100 / 25 => 4 chunk
	// 99 / 25 => 3 chunk
	// 24 / 25 => 1 chunk
	if pt.ChunkSize > 0 && pt.TotalSize > pt.ChunkSize {
		return pt.TotalSize / pt.ChunkSize
	}
	return 1
}

func (pt *FilePartInfo) GetChunkFilename(chunkNumber uint) string {
	identifier := pt.CleanIdentifier()
	return fmt.Sprintf(
		"%s/%s/%s/uploader-%s/chunk%d",
		pt.SaveDir,
		carbon.Now().ToDateString(),
		ChunkTmpPath,
		identifier,
		chunkNumber,
	)
}

func (pt *FilePartInfo) GetChunkFilenameWithoutChunkNumber() string {
	identifier := pt.CleanIdentifier()
	return fmt.Sprintf(
		"%s/%s/%s/uploader-%s/chunk",
		pt.SaveDir,
		carbon.Now().ToDateString(),
		ChunkTmpPath,
		identifier,
	)
}

func (pt *FilePartInfo) GetUploadRootPath() string {
	return fmt.Sprintf(
		"%s/%s",
		pt.SaveDir,
		carbon.Now().ToDateString(),
	)
}

func (pt *FilePartInfo) GetChunkRootPath() string {
	identifier := pt.CleanIdentifier()
	return fmt.Sprintf(
		"%s/%s/%s/uploader-%s",
		pt.SaveDir,
		carbon.Now().ToDateString(),
		ChunkTmpPath,
		identifier,
	)
}

func (pt *FilePartInfo) Validate() error {
	filePart := pt
	if filePart == nil {
		return errors.WithStack(fmt.Errorf("file params invalid"))
	}
	if filePart.ChunkNumber == 0 ||
		filePart.ChunkSize == 0 ||
		filePart.TotalSize == 0 ||
		filePart.Identifier == "" ||
		filePart.Filename == "" {
		return errors.WithStack(fmt.Errorf("file name or file size invalid"))
	}

	totalChunk := filePart.GetTotalChunk()
	if filePart.ChunkNumber > totalChunk {
		return errors.WithStack(fmt.Errorf("file chunk number invalid"))
	}

	if filePart.CurrentSize != nil {
		if int64(*filePart.CurrentSize) > int64(pt.SingleMaxSize)<<20 {
			return errors.WithStack(fmt.Errorf("the file size exceeds the maximum: %dMB, current: %dB", pt.SingleMaxSize, int64(*filePart.CurrentSize)))
		}

		if filePart.ChunkNumber < totalChunk && *filePart.CurrentSize != filePart.ChunkSize {
			return errors.WithStack(fmt.Errorf("inconsistent file block size: [%d:%d]", filePart.CurrentSize, filePart.ChunkSize))
		}

		if totalChunk > 1 &&
			filePart.ChunkNumber == totalChunk &&
			*filePart.CurrentSize != filePart.TotalSize%filePart.ChunkSize+filePart.ChunkSize {
			return errors.WithStack(fmt.Errorf("inconsistent file last block size: [%d:%d]", filePart.CurrentSize, filePart.TotalSize%filePart.ChunkSize+filePart.ChunkSize))
		}
		if totalChunk == 1 &&
			*filePart.CurrentSize != filePart.TotalSize {
			return errors.WithStack(fmt.Errorf("inconsistent file first block size: [%d:%d]", filePart.CurrentSize, filePart.TotalSize))
		}
	}
	return nil
}
