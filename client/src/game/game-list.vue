<script setup lang="ts">
import { useRouter } from 'vue-router'
import { gameClient, GameEntry } from './game-api'
import { ref } from 'vue'

const router = useRouter()
async function createGame() {
  const { response } = await gameClient.createGame({})
  router.push(`/game/${response.id}`)
}

const games = ref<GameEntry[]>([])
async function loadGames() {
  const { response } = await gameClient.list({})
  games.value = response.games
}
loadGames()
</script>

<template>
  <div class="grid gap-2 p-3">
    <h1 class="text-xl">Games</h1>
    <div>
      <button @click="createGame">Create New Game</button>
    </div>
    <ul>
      <li v-for="game in games" :key="game.id">
        <RouterLink :to="`/game/${game.id}`">
          Open game with {{ game.state?.players.length }} players
        </RouterLink>
      </li>
    </ul>
  </div>
</template>
