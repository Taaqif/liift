<script setup lang="ts">
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import * as z from "zod";
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
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { useAuth } from "@/composables/useAuth";
import { useI18n } from "vue-i18n";
import { ref } from "vue";

const { t } = useI18n();

const formSchema = toTypedSchema(
  z.object({
    username: z
      .string()
      .min(3, t("auth.validation.usernameMin"))
      .max(100, t("auth.validation.usernameMax")),
    password: z.string().min(8, t("auth.validation.passwordMin")),
    email: z
      .union([
        z.string().email(t("auth.validation.emailInvalid")),
        z.literal(""),
      ])
      .optional()
      .transform((val) => (val === "" ? undefined : val)),
  }),
);

const { register, loading } = useAuth();

const form = useForm({
  validationSchema: formSchema,
});
const error = ref("");
const onSubmit = form.handleSubmit(async (values) => {
  try {
    await register(values);
  } catch (err) {
    error.value = t("auth.registrationFailed");
  }
});
</script>

<template>
  <div class="flex items-center justify-center min-h-[calc(100vh-200px)]">
    <Card class="w-full max-w-md" as-child>
      <form @submit="onSubmit">
        <CardHeader>
          <CardTitle>{{ $t("auth.register") }}</CardTitle>
          <CardDescription>
            {{ $t("auth.createAccount") }}
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
          <FormField v-slot="{ componentField }" name="email">
            <FormItem>
              <FormLabel>{{ $t("auth.emailOptional") }}</FormLabel>
              <FormControl>
                <Input type="email" v-bind="componentField" />
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
            {{ loading ? $t("auth.registering") : $t("auth.register") }}
          </Button>
          <div class="text-sm text-center text-muted-foreground">
            {{ $t("auth.alreadyHaveAccount") }}
            <Button variant="link" as="a" class="p-0">
              <router-link to="/login"> {{ $t("auth.login") }} </router-link>
            </Button>
          </div>
        </CardFooter>
      </form>
    </Card>
  </div>
</template>
