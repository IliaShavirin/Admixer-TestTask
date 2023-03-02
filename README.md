###Пример использования GET запроса:

GET http://localhost:8080/url?ip=192.168.0.1&url_package=1&url_package=2&url_package=4 HTTP/1.1
User-Agent: Fiddler
Host: localhost:8080

###Пример использования POST запроса:

POST http://localhost:8080/url HTTP/1.1
User-Agent: Fiddler
Host: localhost:8080
Content-Type: application/json
Content-Length: 80

{
"request_id": 1,
"url_package": [1, 4, 2],
"ip": "192.168.1.1"
}