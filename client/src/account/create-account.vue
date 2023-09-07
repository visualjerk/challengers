<script setup lang="ts">
import { ref, unref } from 'vue'
import { useAccountApi } from './account-api'
import { useRouter } from 'vue-router'

const { create, pending } = useAccountApi()

const router = useRouter()

const name = ref('')
async function createAccount() {
  await create(unref(name))
  router.push('/')
}
</script>

<template>
  <div class="grid gap-2 p-3">
    <h1 class="text-xl">Create Account</h1>
    <form @submit.prevent="createAccount" class="grid gap-2">
      <input
        v-model="name"
        placeholder="How shall we call you?"
        class="p-2 border border-slate-400"
      />
      <button
        type="submit"
        :disabled="pending"
        class="bg-indigo-600 text-white p-2"
      >
        Create Account
      </button>
    </form>
  </div>
</template>
