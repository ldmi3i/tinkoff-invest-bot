## Описание
<p>
В репозитории представлен бот для автоматизации торговли с использованием Tinkoff API.
Бот предназначен для полностью автоматического ведения торговли.

В целом платформа является расширяемой, но по умолчанию содержит алгоритм, основанный на вычислении оконного среднего 
(а точнее 2х, короткое среднее и длинное средне). 
Соответственно попытка покупки/продажи происходит при "пробое" длинного среднего коротким.
Установлены различные дополнительные ограничения на продажу (чтобы не продавал дешевле, чем купил и т.д.).

В текущем виде стратегия более-менее работает на росте, но при падении с высокой вероятностью акции зависают 
и будут ожидать последующего подъема цены.

На текущий момент полноценно поддерживается торговля только акциями.

Интерфейс представлен с помощью REST API - т.е. можно как использовать приложения (postman, insomnia etc) так и команды curl.
Запросы подразумевают Content-Type=application/json .
</p>

## Запуск приложения
Для быстрого запуска в корне проекта размещен docker-compose.yml.
Перед запуском необходимо заполнить переменную среды `TIN_TOKEN` своим токеном с полным доступом (иначе нельзя будет запустить торговлю).

Также необходимо убедиться не заняты ли порты 5433 и 8017 и при необходимости скорректировать внешний порт в 
docker-compose.yml. </br>
Для запуска можно использовать команду `docker compose up -d` из корневого каталога.
При этом приложение запустится и будет доступно на порту 8017, база данных (Postgres) на порту 5433.
Далее можно использовать его API.

## Работа с историческими данными
Для работы с историей созданы следующие действия:
- Выгрузка данных в базу данных
- Анализ истории с фиксированными параметрами
- Анализ истории с использованием варьирования параметров и поиск лучших

При работе с историей алгоритм и действия в базе данных не сохраняются.

### Выгрузка данных
Перед началом анализа истории необходимо выгрузить данные в базу данных.
Для этого служит метод:</br>
`POST localhost:8017/history/load`
```json5
{
"figis": ["BBG00F9XX7H4", "BBG004S68BH6", "BBG004730N88"], //идентификаторы инструментов для выгрузки
"start_time": 1651634710, //unix time начала интервала
"end_time": 1652867535, //unix time конца интервала
"interval": 1 //тип интервала по Tinkoff API ( 1 - 1 минута, рекомендуется всегда использовать 1)
}
```
Метод разбивает заданный интервал на допустимые по размеру диапазоны и составляет из них полный временной ряд.
Соответственно можно упереться в лимиты API т.к. для интервала 1 минута 
 один запрос выгружает 1 день. С учетом того, что максимальный лимит по запросам в минуту 100-200 в зависимости от грейда, 
получаем соответственно для одного инструмента максимальный диапазон - 100 - 200 дней, 
а если делать выгрузку для 2х инструментов - соответственно 50 - 100 дней и т.д.

**Внимание!** Метод не добавляет данные в базу, а полностью их заменяет на новые во избежание коллизий по временным интервалам.

### Анализ истории с фиксированными параметрами
Имеется возможность провести некоторый анализ алгоритма с фиксированными параметрами алгоритма.
При этом запускается оригинал алгоритма с урезанным логгером чтобы не перегружать лог.
На текущий момент при использовании анализа считается, что все сделки проходят по рыночным текущим ценам.
Т.е. по сути тестируется успешность выбора момента открытия и закрытия по параметрам алгоритма. </br>
`POST localhost:8017/history/analyze`
<details><summary>Описание запроса Click</summary>
<p>
Тело запроса:

```json5
{
	"figis": ["BBG004S68BH6"], //Список идентификаторов figi для торговли
	"strategy": "avr", //Стратегия алгоритма
	"limits": [ //Доступные лимиты по валютам
		{
			"currency": "rub",
			"value": 900
		}
	],
	"params": { //Параметры алгоритма
		"long_dur": "840", //Длительность длинного среднего в секундах
		"short_dur": "790", //Длительность короткого среднего в секундах
		"stop_loss": "3" //Процент просадки цены после которого произойдет продажа по рыночной цене
	}
}
```
Ответ

```json5
{
	"buyOpNum": 4, //Проведено операций покупки (равно числу запросов на покупку)
	"sellOpNum": 3, //Проведено операций продажи
	"curBalance": { //Баланс по валютам (приведенный по стоимости активов на конец периода - т.е. "прибыль")
		"rub": "2.1"
	}
}
```
</p>
</details>

