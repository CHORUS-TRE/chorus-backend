package migration

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	MigrationTableName = "chorus_migrations"
)

func getMigration(path string) (map[string]string, error) {

	files, err := listMigrationFiles(MigrationEmbed, path)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list '%s' migration files", path)
	}

	res := map[string]string{}
	for _, file := range files {
		content, err := readFile(MigrationEmbed, filePath(path, file))
		if err != nil {
			return nil, errors.Wrapf(err, "unable to read embedded file '%s'", file)
		}
		res[removeFileExtension(file)] = content
	}
	return res, nil
}

func readFile(migrationFS embed.FS, file string) (string, error) {
	r, err := migrationFS.Open(file)
	if err != nil {
		return "", err
	}
	defer r.Close()

	contents, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func filePath(storageType, fileName string) string {
	return fmt.Sprintf("%s/%s", storageType, fileName)
}

func removeFileExtension(f string) string {
	extension := filepath.Ext(f)
	return strings.TrimSuffix(f, extension)
}

func listMigrationFiles(migrationFS embed.FS, path string) ([]string, error) {
	files := []string{}

	err := fs.WalkDir(migrationFS, path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, info.Name())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func GetMigration(storageType string) (map[string]string, string, error) {
	switch storageType {
	case POSTGRES:
		migrations, err := getMigration("postgres")
		fmt.Println("migrations", migrations)
		if err != nil {
			return nil, "", err
		}
		return migrations, MigrationTableName, nil
	default:
		return nil, "", fmt.Errorf("unknown storage type %q for chorus migrations", storageType)
	}
}
