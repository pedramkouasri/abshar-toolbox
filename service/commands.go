package service

const (
	composerInstallCommand = "composer install --no-scripts --no-interaction"
	composerDumpCommand    = "composer dump-autoload --no-interaction"
	migrateCommand         = "php artisan migrate --force"
	viewClearCommand       = "php artisan view:clear"
	configCacheCommand     = "php artisan config:cache"
	removeConfig           = "rm -rf bootstrap/cache/*"
	sqlDumpCommand         = "mariadb-dump --user=%s --password=%s --host=%s --port=%s  %s"
	gitSafeDirectory       = "git config --global --add safe.directory *"
)