### Анализ истории с использованием варьирования параметров
Есть возможность провести анализ истории с варьированием параметров алгоритма.
При реализации собственного алгоритма в коде для этого нужно реализовать метод разбивки параметров в диапазон.
Для варьирования параметров можно задать границы параметров и шаг. (в случае если шаг не задан - по умолчанию будет взят 1)


`POST localhost:8017/history/analyze/range`
<details><summary>Описание запроса Click</summary>
<p>
Тело запроса

```json5
{
	"figis": ["BBG004S68BH6"], //Список инструментов для анализа
	"strategy": "avr", //Алгоритм для анализа
	"limits": [ //Лимиты по покупкам для алгоритма
		{
			"currency": "rub",
			"value": 900
		}
	],
	"params": { //Параметры варьирования
        "long_dur": "10:100:1500", //Означает с 10 до 1500 с шагом 100 (шаг прибавляется и проводится симуляция пока < верхнего лимита)
        "short_dur": "10:100:1500"
	}
}
```
Ответ

```json5
{
	"bestRes": { //Лучший результат
		"buyOpNum": 7, //Число операций покупки
		"sellOpNum": 6, //Число операций продажи
		"curBalance": { //Баланс по валютам (приведенный по стоимости активов на конец периода - т.е. "прибыль")
			"rub": "9.8"
		}
	},
	"params": { //Параметры алгоритма с лучшим результатом
		"long_dur": "310",
		"short_dur": "110"
	}
}
```
</p>
</details>

**Внимание!** При работе истории выполняется полная симуляция оригинального алгоритма. Оригинальный алгоритм продолжает обработку данных в ожидании
результатов торговых поручений и работает асинхронно. Чтобы все данные не обработались при ожидании ответа от mockTrader в генератор исторических данных добавлена пауза между
сигналами алгоритму в 1мс. Поэтому симуляция может выдавать неверные результаты в случае если ПК перегружен. И по результатам
симуляции с варьированием параметров рекомендуется проверить результат запуском симуляции по фиксированным параметрам.</br>

## Торговля
<p>
Так как алгоритм выставляет лимитные заявки и может их отменять, то оценить работу алгоритма на исторических данных
полностью нельзя (все поручения считаются исполненными по текущей цене в независимости от типа). 
Для более правдоподобной оценки работы алгоритма удобно использовать песочницу.
Со стороны бота нет ограничений на количество параллельно работающих алгоритмов.
Однако с учетом того, что на каждый отдельный алгоритм открывается отдельный stream котировок (и на песочницу и на прод),
то ограничение будет зависеть от грейда (<a href="https://tinkoff.github.io/investAPI/limits/">лимиты</a>).
</p>

### Запуск торговли
Для запуска торгового алгоритма на песочнице используется метод</br>
`POST localhost:8017/trade/sandbox`

После тестирования на песочнице можно запускать алгоритм на прод API.</br>
`POST localhost:8017/trade/prod`

Содержимое запроса одинаковое не зависимо от окружения на котором запускается.

<details><summary>Описание запроса Click</summary>
<p>
Тело запроса

```json5
{
	"figis": [ //Список торгуемых инструментов
		"BBG004S681B4",
		"BBG004S68696",
		"BBG002B9MYC1",
	],
	"strategy": "avr", //Тип алгоритма
	"accountId": "account id", //Идентификатор счета
	"limits": [ //Лимиты по валютам
		{
			"currency": "usd",
			"value": 1000.0
		},
		{
			"currency": "rub",
			"value": 5000
		}
	],
	"params": { //Настройки алгоритма
		"long_dur": "360", //Длина длинного окна среднего в секундах
		"short_dur": "100", //Длина короткого окна среднего в секундах
		"order_expiration": "300", //Время отмены лимитных заявок в секундах, не обязательное, по умолчанию 300
		"stop_loss": "3" //Процент просадки цены после которого произойдет продажа по рыночной цене
	},
	"instrInit": { //Опционально! Исходное количество доступных инструментов (алгоритм по среднем будет сначала искать продажу, а потом перейдет к покупке)
		"instruments": [ //Массив исходных инструментов
		{
			"figi": "BBG004S681B4", //Figi идентификатор инструмента
			"amount": 6, //Количество имеющихся лотов инструмента
			"buyPrice": 150.6 //Цена покупки - будет искаться цена продажи не ниже цены покупки
		}]
	}
}
```
Ответ
```json5
{
	"Info": "Successfully started", //Информация о статусе
	"AlgorithmID": 26 //ID запущенного алгоритма в бд
}
```
</p>
</details>
<p>
При запуске торговли на песочнице или на прод - все запущенные алгоритмы и связанные действия сохраняются в бд, 
которая в случае использования docker-compose.yml доступна на порту 5433.
</p>

