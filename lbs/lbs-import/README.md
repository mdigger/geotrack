# Импорт данных в базу

Данная программа позволяет импортировать данные о координатах сотовых вышек, которые потом используются для вычисления координат для LBS.

	Import LBS database data
	./lbs-import [-params] datafile.csv
	  -country string
	    	filter for country (comma separated) (default "250")
	  -minsample int
	    	filter for min samples count
	  -mongo string
	    	mongoDB connection URL (default "mongodb://localhost/watch")
	  -radio string
	    	filter for radio (comma separated) (default "gsm")

Т.к. импорт данных занимает некоторое время, в целях отладки можно указать фильтры, которые будут применены при импорте данных. В этом случае база будет содержать только те данные, которые подпадают под данный фильтр. В качестве фильтра можно указывать список типов радио-вышек и кодов стран, разделенные запятой, а так же количество подтверждений данных.

Данные в формате CSV можно загрузить с сервера <http://opencellid.org/#action=database.downloadDatabase>. Для загрузки необходимо будет использовать API key, который необходимо будет получить.

Кроме этого, базу можно скачать с сервера [Mozilla Locator](https://location.services.mozilla.com/downloads) — эти данные несколько больше и актуальнее, чем предлагает OpenCellId.