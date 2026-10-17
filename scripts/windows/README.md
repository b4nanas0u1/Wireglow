# Wireglow — временный Windows-комплект для разработки

Запуск: открыть `Wireglow.exe` в корне комплекта. Папку переносить целиком.
Приложение само находит controller, free-turn-proxy и mihomo в `bin/`.
Корень содержит полный Flutter release bundle; `bin/` — Go controller и ядра.
`SHA256SUMS` содержит SHA-256 всех файлов комплекта, кроме самого списка.
Профили и настройки хранятся отдельно, в пользовательских каталогах.

Это эквивалент текущей Linux-папки сборки, не готовый Windows ZIP-релиз.
Visual C++ runtime включён рядом с Wireglow.exe: msvcp140.dll, vcruntime140.dll
и vcruntime140_1.dll (x64). Отдельная установка runtime, Go или Flutter
пользователю не требуется. Целевая платформа — Windows 11 x64.
Приёмка на чистой Windows остаётся отдельной проверкой.

Сборка: `./scripts/build-windows.ps1` из корня исходного репозитория. Go, Flutter, Visual
Studio Build Tools с C++ и Windows SDK нужны на машине сборки.
DLL берутся из Redistributable той установки Visual Studio, которую выбрал
Flutter/CMake. При отсутствии полного набора сборка не публикует новую папку.
В дальнейшем временные Windows/ Linux-скрипты должны быть заменены общей
кроссплатформенной системой сборки.
