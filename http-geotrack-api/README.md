# API

>Это предварительная версия протокола, который все еще находится в разработке. Поэтому в нем возможны любые изменения, в том числе и принципиальные: совместимость пока никто не гарантирует. По мере изменения API, здесь будут публиковаться новые примеры и добавляться описание новых функций.

Данное API старается следовать рекомендациям **RESTful API**, насколько это возможно, если это не принуждает к излишним действиям. 

В качестве формата данных используется _JSON_: как в ответах, так и в запросах. Поэтому настоятельно рекомендуется при запросах указывать в качестве формата данных `"application/json;charset=utf-8"` во избежание недоразумений с кодировками.

Теущая версия использует префикс `/api/v1/` для всех запросов.

Сервер может возвращать следующие коды HTTP в ответ на запросы:

- `200` — запрос успешен, получите ответ
- `201` — запрос принят, ресурс создан (в заголовке будет ссылка на новый ресурс)
- `204` — запрос принят и обработан, ответа не требуется
- `400` — что-то не понравилось в переданных в запросе параметрах
- `401` — требуется авторизация (только при логине)
- `403` — доступ запрещен (нет токена или его время жизни прошло)
- `404` — ресурс с указанным в URL идентификатором не найден
- `405` — используемый HTTP-метод не применим к данному URL
- `415` — тип переданных данных не является форматом JSON
- `500` — внутренняя ошибка сервера
- `501` — данные метод пока не реализован на сервере полностью

Только в случае кода ответа `200`, тело ответа будет содержать данные в формате _JSON_.


## Авторизация

	curl -u login:password http://localhost:8080/api/v1/login

В ответ приходит токен в формате JWT (_mime-type_: `"application/jwt"`):

	eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NDkxMjEyNjMsImdyb3VwIjoiNTQwZGE1NDQtOTgxYy0xMWU1LWEyMmUtMjhjZmU5MWE4NmE3IiwiaWQiOiI1NjVmYTgzZTM0NWVkOTliOTdjNGVhNTYiLCJpc3MiOiJjb20ueHl6cmQudHJhY2tlciJ9.Qpb8vt_BAYalpHJnMKmkjHN3pvxZNEtikhO6qkWXV5I

Пока токен действителен в течение _30_ минут, после чего его нужно получать заново. Это сделано специально для отладки: в дальнейшем время жизни токена может увеличиться до _72_ часов.

При каждой перезагрузке сервера случайным образом меняется ключ, которым подписывается данный токен. Поэтому придется получать этот токен заново: старый будет уже не действителен.

В расшифрованном виде токен содержит время жизни, идентификатор сервиса и минимальную информацию об идентификаторах пользователя:

	{
		"exp": 1449117063,
		"iss": "com.xyzrd.geotracker",
		"id": "565fa83e345ed99b97c4ea56",
		"group": "540da544-981c-11e5-a22e-28cfe91a86a7",
	}

Все обращения к API должны в обязательном порядке использовать полученный токен при запросе. Токен лучше всего передавать в заголовке авторизации HTTP, используя в качестве метода авторизации имя `Bearer`:

	curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/users

В качетсве альтернативы токен можно так же указывать и в URL самого запроса, используя для этого параметр с именем `token`:

	curl http://localhost:8080/api/v1/users?token=<token>

Последний вариант запроса рекомендуется использовать исключительно для отладки, т.к. он обеспечивает меньшую безопасность. В дальнейшем этот вариант передачи токена может быть заблокирован.


## Пользователи

### Получение списка пользователей

	curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/users

Возвращает список всех зарегистрированных пользователей в данной группе, **включая и себя**:

	[
		{
			"ID": "565fa83e345ed99b97c4ea56",
			"Login": "login1",
			"Name": "User #1",
			"Icon": 0
		},
		{
			"ID": "565fa83e345ed99b97c4ea57",
			"Login": "login2",
			"Icon": 1
		}
	]


## Места

### Получение списка всех определенных пользователями мест

	curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/places

Возвращает список всех мест, зарегистрированных для данной группы:

	[
		{
			"ID": "565f3d41345ed988d93cacd3",
			"Name": "Работа",
			"Polygon": [
				[[37.5667, 55.7152], [37.5688, 55.7167], [37.5703, 55.7169], [37.5706, 55.7168],
				 [37.5726, 55.7159], [37.5728, 55.7158], [37.5731, 55.7159], [37.5751, 55.7152],
				 [37.5758, 55.7148], [37.5755, 55.7144], [37.5749, 55.7141], [37.5717, 55.7131],
				 [37.5709, 55.7128], [37.5694, 55.7125], [37.5661, 55.7145], [37.566, 55.7147],
				 [37.5667, 55.7152]]
			]
		},
		{
			"ID": "565f3d41345ed988d93cacd4",
			"Name": "Дом",
			"Circle": {
				"Center": [37.589248, 55.765944],
				"Radius": 200
			}
		},
		{
			"ID": "565f3d41345ed988d93cacd5",
			"Name": "Знаменский монастырь",
			"Polygon": [
				[[37.6256, 55.7522], [37.6304, 55.7523], [37.631, 55.7527], [37.6322, 55.7526],
				 [37.632, 55.7521], [37.6326, 55.7517], [37.6321, 55.7499], [37.6305, 55.7499],
				 [37.6305, 55.7502], [37.6264, 55.7504], [37.6264, 55.75], [37.6254, 55.75],
				 [37.6253, 55.752], [37.6256, 55.7522]]
			]
		}
	]

