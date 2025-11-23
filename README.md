# Запуск
Запуск осуществляется командой docker-compose up --build.

Сервер поднимается на порту 8080

# Решение
Из исходной openapi спецификации был сгенерирован [базовый код](./gen/api/api.gen.go), используя oapi-codegen. 
В файле [server.go](./internal/server/server.go) реализованы хендлеры, разбирающие http запрос и перенаправляющие его в 
прослоку для работы с базой данных. Работа с БД реализована в [repository.go](./internal/repository.go) с ипользованием 
gorm. Используемые gorm сущности созданы в файле [entities.go](internal/repository/entities.go).

В файле [wrappers.go](internal/wrappers/wrappers.go) реализованы обёртки для корректного парсинга в json и назад.

# Условие
Исходное условие можно найти [здесь](task.md)
