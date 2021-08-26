# janus_stress_test
Meetecho Janus WebRtc stress test

Test WebRtc connections to Janus Stream Plugin from other Janus Server (early worked version alpha)

test WebRtc connection from Janus server (plugin streaming) to Janus client (plugin videoroom)

Server - тестируемый сервер
Client - клиент, который будет подключаться к серверу и смотреть стримы

Со стороны сервера идет вещание потоков через janus.plugin.streaming
Со стороны клиента идет получение потоков через janus.plugin.videoroom

Janus1 (videoroom) -> webrtc connect to -> Janus2 (streaming) -> Janus2 watch stream -> Janus2 webrtc streams to -> Janus1 (videoroom)

Идея: с одной стороны стоит Janus сервер с плагином videoroom, который выступает в качестве клиента и соединяется с другим Janus сервером, который необходимо протестировать и на котором стоит плагин streaming.

start:
configure hardcoded params in main.go
go run ./