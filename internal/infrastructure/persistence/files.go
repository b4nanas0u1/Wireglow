package persistence

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// WriteAtomic публикует закрытый файл после записи и fsync; старый файл остаётся при ошибке.
func WriteAtomic(path string, data []byte) (err error) {
	f, err := os.CreateTemp(filepath.Dir(path), ".write-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer func() { _ = f.Close(); _ = os.Remove(name) }()
	if err = f.Chmod(0600); err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	// Root.Rename uses replacement with POSIX semantics on Windows. Combined
	// with delete-sharing readers, a lease can be published while its previous
	// version is being read, without exposing a missing or partial file.
	dir, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	if err = dir.Rename(filepath.Base(name), filepath.Base(path)); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(path))
}

func WriteJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return WriteAtomic(path, append(data, '\n'))
}

func ReadJSON(path string, value any) error {
	f, err := openReadFile(path)
	if err != nil {
		return err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	if !stat.Mode().IsRegular() || stat.Size() > 2*1024*1024 {
		return errors.New("storage: invalid file")
	}
	d := json.NewDecoder(io.LimitReader(f, 2*1024*1024+1))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return errors.New("storage: invalid JSON")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return errors.New("storage: trailing JSON")
	}
	return nil
}
