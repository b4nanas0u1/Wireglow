// ignore: unused_import
import 'package:intl/intl.dart' as intl;

import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Russian (`ru`).
class AppLocalizationsRu extends AppLocalizations {
  AppLocalizationsRu([String locale = 'ru']) : super(locale);

  @override
  String get notSet => 'Не задано';

  @override
  String get starting => 'Запускаем controller…';

  @override
  String get unknownState => 'Состояние подключения неизвестно';

  @override
  String get disconnecting => 'Отключаемся…';

  @override
  String get active => 'Подключение активно';

  @override
  String get connecting => 'Подключение…';

  @override
  String get failed => 'Не удалось подключиться';

  @override
  String get addToConnect => 'Добавьте ссылку для подключения';

  @override
  String get pressToConnect => 'Нажмите на кнопку для подключения';

  @override
  String get disconnect => 'Отключить';

  @override
  String get cancelConnection => 'Отменить подключение';

  @override
  String get retry => 'Попробовать снова';

  @override
  String get connect => 'Подключить';

  @override
  String get connections => 'Подключения';

  @override
  String get about => 'О клиенте';

  @override
  String get profiles => 'Профили';

  @override
  String get addByLink => 'Добавить по ссылке';

  @override
  String get unnamed => 'Без названия';

  @override
  String get connected => 'Подключено';

  @override
  String get savedProfile => 'Сохранённый профиль';

  @override
  String get deleteProfile => 'Удалить профиль';

  @override
  String get checkingLatency => 'Проверяем задержку…';

  @override
  String get checkFailed => 'Проверка не прошла';

  @override
  String get latencyNotMeasured => 'Задержка ещё не измерена';

  @override
  String get latencyHelp =>
      'Время HTTPS-запроса через подключение. Проверяется каждые 30 секунд.';

  @override
  String get checkLatency => 'Проверить задержку';

  @override
  String get noConnections => 'Нет подключений';

  @override
  String get retryHelp =>
      'Проверьте настройки подключения и доступ к сети, затем повторите попытку.';

  @override
  String get errorDetails => 'Подробности ошибки';

  @override
  String get noErrorDetails => 'Controller не сообщил подробностей.';

  @override
  String get properties => 'Свойства подключения';

  @override
  String get systemProxy =>
      'Системный прокси: подключение используют приложения с поддержкой системных настроек прокси.';

  @override
  String get localProxyHelp =>
      'Локальный прокси: задайте показанный адрес в приложении. Системные настройки не изменяются.';

  @override
  String get provider => 'Провайдер';

  @override
  String get turnTransport => 'Транспорт TURN';

  @override
  String get notApplicable => 'Не применяется';

  @override
  String get relayMode => 'Режим relay';

  @override
  String get streams => 'Потоки';

  @override
  String get streamsPerCredential => 'Потоков на кред.';

  @override
  String get obfuscation => 'Обфускация';

  @override
  String get confirmDelete => 'Удалить профиль?';

  @override
  String get cancel => 'Отмена';

  @override
  String get delete => 'Удалить';

  @override
  String get insertLink => 'Вставьте ссылку подключения.';

  @override
  String get addProfile => 'Добавить профиль';

  @override
  String get closeAdd => 'Закрыть добавление';

  @override
  String get pasteHelp => 'Вставьте ссылку, которую вам прислали.';

  @override
  String get connectionLink => 'Ссылка подключения';

  @override
  String get paste => 'Вставить';

  @override
  String get nameFromLink => 'Название берётся из ссылки.';

  @override
  String get add => 'Добавить';

  @override
  String get browserFailed => 'Не удалось открыть ссылку в браузере';

  @override
  String get openGithub => 'Открыть на GitHub';

  @override
  String get repositoryMissing => 'Репозиторий не указан при сборке';

  @override
  String get clientVersion => 'Версия';

  @override
  String get commit => 'Коммит';

  @override
  String get tag => 'Тег';

  @override
  String get buildTime => 'Время сборки';

  @override
  String get engines => 'Ядра';

  @override
  String get unknown => 'неизвестен';

  @override
  String get language => 'Язык';

  @override
  String get systemLanguage => 'Системный';

  @override
  String get lostController =>
      'Связь с controller потеряна. Состояние подключения неизвестно.';

  @override
  String get waitController => 'Сначала отключитесь или дождитесь controller.';

  @override
  String get startFailed => 'Не удалось запустить controller.';

  @override
  String get incompleteBundle =>
      'Не найдены файлы приложения. Повторно распакуйте папку приложения целиком.';

  @override
  String get controllerFailed =>
      'Controller не запустился. Проверьте доступ к папке данных и отсутствие другого запущенного клиента.';

  @override
  String get settingsFailed => 'Не удалось сохранить выбор языка.';

  @override
  String latency(String delay) {
    return 'Задержка: $delay мс';
  }

  @override
  String lastSuccess(String time) {
    return 'Последний успех: $time';
  }

  @override
  String localProxy(String address) {
    return 'Локальный HTTP/SOCKS proxy: $address';
  }

