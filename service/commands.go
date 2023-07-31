package service

const (
	composerInstallCommand = "composer install --no-scripts --no-interaction"
	composerDumpCommand    = "composer dump-autoload --no-interaction"
	migrateCommand         = "php artisan migrate --force"
	dumpCommand            = "mariadb-dump -u baadbaan -p'secret' baadbaan"
	gitSafeDirectory       = "git config --global --add safe.directory *"
)
