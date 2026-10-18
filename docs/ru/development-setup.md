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

## Частые команды

Из корня репозитория:

```bash
go test ./...
```

Сборка полного дистрибутива выполняется платформенным сценарием выше. Сценарии сборки сами выполняют `flutter pub get`. Для работы только с Flutter UI перейдите в `ui/` и используйте обычные команды Flutter, например `flutter run -d windows` или `flutter run -d linux`.

Поддерживаемые локальные сборочные цели сейчас — Windows x64 и Linux x64. Сценарии Bash и PowerShell временные и независимые; единая кроссплатформенная система сборки пока не выбрана.
