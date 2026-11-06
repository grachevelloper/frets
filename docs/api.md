# API frets

Префикс `/api/v1`. JSON использует camelCase. Ошибка имеет форму
`{"code":"..."}`. Изменяющим запросам нужен заголовок `Origin`, совпадающий
с `PUBLIC_ORIGIN`.

| Метод и путь | Результат |
| --- | --- |
| GET /auth/login/{provider} | 302 на OAuth. Допустимы `yandex` и `google`. Необязательный `next` возвращает к маршруту песни. |
| GET /auth/callback/{provider} | Одноразовая проверка состояния, обмен кода с PKCE, профиль провайдера, cookie и 302. |
| POST /auth/local | 204 и cookie локального профиля. Доступен только при `LOCAL_LOGIN=true` и локальном origin. |
| GET /me | 200 с `id`, `nickname`, `email`, `createdAt`. |
| POST /auth/logout | 204, отзыв сессии и очистка cookie. |
| GET /songs | 200 со списком песен. Поля `id`, `title`, `createdAt`. |
| POST /songs | Multipart `title`, `sha256`, `file` и заголовок `Idempotency-Key`. Максимум 16 МиБ. 201 при создании и 200 при повторе. |
| GET /songs/{songId} | 200 с песней и массивом `fragments`. |
| GET /songs/{songId}/file | 200 с исходными байтами и `Cache-Control: no-store`. |
| PUT /songs/{songId}/fragments/{fragmentId} | UUID задаёт клиент. JSON `title`, `startBar`, `endBar`. 201 при создании, 200 при совпадающем повторе, 409 при конфликте. |
| DELETE /songs/{songId} | 202 при незавершённом удалении, 204 после завершения, включая повтор. |

Фрагмент содержит `id`, `songId`, `title`, `startBar`, `endBar`.
Отсутствующий или чужой ресурс даёт `404 not_found`. Сессия, которую нельзя
использовать, даёт `401 unauthorized`, а ошибка зависимости или таймаут —
`503 unavailable`. `409 upload_conflict` означает несовпадение параметров,
`409 upload_in_progress` — активную попытку, `409 upload_expired` — истёкший
срок. В последних двух случаях повторите отправку с тем же ключом.

Закрытый `POST /internal/sessions/validate` получает cookie сессии и
проверяет `Authorization: Bearer …` по общему секрету. При успехе возвращает
`{"userId":"…"}`. HTTPS-режим включает атрибуты cookie Secure, HttpOnly
и SameSite=Lax.

Email профиля — обязательная строка. OAuth-вход без корректного email отклоняется.
Повторный вход обновляет email существующего профиля. Старые сессии профилей
без email возвращают 401 и требуют повторного входа. Локальный тестовый
профиль использует служебный адрес local@frets.invalid.
