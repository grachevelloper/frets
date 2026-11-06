# Go-скиллы для frets

Оригинальные инструкции и справочные
файлы сохранены без изменений.

| Скилл | Источник | Для чего |
| --- | --- | --- |
| golang-context | [samber/cc-skills-golang](https://github.com/samber/cc-skills-golang/tree/main/skills/golang-context) | Отмена операций, сроки выполнения и передача context |
| golang-error-handling | [samber/cc-skills-golang](https://github.com/samber/cc-skills-golang/tree/main/skills/golang-error-handling) | Сохранение причин ошибок и обработка сбоев |
| golang-security | [samber/cc-skills-golang](https://github.com/samber/cc-skills-golang/tree/main/skills/golang-security) | Проверка входа, доступа, файлов и сетевых запросов |
| golang-testing | [affaan-m/ecc](https://github.com/affaan-m/ecc/tree/main/skills/golang-testing) | Тестирование поведения и исправлений ошибок |

Правила корневого [AGENTS.md](../../AGENTS.md) имеют приоритет над общими
рекомендациями этих скиллов. В частности:

- Версия Go определяется `backend/go.mod`; примеры для более новых версий
  требуют проверки совместимости.
- Сохраняются существующие слои, PostgreSQL, pgx и механизм миграций.
- Репозитории проверяются на настоящей PostgreSQL. Примеры с SQLite
  не заменяют эти проверки.
- Новые зависимости и инструменты добавляются при конкретной необходимости.
  Упоминание samber/oops, testify или сканера в скилле само по себе
  не является причиной менять зависимости frets.
- При изменении зависимостей обновляются `go.mod`, `go.sum` и `vendor`.
- Правила Git, рабочих каталогов и согласования изменений берутся из AGENTS.md.

Ссылки на другие скиллы в оригиналах оставлены как справочные.
Остальные скиллы из этих коллекций не установлены.
