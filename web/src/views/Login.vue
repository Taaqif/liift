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
import { useAuth } from "@/composables/useAuth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useRouter } from "vue-router";
import { ref } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const router = useRouter();
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
  <div class="flex items-center justify-center min-h-[calc(100vh-200px)]">
    <Card class="w-full max-w-md" as-child>
      <form @submit="onSubmit">
        <CardHeader>
          <CardTitle>{{ $t("auth.login") }}</CardTitle>
          <CardDescription>
            {{ $t("auth.enterCredentials") }}
          </CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
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
        </CardContent>
        <CardFooter class="flex flex-col space-y-4">
          <Button type="submit" class="w-full" :disabled="loading">
            {{ loading ? $t("auth.loggingIn") : $t("auth.login") }}
          </Button>
          <div class="text-sm text-center text-muted-foreground">
            {{ $t("auth.dontHaveAccount") }}
            <Button variant="link" as="a" class="p-0">
              <router-link to="/register">
                {{ $t("auth.register") }}
              </router-link>
            </Button>
          </div>
        </CardFooter>
      </form>
    </Card>
  </div>
</template>
