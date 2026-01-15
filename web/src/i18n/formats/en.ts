import type { DateTimeFormats, NumberFormats } from "./types";

const numberFormats: NumberFormats = {
  currency: {
    style: "currency",
    currency: "USD",
  },
  decimal: {
    style: "decimal",
    minimumFractionDigits: 2,
  },
};
const datetimeFormats: DateTimeFormats = {
  short: {
    year: "numeric",
    month: "numeric",
    day: "numeric",
  },
  long: {
    year: "numeric",
    month: "long",
    day: "numeric",
    weekday: "long",
  },
};
export default { numberFormats, datetimeFormats };