**Внимание!** При торговле следует учитывать, что каждый параллельно запущенный алгоритм использует одно stream соединение
по получению котировок. И в связи с этим в зависимости от грейда можно получить ошибку из-за лимитов.
Минимально доступно 2 канала на прод - т.е. 2 параллельно торгующих алгоритма на прод.

### Получение активных алгоритмов
Можно получить текущие активные алгоритмы, торгующие на песочнице и на прод окружениях.

Для получения активных алгоритмов на песочнице:</br>
`GET localhost:8017/trade/algorithms/active/sandbox`

Для получения активных алгоритмов на прод:</br>
`GET localhost:8017/trade/algorithms/active/prod`

### Остановка алгоритма
Имеется возможность остановить торгующий алгоритм.

Для этого можно использовать следующий запрос:</br>
`POST localhost:8017/trade/algorithms/stop?algorithmId={id_of_algorithm}`

Идентификатор активного алгоритма соответственно можно получить из списка активных алгоритмов, 
либо он же - id возвращаемый после старта торговли.

## Статистика
На текущий момент статистика собирается по сохраненным данным действий в базе данных, 
которые формируются по результатам получения статусов торговых поручений.

### Получение текущей статистики по работе алгоритма
<p>
Имеется возможность получить результаты действий по конкретному алгоритму.
Данные включают в себя общее число торговых поручений, число ошибок при выставлении поручения, число отмененных поручений, 
а также текущий баланс по валютам и по инструментам.
</p>

Запрос для получения данных по алгоритму:</br>
`GET 192.168.88.109:8017/stat/algorithm?algorithm_id={id_of_algorithm}`

<details><summary>Описание ответа Click</summary>
<p>
Пример ответа выглядит следующим образом:

```json5
{
	"AlgorithmID": 22, //Id алгоритма
	"SuccessOrders": 4, //Число успешных торговых поручений
	"FailedOrders": 0, //Число торговых поручений, завершившихся с ошибкой
	"CanceledOrders": 0, //Число отмененных торговых поручений
	"MoneyChanges": [ //Изменения по валютам относительно 0
		{
			"Currency": "rub", //Валюта
			"FinalValue": "3.31", //Финальный баланс на окончание работы/текущий момент
			"OperationNum": 4 //Число операций по валюте
		}
	],
	"InstrumentChanges": [ //Изменения по инструментам
		{
			"InstrFigi": "BBG004S68BH6", //Figi инструмента
			"FinalAmount": 0, //Финальный баланс по инструменту
			"OperationNum": 4, //Число операций по инструменту
			"LastLotPrice": "583.71", //Стоимость лота по последней операции
			"FinalMoneyVal": "3.31", //Баланс после последней операции (если завершится на купленном инструменте - баланс будет отрицательный)
			"Currency": "rub" //Валюта инструмента
		}
	]
}
```
</p>
</details>

## Пакеты, настройки и архитектура

### Структура пакетов
В проекте используется слоистая архитектура

#### Пакет internal
`collections` Вспомогательные коллекции для проекта.</br>
`convert` Методы для конвертации quotation и т.д.</br>
`entities` Сущности б А также методы по преобразованию в/из DTO (ссылается на пакет dto)</br>
`dto` DTO для взаимодействия с ботом, содержит пакет `dtotapi` с моделями взаимодействия с Тинькофф API.</br>
`errors` Вспомогательные ошибочные методы и типы.</br>
`env` Методы по обработке переменных среды.
`connections` Содержит пакеты `db` и `grpc` с методами подключения к бд и grpc. Запускает автоматическую миграцию при установлении соединения с бд.</br>
`repository` Слой репозиториев.</br>
`robot` Основное API бота, содержит context через который предполагается работа с API.</br>
`service` Сервисный слой над прямыми запросами к Тинькофф API, скрывает логику общения с API.</br>
`strategy` Содержит фабрику алгоритмов и их реализации в качестве вложенных пакетов.</br>
`tapigen` Сгенерированные proto сервисы grpc.</br>
`tinapi` Обертка над proto сервисами проксирует запросы на API (конвертация dto в запросы Тинькофф API).</br>
`trade` Пакет с трейдерами, которые инкапсулируют логику торговли, отделяя ее от алгоритмов.</br>