  @override
  String deleteDescription(String name) {
    return 'Профиль «$name» будет удалён. Чтобы добавить его снова, понадобится ссылка подключения.';
  }

  @override
  String get proxyMode => 'Режим прокси';

  @override
  String get localProxyMode => 'Локальный прокси';

  @override
  String get systemProxyMode => 'Системный прокси';

  @override
  String get copyProxyAddress => 'Скопировать адрес прокси';

  @override
  String get switchingProxyMode => 'Меняем режим прокси…';

  @override
  String get proxyModeUnknown => 'Режим прокси неизвестен';

  @override
  String get clientSettings => 'Настройки';

  @override
  String get localAddresses => 'Локальные адреса';

  @override
  String get proxyAddressLabel => 'HTTP/SOCKS-прокси';

  @override
  String get relayAddressLabel => 'Relay FreeTurn';

  @override
  String get addressesHelp =>
      'Формат адреса: 127.0.0.1:порт. Все порты — от 1 до 65535 и должны отличаться. Изменения действуют для всех профилей при следующем подключении.';

  @override
  String get activeSettingsHelp =>
      'Активное подключение использует прежние адреса до переподключения.';

  @override
  String get invalidLocalAddress => 'Введите 127.0.0.1 и порт от 1 до 65535.';

  @override
  String get collidingPorts => 'Порты должны отличаться.';

  @override
  String get saveSettings => 'Сохранить';

  @override
  String get settingsSaved => 'Настройки сохранены';

  @override
  String get resetSettings => 'Вернуть значения по умолчанию';

  @override
  String get confirmResetSettings => 'Сбросить настройки клиента?';

  @override
  String get resetSettingsDescription =>
      'Вернуть адреса по умолчанию, локальный режим прокси и системный язык. Сохранённые профили останутся. Адреса активного подключения изменятся после переподключения.';

  @override
  String get reset => 'Сбросить';

  @override
  String get settingsReset => 'Настройки по умолчанию восстановлены';

  @override
  String get generalSettings => 'Общее';

  @override
  String get networkSettings => 'Сеть';

  @override
  String get relaySettings => 'Релей';

  @override
  String get apiPortLabel => 'Порт API mihomo';

  @override
  String get invalidAPIPort => 'Введите порт от 1 до 65535.';

  @override
  String get profileActions => 'Действия с профилем';

  @override
  String get editProfile => 'Редактировать';

  @override
  String get colorTheme => 'Цветовая тема';

  @override
  String get systemTheme => 'Системная';

  @override
  String get darkTheme => 'Тёмная';

  @override
  String get lightTheme => 'Светлая';

  @override
  String get tunMode => 'TUN';

  @override
  String get tunHelpWindows =>
      'Экспериментальный Windows TUN требует portable-сборку и повышение прав через UAC. Он переподключает сессию и использует прямой HTTPS bootstrap DNS. Другой активный VPN блокирует включение. Для TURN выберите транспорт TCP; UDP-трафик приложений поддерживается.';

  @override
  String get tunHelp =>
      'TUN перехватывает IP-трафик. Включение переподключает сессию и использует 1.1.1.1 как прямой bootstrap DNS. Другой активный VPN блокирует включение.';

  @override
  String get editorTitle => 'Редактор профиля';

  @override
  String get editorProfile => 'Профиль';

  @override
  String get editorName => 'Имя профиля';

  @override
  String get editorNameRequired => 'Укажите имя профиля.';

  @override
  String get editorVK => 'VK Звонки';

  @override
  String get editorCallLinks => 'Ссылки звонков';

  @override
  String editorCallLink(int number) {
    return 'Ссылка звонка $number';
  }

  @override
  String get editorRemoveLink => 'Удалить ссылку звонка';

  @override
  String get editorAddLink => 'Добавить ссылку';

  @override
  String get editorPerformance => 'Производительность';

  @override
  String get editorStreams => 'Потоки на звонок';

  @override
  String get editorCredentials => 'Потоки на набор TURN-доступа';

  @override
  String get editorStreamsHelp => 'Параллельные TURN-сессии для каждой ссылки.';

  @override
  String get editorCredentialsHelp => 'Сессии с общими учётными данными от VK.';

  @override
  String get editorAdvanced => 'Дополнительные настройки';

  @override
  String get editorVKAuth => 'Авторизация VK';

  @override
  String get editorPlatform => 'Тип устройства';

  @override
  String get editorCaptcha => 'Ручная CAPTCHA';

  @override
  String get editorCaptchaHelp =>
      'Решать в браузере вместо автоматического решения.';

  @override
  String get editorTransport => 'Транспорт';

  @override
  String get editorUDPHelp => 'UDP-транспорт недоступен в Windows TUN.';

  @override
  String get editorTurnAddress => 'Адрес узла';

  @override
  String get editorTurnHint => 'Хост или хост:порт';

  @override
  String get editorTurnHelp =>
      'Оставьте пустым для выбора узла через VK. Без порта используется порт от VK. IPv6 с портом: [адрес]:порт.';

