// Общее название приложения для Flutter, native-окон и desktop-ярлыка.
const appName = 'Wireglow';

// Метаданные фиксируются при сборке через --dart-define.
const appCommit = String.fromEnvironment('APP_COMMIT', defaultValue: '');
const appBuildTime = String.fromEnvironment('APP_BUILD_TIME', defaultValue: '');
const appTag = String.fromEnvironment('APP_TAG');

const appVersion = String.fromEnvironment(
  'APP_VERSION',
  defaultValue: '0.1.0-dev',
);

// Версии и репозитории ядер берутся из общего манифеста при сборке.
const mihomoVersion = String.fromEnvironment(
  'MIHOMO_VERSION',
  defaultValue: '',
);
const freeTurnVersion = String.fromEnvironment(
  'FREE_TURN_VERSION',
  defaultValue: '',
);
const mihomoRepository = String.fromEnvironment(
  'MIHOMO_REPOSITORY',
  defaultValue: 'MetaCubeX/mihomo',
);
const freeTurnRepository = String.fromEnvironment(
  'FREE_TURN_REPOSITORY',
  defaultValue: 'hackdiaz-dev/free-turn-proxy',
);