Места могут быть определены как полигон *или* круг (см. пример выше).


### Получение информации о месте

	curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/places/<place-id>

Возвращает описание места с указанным идентификатором:

	{
		"ID": "565f3d41345ed988d93cacd3",
		"Name": "Работа",
		"Polygon": [
			[[37.5667, 55.7152], [37.5688, 55.7167], [37.5703, 55.7169], [37.5706, 55.7168],
			 [37.5726, 55.7159], [37.5728, 55.7158], [37.5731, 55.7159], [37.5751, 55.7152],
			 [37.5758, 55.7148], [37.5755, 55.7144], [37.5749, 55.7141], [37.5717, 55.7131],
			 [37.5709, 55.7128], [37.5694, 55.7125], [37.5661, 55.7145], [37.566, 55.7147],
			 [37.5667, 55.7152]]
		]
	}


### Добавление нового места

	curl -H "Authorization: Bearer <token>" -X POST http://localhost:8080/api/v1/places \
		-H "Content-Type: application/json" \
		-d $'{
			"Name": "Название места",
			"Circle": {
				"Center": [37.589248, 55.765944],
				"Radius": 200
			}
		}'

Данный запрос создает описание нового места, заданного в виде круга. Аналогичный запрос так же может содержать и описание полигона:

	curl -H "Authorization: Bearer <token>" -X POST http://localhost:8080/api/v1/places \
		-H "Content-Type: application/json" \
		-d $'{
			"Name": "Знаменский монастырь",
			"Polygon": [
				[[37.6256, 55.7522], [37.6304, 55.7523], [37.631, 55.7527], [37.6322, 55.7526],
				 [37.632, 55.7521], [37.6326, 55.7517], [37.6321, 55.7499], [37.6305, 55.7499],
				 [37.6305, 55.7502], [37.6264, 55.7504], [37.6264, 55.75], [37.6254, 55.75],
				 [37.6253, 55.752], [37.6256, 55.7522]]
			]
		}'

Полигон должен описывать замкнутую кривую. Если последняя точка полигона не соответствует в точности начальной точке, то она будет автоматически добавлена. 

Кроме того, полигон может состоять из нескольких кривых. Первая из них будет описывать внешний контур, остальные — изъятия (дыры) в нем. То, что эти многоугольники реально вложенные не проверяется и лежит на совести программы, которая генерирует эти данные: если они будут не корректны, то данное описание места просто никогда не сработает.

При успешном добавлении нового места будет возвращен HTTP-код `201` и в заголовке `Location` будет указан путь и идентификатор этого мета и идентификатор созданного места:

	Location: /api/v1/places/5666410b345ed954eb51bd74

	{"ID": "565f3d41345ed988d93cacd3"}


### Переопределение существующего места

	curl -H "Authorization: Bearer <token>" -X PUT http://localhost:8080/api/v1/places/<place-id> \
		-H "Content-Type: application/json" \
		-d $'{
			"Name": "Название места",
			"Circle": {
				"Center": [37.589248, 55.765944],
				"Radius": 200
			}
		}'

В принципе, запрос полностью аналогичен созданию нового места, только в качестве метода HTTP используется `PUT` и в URL необходимо указать идентификатор уже существующего места, которое и будет заменено на новое описание.

При успешном выполнении возвращается код `200` с пустым телом.


### Удаление информации о месте

	curl -H "Authorization: Bearer <token>" -X DELETE http://localhost:8080/api/v1/places/<place-id>

В качестве HTTP-метода используется `DELETE`, а в URL указывается идентификатор места.

При успешном выполнении возвращается код `200` с пустым телом.


## Устройства и данные о нем

### Получение списка устройств

	curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/devices

Возвращает список всех идентификаторов устройств, зарегистрированных для данной группы:

	[
		"test0123456789",
		"test9876543210"
	]

Сейчас возвращаются только те идентификаторы устройств, по которым есть данные трекинга. В дальнейшем этот механизм будет изменен и будут возвращаться идентификаторы всех зарегистрированных устройств.


### Получение истории гео-данных устройства

	curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/devices/test0123456789/tracks

Возвращает список треков, полученных с данного устройства:

	[
		{
			"ID": "565fa11de401cc50ba1284d1",
			"Time": "2015-12-03T00:59:09.645+03:00",
			"Location": [37.68454428212746, 55.72747138508032]
			"Accuracy": 1000.1,
			"Method": 4,
			"Power": 92
		},
		{
			"ID": "565fa11de401cc50ba1284d0",
			"Time": "2015-12-03T00:57:59.8+03:00",
			"Location": [37.68419267639476, 55.72766069897311],
			"Accuracy": 30,
			"Method": 1,
			"Power": 92
		},
		{
			"ID": "565fa11de401cc50ba1284cf",
			"Time": "2015-12-03T00:54:24.267+03:00",
			"Location": [37.683480852345724, 55.7280431528877],
			"Accuracy": 641.5,
			"Method": 6,
			"Power": 93
		}
	]

