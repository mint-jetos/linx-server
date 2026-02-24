<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import axios from "axios";
import {
  Card,
  CardContent,
  CardHeader,
} from "@/components/ui/card/index.js";
import { Input } from "@/components/ui/input/index.js";
import LockIcon from "~icons/material-symbols/lock";
import { toast } from "vue-sonner";


const password = ref("");
const isLoading = ref(false);
const router = useRouter();

const handleLogin = async () => {
  if (!password.value) return;

  isLoading.value = true;
  try {
    const response = await axios.post("/api/admin/auth", { password: password.value });
    if (response.status === 200) {
      toast.success("Login successful");
      router.push("/");
    }
  } catch (error: any) {
    console.error(error);
    toast.error(error.response?.data?.error || "Login failed");
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div class="flex items-center justify-center min-h-[50vh]">
    <Card class="w-full max-w-md">
      <CardHeader class="space-y-1">
        <div class="flex justify-center mb-4">
          <div class="p-3 rounded-full bg-primary/10 text-primary">
            <LockIcon class="w-8 h-8" />
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleLogin" class="space-y-4">
          <div class="flex items-center space-x-2">
            <Input
              id="password"
              v-model="password"
              type="password"
              placeholder="••••••••"
              required
              :disabled="isLoading"
              autofocus
              class="flex-grow"
              @keyup.enter="handleLogin"
            />
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
