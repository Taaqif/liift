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
  monthDay: {
    month: "short",
    day: "numeric",
  },
  monthDayYear: {
    month: "short",
    day: "numeric",
    year: "numeric",
  },
};
export default { numberFormats, datetimeFormats };
