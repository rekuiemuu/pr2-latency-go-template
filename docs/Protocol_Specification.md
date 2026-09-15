# Спецификация протокола

Заголовок: `packetType uint8 | sequenceNumber uint16 | payloadSize uint16 | protocolVersion uint16` (7 байт, big-endian, версия 1).

PING: `clientSendTimeUs uint64` (8 байт). PONG: исходная метка клиента и две серверные метки `serverReceiveTimeUs`, `serverSendTimeUs` (24 байта). RTT вычисляется только по часам клиента.

TODO: описать причины отбрасывания недействительных датаграмм.
