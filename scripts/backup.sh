#!/usr/bin/env bash
# Бэкап PostgreSQL в docker-compose окружении.
#
# Использование:
#   ./scripts/backup.sh                 # разовый бэкап в ./backups
#   ./scripts/backup.sh --restore FILE  # восстановить из дампа
#   ./scripts/backup.sh --prune 7       # удалить бэкапы старше 7 дней
#
# Автоматизация (cron, каждый день в 03:00):
#   0 3 * * * cd /path/to/project && ./scripts/backup.sh >> /var/log/marketplace-backup.log 2>&1
#
# Рекомендации по хранению:
#   - Локально: минимум 7 ежедневных + 4 еженедельных (см. --prune)
#   - Обязательно копировать дампы в удалённое хранилище (S3, другой сервер)
#   - Шифровать дампы перед отправкой в облако (gpg/age), т.к. внутри персональные данные
#   - Проверять восстановление хотя бы раз в месяц (--restore на тестовую БД)

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-./backups}"
COMPOSE="docker compose"
DB_CONTAINER="db"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-marketplace}"

mkdir -p "$BACKUP_DIR"

case "${1:-}" in
  --restore)
    FILE="${2:?Usage: $0 --restore FILE}"
    echo "Restoring $FILE into $DB_NAME..."
    $COMPOSE exec -T "$DB_CONTAINER" pg_restore \
      --clean --if-exists --no-owner \
      -U "$DB_USER" -d "$DB_NAME" < "$FILE"
    echo "Restore done."
    ;;
  --prune)
    DAYS="${2:-7}"
    echo "Removing backups older than $DAYS days..."
    find "$BACKUP_DIR" -name '*.dump' -mtime "+$DAYS" -delete
    echo "Prune done."
    ;;
  *)
    TS="$(date +%Y%m%d_%H%M%S)"
    FILE="$BACKUP_DIR/marketplace_${TS}.dump"
    echo "Creating backup: $FILE"
    $COMPOSE exec -T "$DB_CONTAINER" pg_dump \
      -Fc -U "$DB_USER" "$DB_NAME" > "$FILE"
    echo "Backup saved: $FILE ($(du -h "$FILE" | cut -f1))"
    ;;
esac