#### bot
Содержит контейнер с набором API по управлению ботом.
Можно получить торговые API для торговли на песочнице или проде.
API для работы с историей, API для 

#### web
Реализация интерфейса бота в виде REST API.

### Корневые файлы
`docker-compose.yml` Docker compose файл для развертывания локально сервиса вместе с базой данных.</br>
`Dockerfile` Файл для сборки образа</br>
`gen-proto.sh` Скрипт для генерирования сервисов из proto файлов. Предполагается, что .proto файлы находятся в корне в папке /proto .</br>
`main.go` Стартовый файл приложения.</br>

### Переменные среды
`TIN_TOKEN` Токен доступа к Tinkoff API. Для корректной работы требуется токен с полным доступом. Если не будет задан - приложение не стартует.</br>
`TIN_ADDRESS` Адрес Tinkoff API. По умолчанию "invest-public-api.tinkoff.ru:443".</br>
`DB_USER` Имя пользователя доступа к бд. По умолчанию "postgres".</br>
`DB_PASSWORD` Пароль доступа к бд. По умолчанию "postgres".</br>
`DB_HOST` Хост бд. По умолчанию "localhost".</br>
`DB_PORT` Порт бд. По умолчанию "5432".</br>
`DB_NAME` Имя бд. По умолчанию "invest-bot".</br>
`RETRY_NUM` Количество повторных попыток создания канала котировок в случае его обрыва, ошибки.</br>
`RETRY_INTERVAL_MIN` Интервал между повторными попытками создания канала котировок в случае его обрыва.</br>
`SERVER_PORT` Порт сервера API.</br>
`LOG_FILE_PATH` Путь к файлу с логами. Если не указан - файл не будет писаться.</br>

### Общая логика работы
<p>
Основу приложения составляют "трейдеры" - trade/ProdTrader, trade/SandboxTrader, которые запускаются при выполнении
метода инициализации robot.StartBgTasks() . Трейдеры работают в фоне, им можно добавлять подписчиков - алгоритмы.
Подписка содержит канал для запросов и канал для ответов. При этом входящий канал от алгоритма начинает прослушиваться 
и ожидаться команды на выполнение поручений.<br>
Основная задача "трейдеров" - по запросу алгоритма проводить проверки, выставлять торговые поручения и уведомлять алгоритмы о результатах.
В случае обработки истории - trade/MockTrader создается отдельный для каждого алгоритма и завершает работу после окончания симуляции.
Трейдеры раз в 30 секунд проверяют статусы всех активных торговых поручений и в случае изменения статусов уведомляют алгоритмы.
Также в эти моменты проверяются эстимейты поручений (задаются алгоритмом) и в случае превышения таймаута - поручения отменяются.

Для полноценной работы каждый алгоритм должен иметь фабричный метод создания для окружений прод, песочница, исторические данные.
В текущем варианте реализации алгоритм используется единый - меняются поставщики данных /strategy/avr/(hdataproc/pdataproc).
Для возможности варьирования параметров для алгоритма должна быть создана реализация интерфейса /strategy/stmodel/ParamSplitter
и добавлена в фабрику.

В случае обрыва стрима для алгоритма будут произведены попытки его восстановления в зависимости от настроек (см переменные среды).
По умолчанию будет происходить 3 попытки с интервалом в 3 минуты.

Контекст запроса пробрасывается во все длинные действия и в случае отмены запроса фоновые действия также должны отменяться.
</p>

## Имеющиеся известные баги, недоработки
### По коду и логике
* В случае рестарта приложения алгоритмы не восстанавливают своей работы. Workaround - можно задавать исходные данные при запуске алгоритма.
* В случае остановки алгоритма, обрыва связи либо просто остановки приложения статус isActive в бд обновлен не будет.
* API возвращает ответы со статусами либо 200, либо 500 без учета типа ошибки, которая вернулась.
* gin использует свой логгер, и при записи в файл логи gin туда не попадают. (TODO переключить gin на zap если возможно)
* Было бы неплохо перехватывать сигнал выхода из приложения и выполнять метод пост процессинга (синхронизация логгера, а также всех данных с бд)

### По самому алгоритму
* Слишком жестко заданы лимиты по покупке и продаже и алгоритм может пропускать циклы покупки/продажи в случае отмены лимитной заявки.
* Не учитывается наклон и другие параметры "пробоя", что понижает селективность алгоритма и приводит к ложным срабатываниям.
