/**
 * global type definitions
 */

import "vue-i18n";
import en from "./i18n/locales/en.json";

type MessageSchema = typeof en;

declare module "vue-i18n" {
  // define the locale messages schema
  export interface DefineLocaleMessage extends MessageSchema {}
}
