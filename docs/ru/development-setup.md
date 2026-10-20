# Настройка среды разработки

Здесь собраны шаги для сборки Wireglow из исходников на Windows и Linux x64. Команды рассчитаны на свежую копию репозитория. Разработка и сборка выполняются из корня Git-репозитория.

## Что установить

| Инструмент | Версия проекта | Ссылка |
|---|---:|---|
| Go | 1.27.1 или новее в пределах совместимой версии из `go.mod` | [go.dev/dl](https://go.dev/dl/) |
| Flutter (включает Dart) | 3.47.7, stable | [Установка Flutter](https://docs.flutter.dev/install) |
| Git | актуальная стабильная | [git-scm.com](https://git-scm.com/downloads) |

Flutter 3.47.7 — версия, закреплённая CI. Go-модули задают минимальную версию 1.27.1. Для воспроизводимой разработки используйте эти версии.

## Windows 11 x64

### 1. Установите инструменты

Установите [Git for Windows](https://git-scm.com/download/win), [Go](https://go.dev/dl/) и [Flutter SDK](https://docs.flutter.dev/install/manual). При ручной установке Flutter распакуйте SDK в каталог без пробелов, например `C:\src\flutter`, и добавьте `C:\src\flutter\bin` в `PATH`.

Для Flutter Windows runner установите [Visual Studio 2022 Build Tools](https://visualstudio.microsoft.com/downloads/) с workload **Desktop development with C++**. Убедитесь, что установлены MSVC C++ build tools, CMake tools for Windows и Windows 10/11 SDK. Это Visual Studio Build Tools, не только Visual Studio Code.

### 2. Проверьте установку

Откройте новое окно PowerShell:

```powershell
git --version
go version
flutter --version
flutter doctor -v
```

В `flutter doctor -v` проверьте, что Flutter видит Visual Studio и Windows SDK. Предупреждения о недостающих Android или веб-инструментах не относятся к сборке Wireglow для Windows.

### 3. Получите исходники и соберите приложение

```powershell
git clone <URL репозитория>
cd <каталог репозитория>
./scripts/build-windows.ps1
```

Подставьте URL репозитория с его публичной страницы. Если репозиторий уже открыт, перейдите в его корень и выполните последнюю команду. Сценарий загрузит закреплённые сетевые ядра, соберёт Go controller и Flutter UI, затем подготовит полный комплект в `dist/windows-x64/`. Запуск: `dist\windows-x64\Wireglow.exe`.

Flutter из `PATH` используется автоматически. Можно также положить SDK в `.cache\flutter-windows\flutter` в корне репозитория — сборочный сценарий обнаружит его там. Кэши сборки хранятся только в `.cache/` внутри репозитория.

### Portable-комплект Windows

Запустите `./scripts/build-windows-portable.ps1`
(или `./scripts/build-windows.ps1 -Portable`). Отдельный результат —
`dist/windows-x64-portable/`; запускайте `Wireglow.exe` напрямую.
Маркер `portable.json` сохраняет профили, настройки UI/backend и журнал
восстановления прокси в `data/` внутри комплекта, а сокеты, блокировки и состояние
launcher — в `data/runtime/`. Portable GUI игнорирует `WIREGLOW_DATA_DIR` и
`WIREGLOW_RUNTIME_DIR`. Папка должна быть доступна для записи; при ошибке перехода
на AppData нет. DLL Visual C++ runtime включены, как в обычном комплекте.

Существующий portable-каталог не заменяется: перед повторной сборкой переместите
его вместе с данными. Перед переносом или обновлением закройте приложение и
дождитесь recovery. Сохраняйте пользовательские файлы в `data/`, заменяя только
ресурсы Flutter (`app.so`, `icudtl.dat`, `flutter_assets/`) и другие файлы поставки.
Не применяйте незавершённый журнал восстановления прокси на другом компьютере.
Распаковывайте по короткому пути: Unix-сокеты Windows ограничивают его длину.
Проверяйте комплектный `SHA256SUMS` до первого запуска.

CI собирает оба Windows-варианта и публикует
`Wireglow-<version>-windows-x64.zip` и `Wireglow-<version>-windows-x64-portable.zip`
как артефакты и файлы релиза. Файлы лежат непосредственно в корне ZIP.
Релизный `SHA256SUMS` включает оба архива.

## Linux x64 (Ubuntu/Debian)

### 1. Установите системные зависимости

Для Flutter Linux runner нужны GTK 3, Clang, CMake, Ninja, pkg-config и liblzma. Команды ниже проверены для Ubuntu/Debian:

```bash
sudo apt-get update
sudo apt-get install -y git curl xz-utils clang cmake ninja-build pkg-config libgtk-3-dev liblzma-dev
```

### 2. Установите Go и Flutter

Установите Go версии 1.27.1 или новее по [официальной инструкции для Linux](https://go.dev/doc/install). Не полагайтесь на версию из стандартного репозитория дистрибутива: она может быть ниже требуемой. После установки проверьте `go version`.

Установите Flutter 3.47.7 stable по [официальной инструкции](https://docs.flutter.dev/install/manual) или выберите эту версию через [Flutter SDK archive](https://docs.flutter.dev/install/archive). Добавьте каталог `flutter/bin` в `PATH`, например для локальной установки:

```bash
export PATH="$HOME/development/flutter/bin:$PATH"
```

Добавьте эту строку в `~/.bashrc` или конфигурацию своей оболочки, если хотите сохранить настройку. Установите Git, если его ещё нет (команда установки выше включает его).

### 3. Проверьте установку

Откройте новый терминал:

```bash
git --version
go version
flutter --version
flutter doctor -v
```

Проверьте, что `flutter doctor -v` распознаёт Linux toolchain. Android и веб toolchains для сборки Wireglow не нужны.

### 4. Получите исходники и соберите приложение

```bash
git clone <repository URL>
cd <repository directory>
./scripts/build-linux.sh
```

Замените адрес на URL репозитория с его публичной страницы. Если репозиторий уже открыт, перейдите в его корень и запустите `./scripts/build-linux.sh`. Полный комплект появится в `dist/linux-x64/`; запустите `./dist/linux-x64/Wireglow`.

Для portable-комплекта выполните `./scripts/build-linux-portable.sh`
(эквивалентно `./scripts/build-linux.sh --portable`). Результат —
`dist/linux-x64-portable/`.
Маркер `portable.json` рядом с `Wireglow` направляет профили, настройки UI/backend
и журнал восстановления в `data/`, временное состояние — в `data/runtime/`.
Без маркера используются прежние системные каталоги. Portable GUI игнорирует
`WIREGLOW_DATA_DIR` и `WIREGLOW_RUNTIME_DIR`; явные CLI-флаги controller сохраняют приоритет.
Повторная сборка не заменяет существующий portable-каталог: сначала переместите
его вместе с пользовательскими данными. При обновлении сохраняйте пользовательские
файлы в `data/`; ресурсы Flutter в этой же папке обновляются из нового комплекта.
Подробности переноса и ограничений приведены в README собранного комплекта.


Сценарий скачивает закреплённые сетевые ядра и собирает controller и UI. Кэши Go, Dart, генераторов и ядер находятся в `.cache/` в корне репозитория. Flutter SDK берётся из `.cache/flutter/` или `PATH`. Соседние папки проекта не требуются.


При публикации тега GitHub Actions отдельно собирает оба Linux-варианта и
прикладывает к релизу `Wireglow-<версия>-linux-x64.zip` и
`Wireglow-<версия>-linux-x64-portable.zip`. В корне каждого ZIP находятся файлы
и папки клиента, без внешней папки дистрибутива. Релизный `SHA256SUMS` содержит
контрольные суммы архивов. Локальные сборочные скрипты создают только папки.

## Проверки и CI

GitHub Actions запускает проверки и обычные/portable-сборки Linux и Windows
при каждом обновлении PR, каждом коммите в `master` и публикации релиза.
Ручная сборка через `workflow_dispatch` по-прежнему доступна для `master`.
Этапы выполняются последовательно: линты → тесты → сборки → публикация релиза.
Обычные и интеграционные тесты запускаются параллельно после успешных линтов
в обеих lint-jobs; сборки ждут завершения всех тестов. Ошибка блокирует следующие этапы.
Go vet, Flutter analyze, Go-тесты, Flutter-тесты, Go-интеграции и Dart↔Go IPC
отображаются отдельными jobs. Обе lint-jobs запускаются только на Linux;
Go vet проверяет Linux- и Windows-код через `GOOS`. Тесты Linux и Windows разделены.
Архивы сохраняются как артефакты каждого запуска. Прикрепление архивов и
`SHA256SUMS` к GitHub Release выполняется только по событию `release: published`
после успешных проверок и сборок.

Команды линтов (Linux), затем тестов (Linux и Windows):

```sh
GOOS=linux go vet ./...
GOOS=windows go vet ./...
cd ui
flutter pub get
flutter analyze --no-pub
```

```sh
go test -count=1 -timeout 120s ./...
cd ui
flutter pub get
flutter test --no-pub --reporter expanded
```

Отдельные Linux-задачи проверяют Go-интеграции (systemd lifecycle, восстановление
прокси, GNOME settings в изолированной D-Bus-сессии, настоящее ядро Mihomo)
и Dart↔Go IPC.
Для локального запуска нужны работающий user systemd, `dbus-run-session`,
`dconf-cli`, `gsettings-desktop-schemas`, Go и Flutter. Из корня репозитория:

```bash
source scripts/build-env.sh
./scripts/download-linux-engines.sh
WIREGLOW_SYSTEMD_TEST=1 WIREGLOW_RECOVERY_TEST=1 WIREGLOW_GSETTINGS_TEST=1 \
  WIREGLOW_MIHOMO_TEST_BINARY="$cache_root/engines/linux-amd64/bin/mihomo" \
  go test -count=1 -timeout 180s -v ./cmd/controller ./internal/infrastructure/platform ./internal/infrastructure/engines
mkdir -p "$cache_root/bin"
go build -o "$cache_root/bin/controller" ./cmd/controller
(cd ui && flutter pub get)
./ui/tool/smoke.sh
```

Без этих переменных обычный Go-прогон пропускает дополнительные Linux/Mihomo
интеграции. Smoke использует фиктивные ядра; реальное VPN-подключение он не проверяет.

## Частые команды

Из корня репозитория:

```bash
go test ./...
```

Сборка полного дистрибутива выполняется платформенным сценарием выше. Сценарии сборки сами выполняют `flutter pub get`. Для работы только с Flutter UI перейдите в `ui/` и используйте обычные команды Flutter, например `flutter run -d windows` или `flutter run -d linux`.

Поддерживаемые локальные сборочные цели сейчас — Windows x64 и Linux x64. Сценарии Bash и PowerShell временные и независимые; единая кроссплатформенная система сборки пока не выбрана.
