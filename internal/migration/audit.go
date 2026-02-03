package migration

import "fmt"

const (
	AuditMigrationTableName = "audit_migrations"
)

func getAuditMigration(path string) (map[string]string, error) {

	files, err := listMigrationFiles(AuditMigrationEmbed, path)
	if err != nil {
		return nil, fmt.Errorf("unable to list %q migration files: %w", path, err)
	}

	res := map[string]string{}
	for _, file := range files {
		content, err := readFile(AuditMigrationEmbed, filePath(path, file))
		if err != nil {
			return nil, fmt.Errorf("unable to read embedded file %q: %w", file, err)
		}
		res[removeFileExtension(file)] = content
	}
	return res, nil
}

func GetAuditMigration(storageType string) (map[string]string, string, error) {
	switch storageType {
	case POSTGRES:
		migrations, err := getAuditMigration("audit/postgres")
		if err != nil {
			return nil, "", err
		}
		return migrations, AuditMigrationTableName, nil
	default:
		return nil, "", fmt.Errorf("unknown storage type %q for chorus migrations", storageType)
	}
}
