<script setup lang="ts">
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useRouter } from "vue-router";
import LoginForm from "@/lib/auth/components/LoginForm.vue";

const router = useRouter();
</script>

<template>
  <div class="flex-1 flex items-center justify-center">
    <Card class="w-full max-w-md">
      <CardHeader>
        <CardTitle>{{ $t("auth.login") }}</CardTitle>
        <CardDescription>
          {{ $t("auth.enterCredentials") }}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <LoginForm />
      </CardContent>
      <CardFooter>
        <div class="w-full text-sm text-center text-muted-foreground">
          {{ $t("auth.dontHaveAccount") }}
          <Button variant="link" as="a" class="p-0">
            <router-link to="/register">
              {{ $t("auth.register") }}
            </router-link>
          </Button>
        </div>
      </CardFooter>
    </Card>
  </div>
</template>
