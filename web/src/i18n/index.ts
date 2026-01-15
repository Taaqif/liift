import { createI18n } from "vue-i18n";
import enFormats from "./formats/en";
import en from "./locales/en.json" with { type: "json" };
// https://isiungk.medium.com/full-example-for-vue-3-project-using-vue-i18n-v9-with-dynamically-loading-locale-files-and-f22cb02d8693
const i18n = createI18n({
  legacy: false,
  locale: "en",
  fallbackLocale: "en",
  messages: {
    en,
  },
  numberFormats: {
    en: enFormats.numberFormats,
  },
  datetimeFormats: {
    en: enFormats.datetimeFormats,
  },
});

export { i18n };
