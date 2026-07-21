# Проект Docker compose

Проект разделён на три сервиса: entry, hello, world. Сервис entry при получении запроса GET / обращается к сервисам hello и world. Взаимодействие защищено mTLS. Задействованные компоненты используются внутри контейнеров docker, запускаемых внутри каталога testbench командой

```bash
cd testbench/ssl
bash generate.sh
cd ..
docker compose up
```

# mTLS

Для имитации действия неавторизованного пользователя используется отдельный контейнер t. Его попытки обратиться к внутреннему сервису hello неуспешны:

![mTLS fail](screenshots/curl-mtls-fail.png)
![mTLS fail, logs](screenshots/curl-mtls-fail-log.png)

Однако при использовании правильного сертификата:
![mTLS OK](screenshots/curl-mtls-ok.png)

# HTTPS

Для входного HTTPS используется HAProxy в режиме HTTPS.

![HTTPS OK](screenshots/curl-obtain-token.png)

# JWT

Используется низкоуровневая библиотека golang-jwt. При запросе GET / проверяется наличие заголовка X-Token. В случае отсутствия, генерируется и возвращается новый токен сроком действия 1 минута.

![Obtain token](screenshots/curl-obtain-token.png)

В случае наличия, токен проверяется.

![Apply token](screenshots/curl-apply-token.png)

Если срок действия токена истёк, ошибка 403.

![Expired token](screenshots/curl-token-expired.png)
![Expired token log](screenshots/entry-expired-token-log.png)
