package service

const (
	composerInstallCommand = "composer install --no-scripts"
	composerDumpCommand    = "composer dump-autoload"
	migrateCommand         = "php artisan migrate --force"
	dumpCommand            = "mariadb-dump -u baadbaan -p'secret' baadbaan"
)
