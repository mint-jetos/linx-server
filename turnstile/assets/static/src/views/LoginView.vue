<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import axios from "axios";
import {
  Card,
  CardContent,
  CardHeader,
} from "@/components/ui/card/index.js";
import PasswordInput from "@/components/upload/PasswordInput.vue";
import LockIcon from "~icons/material-symbols/lock";
import { toast } from "vue-sonner";
import { obfuscate, deobfuscate } from "@/util/obfuscate.ts";
import { useConfigStore } from "@/stores/config.ts";


const password = ref("");
const isLoading = ref(false);
const router = useRouter();
const config = useConfigStore();
const turnstileToken = ref("");

const onTurnstileSuccess = (token: string) => {
  turnstileToken.value = token;
};

onMounted(() => {
  if (config.site?.turnstile?.enabled) {
    (window as any).onTurnstileSuccess = onTurnstileSuccess;
  }
});

onUnmounted(() => {
  delete (window as any).onTurnstileSuccess;
});

import CryptoJS from 'crypto-js';

const handleLogin = async () => {
  if (!password.value) return;
  
  isLoading.value = true;
  const hashedPassword = CryptoJS.SHA256(password.value).toString(CryptoJS.enc.Hex);  try {
    const encoded = obfuscate(hashedPassword);
    const response = await axios.get(`/api/login/auth?d=${encodeURIComponent(encoded)}&cf-turnstile-response=${turnstileToken.value}&json=true`);
    if (response.status === 200) {
      const data = JSON.parse(deobfuscate(response.data));
      if (data.message === "Login successful") {
        toast.success("Login successful");
        router.push("/");
      }
    }
  } catch (error: any) {
    console.error(error);
    let errorMsg = "Login failed";
    if (error.response?.data) {
      try {
        const data = JSON.parse(deobfuscate(error.response.data));
        errorMsg = data.error || errorMsg;
      } catch (e) {}
    }
    toast.error(errorMsg);
    
    // Reset Turnstile if failed
    if (config.site?.turnstile?.enabled && (window as any).turnstile) {
      (window as any).turnstile.reset();
      turnstileToken.value = "";
    }
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
          <button
            type="button"
            @click="handleLogin"
            :disabled="isLoading"
            class="p-3 rounded-full bg-primary/10 text-primary hover:bg-primary/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <LockIcon class="w-8 h-8" />
          </button>
        </div>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleLogin" class="space-y-4">
          <div class="flex flex-col space-y-4">
            <PasswordInput
              id="password"
              v-model="password"
              placeholder="Password"
              required
              :disabled="isLoading"
              autofocus
              class="flex-grow"
            />
            <div
              v-if="config.site?.turnstile?.enabled"
              class="cf-turnstile flex justify-center"
              :data-sitekey="config.site.turnstile.site_key"
              data-callback="onTurnstileSuccess"
              data-theme="auto"
            ></div>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
