package service

const (
	composerInstallCommand = "composer install --no-scripts --no-interaction"
	composerDumpCommand    = "composer dump-autoload --no-interaction"
	migrateCommand         = "php artisan migrate --force"
	sqlDumpCommand         = "mariadb-dump --user=%s --password=%s --host=%s --port=%s  %s"
	gitSafeDirectory       = "git config --global --add safe.directory *"
)
