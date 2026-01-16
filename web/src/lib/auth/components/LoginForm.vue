<script setup lang="ts">
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import * as z from "zod";
import { useAuth } from "@/lib/auth/composables/useAuth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ref } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const { login, loading } = useAuth();

const formSchema = toTypedSchema(
  z.object({
    username: z.string().min(1, t("auth.validation.usernameRequired")),
    password: z.string().min(1, t("auth.validation.passwordRequired")),
  }),
);

const form = useForm({
  validationSchema: formSchema,
});
const error = ref("");

const onSubmit = form.handleSubmit(async (values) => {
  try {
    await login(values);
  } catch (err) {
    error.value = t("auth.loginFailed");
  }
});
</script>

<template>
  <form @submit="onSubmit" class="space-y-4">
    <FormField v-slot="{ componentField }" name="username">
      <FormItem>
        <FormLabel>{{ $t("auth.username") }}</FormLabel>
        <FormControl>
          <Input type="text" v-bind="componentField" />
        </FormControl>
        <FormMessage />
      </FormItem>
    </FormField>
    <FormField v-slot="{ componentField }" name="password">
      <FormItem>
        <FormLabel>{{ $t("auth.password") }}</FormLabel>
        <FormControl>
          <Input type="password" v-bind="componentField" />
        </FormControl>
        <FormMessage />
      </FormItem>
    </FormField>
    <div v-if="error" class="text-sm text-destructive">
      {{ error }}
    </div>
    <div class="flex flex-col space-y-4">
      <Button type="submit" class="w-full" :disabled="loading">
        {{ loading ? $t("auth.loggingIn") : $t("auth.login") }}
      </Button>
    </div>
  </form>
</template>
