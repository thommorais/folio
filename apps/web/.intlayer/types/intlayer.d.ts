import "intlayer";

declare module 'intlayer' {
  interface __DictionaryRegistry {

  }

  interface __DeclaredLocalesRegistry {
    "en": 1;
    "pt": 1;
  }

  interface __RequiredLocalesRegistry {
    "en": 1;
    "pt": 1;
  }

  interface __SchemaRegistry {

  }

  interface __StrictModeRegistry { mode: 'inclusive' }

  interface __EditorRegistry { enabled : false }

  interface __RoutingRegistry { mode: 'prefix-no-default'; defaultLocale: 'pt' }
}
