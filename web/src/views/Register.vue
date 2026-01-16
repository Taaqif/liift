<script setup lang="ts">
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAuth } from "@/lib/auth/composables/useAuth";
import { useI18n } from "vue-i18n";
import { ref } from "vue";
import RegisterForm from "@/lib/auth/components/RegisterForm.vue";

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
        <CardContent>
          <RegisterForm />
        </CardContent>
        <CardFooter>
          <div class="w-full text-sm text-center text-muted-foreground">
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
