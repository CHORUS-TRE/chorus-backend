package migration

import (
	"fmt"
)

const (
	MigrationTableName = "chorus_migrations"
)

func getMigration(path string) (map[string]string, error) {

	files, err := listMigrationFiles(ChorusMigrationEmbed, path)
	if err != nil {
		return nil, fmt.Errorf("unable to list %q migration files: %w", path, err)
	}

	res := map[string]string{}
	for _, file := range files {
		content, err := readFile(ChorusMigrationEmbed, filePath(path, file))
		if err != nil {
			return nil, fmt.Errorf("unable to read embedded file %q: %w", file, err)
		}
		res[removeFileExtension(file)] = content
	}
	return res, nil
}

func GetMigration(storageType string) (map[string]string, string, error) {
	switch storageType {
	case POSTGRES:
		migrations, err := getMigration("chorus/postgres")
		if err != nil {
			return nil, "", err
		}
		return migrations, MigrationTableName, nil
	default:
		return nil, "", fmt.Errorf("unknown storage type %q for chorus migrations", storageType)
	}
}
