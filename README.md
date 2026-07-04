# izu-cli

CLI утилита для просмотра аниме с TUI интерфейсом.

## Возможности

- Поиск аниме с динамическим fuzzy matching
- 3 источника: AnimeKai, AllAnime, AnimePahe
- Воспроизведение через mpv
- Автоматическая загрузка субтитров
- История просмотра и избранное (SQLite)
- Discord Rich Presence
- Скачивание эпизодов

## Установка

```bash
git clone https://github.com/izu/izu-cli.git
cd izu-cli
go build -o izu ./cmd/izu/
sudo mv izu /usr/local/bin/
```

## Зависимости

- Go 1.22+
- mpv (для воспроизведения)
- ffmpeg (опционально, для скачивания)

## Конфигурация

Конфигурационный файл: `~/.config/izu-cli/config.yaml`

```yaml
general:
  provider: "animekai"  # animekai, allanime, animepahe
  theme: "dark"

player:
  volume: 100

providers:
  animekai:
    enabled: true
    base_url: "https://animekai.to"
```

## Использование

```bash
# Запуск
izu

# Горячие клавиши
/       - Поиск
Tab     - Переключить провайдер
Enter   - Выбрать
f       - Добавить в избранное
h       - История просмотра
q       - Выход
```

## Сборка

```bash
make build    # Собрать бинарник
make install  # Установить в /usr/local/bin
clean         # Очистить
```