  @override
  String get editorDNSMode => 'Режим DNS';

  @override
  String get editorDNSAuto => 'Автоматически';

  @override
  String get editorDNSPlain => 'Обычный DNS';

  @override
  String get editorDNSServers => 'Свои DNS-серверы';

  @override
  String get editorDNSHelp =>
      'DNS для соединений FreeTurn. Адреса IP[:порт], через запятую.';

  @override
  String get editorDoHHelp =>
      'DoH использует встроенные серверы ядра. Свой URL DoH не поддерживается.';

  @override
  String get editorTunDNSHelp =>
      'В TUN DNS выбирает приложение. DNS профиля применяется в локальном и системном прокси.';

  @override
  String get editorObfuscation => 'Обфускация';

  @override
  String get editorObfuscationOff => 'Выключена в этом профиле';

  @override
  String get editorPacing => 'Пейсинг, мс';

  @override
  String get editorPacingHelp =>
      'Интервал между отправками пакетов. 0 — без задержки.';

  @override
  String get editorPacingDisabled =>
      'Доступен при включённой обфускации в импортированном профиле.';

  @override
  String get editorReconnect =>
      'Изменения применятся при следующем подключении.';

  @override
  String editorDefault(String value) {
    return 'По умолчанию ($value)';
  }

  @override
  String get editorInvalidStreams =>
      'Введите 1–1024 или оставьте пустым для значения по умолчанию.';

  @override
  String get editorInvalidLinks =>
      'Укажите хотя бы одну ссылку звонка без запятых.';

  @override
  String get editorInvalidPacing => 'Введите 0–60000 миллисекунд.';

  @override
  String get importProfiles => 'Импорт профилей';

  @override
  String get exportProfiles => 'Экспорт профилей';

  @override
  String get importProfilesAction => 'Импорт…';

  @override
  String get exportProfilesAction => 'Экспорт…';

  @override
  String get profileTransferHelp =>
      'Переносите профили подключений между устройствами или сохраняйте резервную копию.';

  @override
  String get importProfilesSettingsHelp =>
      'Просмотрите файл Wireglow и выберите профили для импорта.';

  @override
  String get exportProfilesSettingsHelp =>
      'Сохраните все профили или выберите отдельные.';

  @override
  String get importProfilesHelp =>
      'Восстановите профили из файла Wireglow. Перед импортом проверьте его содержимое.';

  @override
  String get exportProfilesHelp => 'Выберите профили для сохранения в файл.';

  @override
  String get profileFileType => 'Файл профилей Wireglow (.json)';

  @override
  String get chooseProfileFile => 'Выбрать файл…';

  @override
  String get changeProfileFile => 'Изменить';

  @override
  String get selectAllProfiles => 'Выбрать все';

  @override
  String get profileExportSecretsHelp =>
      'Включает настройки подключений, ключи и ссылки на звонки. Любой, у кого есть этот файл, сможет использовать эти подключения.';

  @override
  String get profileImportReviewHelp =>
      'Существующие профили определяются по ID. Замена профиля применяется при следующем подключении.';

  @override
  String get skipExistingProfile => 'Пропустить';

  @override
  String get copyImportedProfile => 'Создать копию';

  @override
  String get replaceExistingProfile => 'Заменить';

  @override
  String get profileArchiveInvalid =>
      'Некорректный файл Wireglow, неподдерживаемая версия или превышен лимит: 16 МиБ и 1000 профилей.';

  @override
  String get profilePreviewExpired =>
      'Предпросмотр устарел или выполняется операция подключения. Откройте файл и проверьте профили ещё раз.';

  @override
  String get profileTransferFailed =>
      'Не удалось завершить операцию. Проверьте права доступа и свободное место, затем выберите файл ещё раз.';

  @override
  String get profileImportPartial =>
      'Часть профилей не импортирована. Результат для каждого профиля указан ниже. Для повтора выберите файл ещё раз.';

  @override
  String get profileImportDone => 'Импортирован';

  @override
  String get profileImportSkipped => 'Пропущен';

  @override
  String get profileImportChanged =>
      'Изменён после предпросмотра; не импортирован';

  @override
  String get profileImportFailed => 'Не удалось сохранить профиль';

  @override
  String get closeProfileTransfer => 'Закрыть';

  @override
  String profileCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count профиля',
      many: '$count профилей',
      few: '$count профиля',
      one: '$count профиль',
      zero: 'Профили не выбраны',
    );
    return '$_temp0';
  }

  @override
  String exportProfileCount(int count) {
    return 'Экспортировать: $count';
  }

  @override
  String importProfileCount(int count) {
    return 'Импортировать: $count';
  }

  @override
  String profilesExported(int count) {
    return 'Профилей экспортировано: $count';
  }

  @override
  String profilesImported(int count) {
    return 'Профилей импортировано: $count';
  }

  @override
  String profileAlreadyExists(String name) {
    return 'Уже сохранён как «$name»';
  }
}
