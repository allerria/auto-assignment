# auto-assignment
Тестовое задание на позицию стажера-бэкендера.

Задание можно найти по ссылке https://github.com/avito-tech/auto-backend-trainee-assignment

Код написан на Go. 
В качестве хранилища данных используется Postgres 10.

Из дополнительных возможностей реализованы:
 - простая валидация URL
 - возможность задавать кастомные ссылки

Инструкция по запуску:

Склонировать репозиторий 
```shell script
git clone https://github.com/georgiypetrov/auto-assignment
```
Перейти на папку проекта 
```shell script
cd auto-assignment
```
Запустить при помощи docker-compose
```shell script
docker-compose up --build
```