`Location` описывает координаты точки в формате долгота, широта.

`Accuracy` содержит радиус погрешности вычисления координат в метрах.

`Method` — указывает тип полученных координат:

- `0` — неизвестный метод
- `1` — координаты получены через GPS
- `2` — координаты получены по точкам Wi-Fi
- `4` — координаты получены методом вычисления по LBS
- `6` — LBS+Wi-Fi

Если `Method` не указан, то считается, что его значение `0`.


#### Постраничная навигация

Для навигации по страницам истории можно указывать в параметры последний полученный идентификатор: в этом случае вывод начнется с более старых данных, чем в указанном идентификаторе.

	curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/devices/test0123456789/tracks?lastid=565fa11de401cc50ba1284c8

Так же можно задать количество возвращаемых данных, задав параметр `limit`:

	curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/devices/test0123456789/tracks?limit=500

По умолчанию (если не указано) лимит возвращаемых данных установлен в 200.


### Добавление новых данных о треках устройства

	curl -H "Authorization: Bearer <token>" -X POST http://localhost:8080/api/v1/devices/test0123456789/tracks \
		-H "Content-Type: application/json" \
		-d $'[
		{
			"Time": "2015-12-03T00:59:09.645+03:00",
			"Location": [37.68454428212746, 55.72747138508032]
			"Accuracy": 1000.1,
			"Method": 4,
			"Power": 92
		},
		{
			"Time": "2015-12-03T00:57:59.8+03:00",
			"Location": [37.68419267639476, 55.72766069897311],
			"Accuracy": 30,
			"Method": 1,
			"Power": 92
		},
		{
			"Time": "2015-12-03T00:54:24.267+03:00",
			"Location": [37.683480852345724, 55.7280431528877],
			"Accuracy": 641.5,
			"Method": 6,
			"Power": 93
		}
	]'

`Method` — указывает тип полученных координат:

- `0` — неизвестный метод
- `1` — координаты получены через GPS
- `2` — координаты получены по точкам Wi-Fi
- `4` — координаты получены методом вычисления по LBS
- `6` — LBS+Wi-Fi

Если `Method` не указан, то считается, что его значение `0`.

**ВНИМАНИЕ!** При приеме этих данных никакой проверки на их дублирование на сервере не происходит. Поэтому повторная их публикация просто приведет к дублированию информации.


### Получение истории датчиков устройства

	curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/devices/test0123456789/sensors

Возвращает список изменений значений датчиков, полученных с данного устройства:

	[
		{
			"ID": "565fa11de401cc50ba1284d1",
			"Time": "2015-12-03T00:59:09.645+03:00",
			"Data": {
        		"IsBraceletOn": false
    		}
		},
		{
			"ID": "565fa11de401cc50ba1284d0",
			"Time": "2015-12-03T00:57:59.8+03:00",
			"Data": {
        		"NumberOfSteps": 14
    		}
   		}
	]

#### Постраничная навигация

Для навигации по страницам истории можно указывать в параметры последний полученный идентификатор: в этом случае вывод начнется с более старых данных, чем в указанном идентификаторе.

	curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/devices/test0123456789/sensors?lastid=565fa11de401cc50ba1284c8

Так же можно задать количество возвращаемых данных, задав параметр `limit`:

	curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/devices/test0123456789/sensors?limit=500

По умолчанию (если не указано) лимит возвращаемых данных установлен в 200.


### Добавление новых данных о сенсорах устройства

	curl -H "Authorization: Bearer <token>" -X POST http://localhost:8080/api/v1/devices/test0123456789/sensors \
		-H "Content-Type: application/json" \
		-d $'[
		{
			"Time": "2015-12-03T00:59:09.645+03:00",
			"Data": {
        		"IsBraceletOn": false
    		}
		},
		{
			"Time": "2015-12-03T00:57:59.8+03:00",
			"Data": {
        		"NumberOfSteps": 14
    		}
   		}
	]'

<!--
## Поддержка push

### Регистрация токена устройства

	curl -H "Authorization: Bearer <token>" -d "deviceid=<deviceID>&token=<token_string>" http://localhost:8080/api/v1/register/apns

Регистрирует указанный токен устройства в хранилище, чтобы можно было отправлять на это устройства сообщения. Используется метод HTTP POST. В качестве параметров передаются уникальный идентификатор устройства и сам токен устройства, полученный от сервиса Apple для получения push-уведомлений. Так же в URL запроса указывается тип: Apple Push Notification — `apns`, Google Cloud Messaging — `gcm`.


### Удаление токена устройства

Для удаление токена устройства используется запрос с методом HTTP DELETE:

	curl -H "Authorization: Bearer <token>" -X DELETE http://localhost:8080/api/v1/register/apns/<token>

-->
