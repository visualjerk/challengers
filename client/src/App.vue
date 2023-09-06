<script setup lang="ts">
import { createAccount } from './api/game'
import { ref } from 'vue'

const pending = ref(true)
async function createGame() {
  pending.value = true
  await createAccount()
  pending.value = false
}
createGame()
</script>

<template>
  <div v-if="pending">Creating account ...</div>
  <RouterView v-else />
</template>
