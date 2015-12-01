# Сервис для работы с NATS

> По-хорошему, конечно, в дальнейшем данный сервис будет разбит на множество микросервисов, но сейчас мне было удобнее объединить все в один.

Строки соединения с MongoDB и NATS вынесены в параметры и могут изменяться при запуске приложения. Аналогично, в параметры вынесен токен для работы с сервером U-Blox. 

	Usage of ./nats-service:
	  -mongodb string
	    	MongoDB connection URL (default "mongodb://localhost/watch")
	  -nats string
	    	NATS connection URL (default "nats://localhost:4222")
	  -ublox string
	    	U-Blox token (default "I6...VyrA")

Для MongoDB крайне рекомендуется в URL так же указывать название базы данных. Иначе будет использоваться база по умолчанию.

В качестве формата данных для обмена информации внутри NATS все микросервисы используют JSON. Проще всего работать с NATS через его `nats.EncodedConn`, что позволит избежать всех "танцев"" с кодирование/декодированием данных при передаче.

Ниже описаны темы (_subjects_), на которые подписан данный сервер, и параметры запросов и ответов к ним. В файле <main_test.go> можно посмотреть примеры взаимодействия с данным сервером.


### `lbs`

Данный микросервис отвечает за вычисление приблизительных координат на основании данных о вышках сотовой связи.

На входе он получает строку в формате LBS:

	"864078-35827-010003698-fa-2-1e50-772a-95-1e50-773c-a6-1e50-7728-a1-1e50-7725-92-1e50-772d-90-1e50-7741-90-1e50-7726-88"

После разбора возвращает вычисленные координаты (долгота, широта):

	[37.712766,55.735922]

В случае ошибки, или если не смогли разобрать строку с LBS, возвращается `nil`.

Для корректной работы сервиса необходимо импортировать базу с данными сотовых вышек. Для этого можно использовать приложение [`lbs-import`](../lbs/lbs-import).


### `ublox`

Возвращает данные для инициализации GPS-приемника на основании предположительных координат.

На входе получает координаты предполагаемой точки (долгота, широта):

	[37.712766,55.735922]

