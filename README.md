# Проект GeoTrack

#### Вспомогательные библиотеки

- [`geo`](../../tree/master/geo) — минимальное описание географических типов данных (точки, круги, полигоны) и методов работы с ними
- [`pairing`](../../tree/master/pairing) — генерация ключей активации для привязки устройства к группе пользователей
- [`mongo`](../../tree/master/mongo) — прослойка для работы с хранилищем данных MongoDB

#### Работа с хранилищем данных

- [`users`](../../tree/master/users) — работа с зарегистрированными пользователями
- [`places`](../../tree/master/places) — работа с координатами мест, определенных пользователями
- [`tracks`](../../tree/master/tracks) — сохранение информации треков и работа с ними

#### Вспомогательные службы

- [`lbs`](../../tree/master/lbs) — данные координат сотовых вышек и вычисление по ним приблизительных координат пользователя
- [`ublox`](../../tree/master/ublox) — получении эфемерид для инициализации GPS-приемников от сервиса U-Blox

#### Сервисы

- [`nats-geotrack`](../../tree/master/nats-geotrack) — сервис для взаимодействия с другими микросервисами через NATS
- [`http-geotrack-api`](../../tree/master/http-geotrack-api) — веб-сервер с поддержкой REST API сервиса
- [`http-geotrack-map`](../../tree/master/http-geotrack-map) — тестовый веб-сервер для отображения координат на карте
