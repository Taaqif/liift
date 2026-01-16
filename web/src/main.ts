import { createApp } from "vue";
import "./style.css";
import App from "./App.vue";
import router from "./router";
import { i18n } from "./i18n";
import { VueQueryPlugin } from "@tanstack/vue-query";
import { queryClient } from "./lib/queryClient";

const app = createApp(App);
app.use(i18n);
app.use(router);
app.use(VueQueryPlugin, {
  queryClient,
});

app.mount(document.body);
