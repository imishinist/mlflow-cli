package uploader

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

// FileSource implements ArtifactSource for file-based uploads
type FileSource struct {
	filePath     string
	artifactName string
	file         *os.File
	fileInfo     os.FileInfo
}

// NewFileSource creates a new FileSource
func NewFileSource(filePath, artifactName string) (ArtifactSource, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	// Use filename if artifact name is not specified
	if artifactName == "" {
		artifactName = filepath.Base(filePath)
	}

	return &FileSource{
		filePath:     filePath,
		artifactName: artifactName,
		file:         file,
		fileInfo:     fileInfo,
	}, nil
}

func (f *FileSource) Name() string {
	return f.artifactName
}

func (f *FileSource) Reader() io.Reader {
	return f.file
}

func (f *FileSource) Size() int64 {
	return f.fileInfo.Size()
}

func (f *FileSource) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}

// ReaderSource implements ArtifactSource for io.Reader-based uploads
type ReaderSource struct {
	reader       io.Reader
	artifactName string
	size         int64
}

// NewReaderSource creates a new ReaderSource
func NewReaderSource(reader io.Reader, artifactName string, size int64) ArtifactSource {
	return &ReaderSource{
		reader:       reader,
		artifactName: artifactName,
		size:         size,
	}
}

func (r *ReaderSource) Name() string {
	return r.artifactName
}

func (r *ReaderSource) Reader() io.Reader {
	return r.reader
}

func (r *ReaderSource) Size() int64 {
	return r.size
}

func (r *ReaderSource) Close() error {
	// Check if reader implements io.Closer
	if closer, ok := r.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// BytesSource implements ArtifactSource for byte slice uploads
type BytesSource struct {
	data         []byte
	artifactName string
}

// NewBytesSource creates a new BytesSource
func NewBytesSource(data []byte, artifactName string) ArtifactSource {
	return &BytesSource{
		data:         data,
		artifactName: artifactName,
	}
}

func (b *BytesSource) Name() string {
	return b.artifactName
}

func (b *BytesSource) Reader() io.Reader {
	return bytes.NewReader(b.data)
}

func (b *BytesSource) Size() int64 {
	return int64(len(b.data))
}

func (b *BytesSource) Close() error {
	return nil
}
