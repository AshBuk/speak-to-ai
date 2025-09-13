Конкретные улучшения (минимальные правки, high‑impact)

✅ Выполнено:
- Типизация GetConfig() interface{} → *config.Config
- Унификация логгера (websocket/server.go, audio/factory)
- Исправление уровней логов DBus-провайдера
- SRP для SwitchModel: убрано дублирование записи конфига в AudioService

Следующие задачи:
1) Свести сборку в 1 файл/конструктор
Заменить связку ServiceFactory + ServiceAssembler + CallbackWirer одной функцией services.BuildContainer(cfg ServiceFactoryConfig) (*ServiceContainer, error) в internal/services/wire.go. Текущие тела можно перенести почти без изменений; LOC уменьшатся, навигация упростится.

2) Упростить интерфейсы сервисов (поэтапно)
На первом шаге: оставить интерфейсы, но сузить их до реально используемых методов. На втором — перенести «минимальные» интерфейсы в пакет потребителя (internal/app) и в ServiceContainer хранить конкретные типы (*AudioService, *UIService, ...). Это лучше соответствует Go (интерфейсы — на границе потребления).

