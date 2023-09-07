<script setup lang="ts">
import { useRouter } from 'vue-router'
import { gameClient } from './game-api'

const router = useRouter()
async function createGame() {
  const { response } = await gameClient.createGame({})
  router.push(`/game/${response.id}`)
}
</script>

<template>
  <div class="grid gap-2 p-3">
    <h1 class="text-xl">Games</h1>
    <div>
      <button @click="createGame">Create New Game</button>
    </div>
  </div>
</template>
