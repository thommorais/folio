const internationalization = {
  "locales": [
    "en",
    "pt"
  ],
  "requiredLocales": [
    "en",
    "pt"
  ],
  "strictMode": "inclusive",
  "defaultLocale": "pt"
};
const routing = {
  "mode": "prefix-no-default",
  "storage": {
    "cookies": [
      {
        "name": "INTLAYER_LOCALE",
        "attributes": {
          "path": "/"
        }
      }
    ],
    "headers": [
      {
        "name": "x-intlayer-locale"
      }
    ]
  },
  "basePath": ""
};
const editor = {
  "editorURL": "http://localhost:8000",
  "cmsURL": "https://app.intlayer.org",
  "backendURL": "https://back.intlayer.org",
  "port": 8000,
  "enabled": false,
  "dictionaryPriorityStrategy": "local_first",
  "liveSync": false,
  "liveSyncPort": 4000,
  "liveSyncURL": "http://localhost:4000"
};
const log = {
  "mode": "default",
  "prefix": "\u001b[38;5;239m[intlayer] \u001b[0m"
};
const system = {
  "baseDir": "/Users/thommorais/shed/journ/folio/apps/web",
  "moduleAugmentationDir": "/Users/thommorais/shed/journ/folio/apps/web/.intlayer/types",
  "unmergedDictionariesDir": "/Users/thommorais/shed/journ/folio/apps/web/.intlayer/unmerged_dictionary",
  "remoteDictionariesDir": "/Users/thommorais/shed/journ/folio/apps/web/.intlayer/remote_dictionary",
  "dictionariesDir": "/Users/thommorais/shed/journ/folio/apps/web/.intlayer/dictionary",
  "dynamicDictionariesDir": "/Users/thommorais/shed/journ/folio/apps/web/.intlayer/dynamic_dictionary",
  "fetchDictionariesDir": "/Users/thommorais/shed/journ/folio/apps/web/.intlayer/fetch_dictionary",
  "typesDir": "/Users/thommorais/shed/journ/folio/apps/web/.intlayer/types",
  "mainDir": "/Users/thommorais/shed/journ/folio/apps/web/.intlayer/main",
  "configDir": "/Users/thommorais/shed/journ/folio/apps/web/.intlayer/config",
  "cacheDir": "/Users/thommorais/shed/journ/folio/apps/web/.intlayer/cache",
  "tempDir": "/Users/thommorais/shed/journ/folio/apps/web/.intlayer/tmp"
};
const content = {
  "fileExtensions": [
    ".content.ts",
    ".content.js",
    ".content.cjs",
    ".content.mjs",
    ".content.json",
    ".content.json5",
    ".content.jsonc",
    ".content.tsx",
    ".content.jsx",
    ".content.md",
    ".content.mdx",
    ".content.yaml",
    ".content.yml"
  ],
  "contentDir": [
    "/Users/thommorais/shed/journ/folio/apps/web"
  ],
  "codeDir": [
    "/Users/thommorais/shed/journ/folio/apps/web"
  ],
  "excludedPath": [
    "**/node_modules/**",
    "**/dist/**",
    "**/build/**",
    "**/.intlayer/**",
    "**/.next/**",
    "**/.nuxt/**",
    "**/.expo/**",
    "**/.vercel/**",
    "**/.turbo/**",
    "**/.tanstack/**",
    "**/.output/**",
    "**/.svelte-kit/**"
  ],
  "watch": true
};
const ai = {};
const dictionary = {
  "fill": true,
  "contentAutoTransformation": false,
  "location": "local",
  "importMode": "static"
};
const build = {
  "mode": "auto",
  "minify": false,
  "purge": false,
  "traversePattern": [
    "**/*.{tsx,ts,js,mjs,cjs,jsx,vue,svelte,astro}",
    "!**/node_modules/**",
    "!**/dist/**",
    "!**/build/**",
    "!**/.intlayer/**",
    "!**/.next/**",
    "!**/.nuxt/**",
    "!**/.expo/**",
    "!**/.vercel/**",
    "!**/.turbo/**",
    "!**/.tanstack/**",
    "!**/.output/**",
    "!**/.svelte-kit/**",
    "!**/*.config.*",
    "!**/*.test.*",
    "!**/*.spec.*",
    "!**/*.stories.*",
    "!**/*.d.ts",
    "!**/*.d.ts.map",
    "!**/*.map"
  ],
  "outputFormat": [
    "esm",
    "cjs"
  ],
  "cache": true,
  "checkTypes": false
};
const compiler = {
  "enabled": true,
  "dictionaryKeyPrefix": "",
  "noMetadata": false,
  "saveComponents": false
};

module.exports.internationalization = internationalization;
module.exports.routing = routing;
module.exports.editor = editor;
module.exports.log = log;
module.exports.system = system;
module.exports.content = content;
module.exports.ai = ai;
module.exports.dictionary = dictionary;
module.exports.build = build;
module.exports.compiler = compiler;