На выходе отдаются бинарные данные (в JSON они кодируются как `base64`):

	"tWILATAAPENlFnhyNSEAAAAAgJaYAAAAUQdRxnoIjgUKABAnAAAAAAAAAAAAAAAAAAAjAAAADFa1YgsxaAACAAAAJYssAgBQ1ACXzVECIsRJAqQTjwLU4MICKCNlAAcAAAJ3P04AnvtlALdaNQKfE+YAByL8AMx5rwKhgBQAXGYNAHwoIwIliAAAZxX7ACbO/wBRo18CphMWAO20BgKuqf8ARg9lAMfetWILMWgAAwAAACWLLAIAUNQAl81RAiLESQKkE48CBODCAigjHgDZ/wACaBIBAK4CHgAGqjQCwBnfAABGAgI+1TMCoSYPAPlgDABnKCMCUuv/AONcGgAn9/8CsnwRAoSzHQDRfg8Aeaf/ArkHHgDSq7ViCzFoAAYAAAAliywCAFDUAJfNUQIixEkCpBOPAgngwgIoI2UCOgAAAPWnDQLs+mUA+MYwAv2u/gAAo/sCt8sbAKF3FQJSuwwCfigjACcAAAD8S1UAJwIAAq4TPAJ7CBcAjlHKAgeq/wK8CmUCZY+1YgsxaAAHAAAAJYssAgBR1ACXzVkAIsRJAKQTjwDo4MIAKCM6Avb/AABkwz4ABAg6AIWKLADZyq8ABNEGApDHnAKhWxMCOIkNAn8oIwCo3/8AJkloAifZ/wKrlnQAksAZAtAyxgCTq/8Cmf06AvZJtWILMWgACAAAACWLLAIAUNQAl81RAiLESQKkE48CCuDCAigjLAD0/wAAnhv+Aib1LAAxvTECrpZsAgBv9gDsz84AoQILAsXnDAJ+KCMA/AwAAJzKjQAn5/8CQnAuAr26IgJVcdAC9aX/AoX8LABRW7ViCzFoAAkAAAAliywCAFDUAJfNUQIixEkCpBOPAgLgwgIoI0gCPgAAAqyHAgBZAEgA8Q40AGnfkAIALQAAKao6AKGQDwJaeA0CfSgjAnwGAAAvYoUAJgAAADNj+ABTIx0CEKjjABSn/wKE7UgCVxS1YgsxaAAQAAAAJYssAgBQ1ACXzVECIsRJAqQTjwLq4MICJyMHAh8AAAKx6vcCvgQHACG2KgKqN6MCBFAEAkabNAKhKQgC61UNADUnIwDU+v8AnwriAig0AAI4jlwCDL4nAjcywwKSp/8C7wUHAhM8tWILMWgAFwAAACWLLAIAUNQAl81RAiLESQKkE48C1eDCAigjNgDu/wACbLPsAFwANgDHzjUC1qpVAgUPAAKgslIAoe8OADkbDgJ/KCMAfF0AANc6zQAmsf8CKEeaAJUaHQC9FjkCCqf/AArwNgDuJLViCzFoABoAAAAliywCAFDUAJfNUQIixEkCpBOPAhDgwgIoI0AAb/8AACsN7QDmBUAAUCsyAB8rgAIAJgUCbZs6AKHRBwC4Zw0CdSgjANEMAABoWOQAJ+//AkI/IgDxSSYCct+2AnWk/wKZCUAA3ka1YgsxaAAbAAAAJYssAgBQ1ACXzVECIsRJAqQTjwIE4MICKCNoAB0AAAAPSQUC6fRoAPhbMAB+rrAAAVD2APi9cAChuAoAD7INAnwoIwL8DgAASxebACcWAAABxIICCtMjAA8UtgL6pf8AzvtoAg3KtWILMWgAHgAAACWLLAIAUNQAl81RAiLESQKkE48CB+DCAigjLAIwAAACS3IIAHkILACDozAACnCQAABLBwAcVdoAod0RAmFiDQJ8KCMCqv7/AHxWTwAmCAAArP3fAoCTGgI4HAUAQqj/AgD9LAABNrViCzFoACAAAAAliywCAFDUAJfNUQIixEkCpBOPAvngwgIoIzIAaQAAAnRcBABFAjIAls43AI7GjAAFvQEC20DTAKGqDgLdKw0AfSgjAlWM/wAsm1cAJnv/AN+CkAAHWR0CCNWKAnSm/wBbBjICVVm1YgsCSAD3/f//AAAAAAAAEL4AAAAAAADovADgBABRBxEAOwcDABEAAAAAAIAyAACAsgAAgLMAAAA0AADsRwAAYMgAAIDHAABQSQcAAADppA=="

В случае ошибки возвращается `nil`.

Сервис может обращаться к внешнему серверу. Поэтому рекомендуется поставить время ожидания ответа не очень маленькое.


### `imei`

Микросервис идентификации (авторизации) браслетов по их уникальным идентификаторам.

> На данный момент это не реальная реализация кода, а предварительная — исключительно для тестирования.

На вход поступает уникальный идентификатор браслета в виде строки:

	"12345678901234"

На выходе — уникальный идентификатор группы пользователей, к которой привязан данный браслет, и список идентификаторов пользователей (включая номер иконки через дефис):

	{
		"GroupID": "206c591e-a151-4540-bdcb-00c35f95792b",
		"Users": [
			"565c7579345ed92c8277640d-0",
			"565c7579345ed92c8277640e-1",
			"565c7579345ed92c8277640f-2",
			"565c7579345ed92c82776410-3",
			"565c7579345ed92c82776411-4"
		]
	}

Если браслет с таким идентификатором не зарегистрирован, то в ответ возвращается `nil`.


### `tracks`

Принимает данные с трекингом и сохраняет их в хранилище.

На входе данные для трекинга (координаты точки в формате [долгота, широта]):

	{
		"GroupID": "206c591e-a151-4540-bdcb-00c35f95792b",
		"DeviceID": "12345678901234",
		"Time": "2015-11-30T18:32:25.237+03:00",
		"Point": [37.589248,55.765944]
	}

Ответа никакого не возвращается.


